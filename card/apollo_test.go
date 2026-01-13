// Package card provides functionalities related to card operations.
package card

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	testhelpers "github.com/ubavic/bas-celik/v2/test_helpers"
)

func Test_InitCard(t *testing.T) {
	card := &Apollo{}
	err := card.InitCard()
	if err != nil {
		t.Fatalf("InitCard() returned an error: %v", err)
	}
}

func withStatus(data []byte) []byte {
	out := append([]byte{}, data...)
	out = append(out, 0x90, 0x00)
	return out
}

func header(length uint16) []byte {
	h := make([]byte, 6)
	binary.LittleEndian.PutUint16(h[4:], length)
	return h
}

func TestApolloReadCard(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(cm *testhelpers.CardMock)
		wantErr    bool
		errSubstr  string
		assertData func(t *testing.T, c *Apollo)
	}{
		{
			name: "success",
			setup: func(cm *testhelpers.CardMock) {
				for _, rsp := range [][]byte{
					withStatus(nil), withStatus(header(1)), withStatus([]byte{0x10}), // document
					withStatus(nil), withStatus(header(1)), withStatus([]byte{0x20}), // personal
					withStatus(nil), withStatus(header(1)), withStatus([]byte{0x30}), // residence
					withStatus(nil), withStatus(header(5)), withStatus([]byte{0xFF, 0xEE, 0xDD, 0xCC, 0x40}), // photo
				} {
					cm.On("Transmit", mock.Anything).Return(rsp, nil).Once()
				}
			},
			assertData: func(t *testing.T, c *Apollo) {
				if !bytes.Equal(c.documentFile, []byte{0x10}) {
					t.Fatalf("documentFile = %x", c.documentFile)
				}
				if !bytes.Equal(c.personalFile, []byte{0x20}) {
					t.Fatalf("personalFile = %x", c.personalFile)
				}
				if !bytes.Equal(c.residenceFile, []byte{0x30}) {
					t.Fatalf("residenceFile = %x", c.residenceFile)
				}
				if len(c.photoFile) == 0 {
					t.Fatalf("photoFile should be set")
				}
			},
		},
		{
			name: "document select error",
			setup: func(cm *testhelpers.CardMock) {
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail")).Once()
			},
			wantErr:   true,
			errSubstr: "document file",
		},
		{
			name: "document bad status",
			setup: func(cm *testhelpers.CardMock) {
				cm.On("Transmit", mock.Anything).Return([]byte{0x6A, 0x82}, nil).Once()
			},
			wantErr:   true,
			errSubstr: "document file",
		},
		{
			name: "personal short header",
			setup: func(cm *testhelpers.CardMock) {
				for _, rsp := range [][]byte{
					withStatus(nil), withStatus(header(1)), withStatus([]byte{0x10}), // document ok
					withStatus(nil), withStatus([]byte{0x01}), // personal header too short
				} {
					cm.On("Transmit", mock.Anything).Return(rsp, nil).Once()
				}
			},
			wantErr:   true,
			errSubstr: "personal file",
		},
		{
			name: "photo read error",
			setup: func(cm *testhelpers.CardMock) {
				for _, call := range []struct {
					rsp []byte
					err error
				}{
					{withStatus(nil), nil}, {withStatus(header(1)), nil}, {withStatus([]byte{0x10}), nil}, // document
					{withStatus(nil), nil}, {withStatus(header(1)), nil}, {withStatus([]byte{0x20}), nil}, // personal
					{withStatus(nil), nil}, {withStatus(header(1)), nil}, {withStatus([]byte{0x30}), nil}, // residence
					{withStatus(nil), nil}, {withStatus(header(5)), nil}, {nil, errors.New("boom")}, // photo read fails
				} {
					cm.On("Transmit", mock.Anything).Return(call.rsp, call.err).Once()
				}
			},
			wantErr:   true,
			errSubstr: "photo file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			if tt.setup != nil {
				tt.setup(cm)
			}
			card := &Apollo{smartCard: cm}

			err := card.ReadCard()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.errSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("ReadCard() error = %v", err)
				}
				if tt.assertData != nil {
					tt.assertData(t, card)
				}
			}
			cm.AssertExpectations(t)
		})
	}
}

func TestApolloSelectFile(t *testing.T) {
	tests := []struct {
		name      string
		response  []byte
		respErr   error
		wantErr   bool
		errSubstr string
	}{
		{"success", withStatus(nil), nil, false, ""},
		{"transmit error", nil, errors.New("fail"), true, "fail"},
		{"bad status", []byte{0x6A, 0x82}, nil, true, "response"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			cm.On("Transmit", mock.Anything).Return(tt.response, tt.respErr).Once()

			card := &Apollo{smartCard: cm}
			_, err := card.selectFile([]byte{0xAA}, 2)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.errSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("selectFile() error = %v", err)
			}
			exp := buildAPDU(0x00, 0xA4, 0x08, 0x00, []byte{0xAA}, 2)
			if len(cm.Calls) == 0 || !bytes.Equal(cm.Calls[0].Arguments.Get(0).([]byte), exp) {
				t.Fatalf("selectFile APDU mismatch")
			}
			cm.AssertExpectations(t)
		})
	}
}

func TestApolloReadFile(t *testing.T) {
	tests := []struct {
		name      string
		responses []struct {
			rsp []byte
			err error
		}
		want      []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success",
			responses: []struct {
				rsp []byte
				err error
			}{
				{withStatus(nil), nil}, {withStatus(header(4)), nil}, {withStatus([]byte{1, 2, 3, 4}), nil},
			},
			want: []byte{1, 2, 3, 4},
		},
		{
			name: "select error",
			responses: []struct {
				rsp []byte
				err error
			}{
				{nil, errors.New("sel fail")},
			},
			wantErr:   true,
			errSubstr: "selecting file",
		},
		{
			name: "short header",
			responses: []struct {
				rsp []byte
				err error
			}{
				{withStatus(nil), nil}, {withStatus([]byte{0x01, 0x02, 0x03}), nil},
			},
			wantErr:   true,
			errSubstr: "file too short",
		},
		{
			name: "read data error",
			responses: []struct {
				rsp []byte
				err error
			}{
				{withStatus(nil), nil}, {withStatus(header(2)), nil}, {nil, errors.New("read fail")},
			},
			wantErr:   true,
			errSubstr: "reading file",
		},
		{
			name: "empty chunk stops",
			responses: []struct {
				rsp []byte
				err error
			}{
				{withStatus(nil), nil}, {withStatus(header(2)), nil}, {withStatus(nil), nil},
			},
			want: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			for _, r := range tt.responses {
				cm.On("Transmit", mock.Anything).Return(r.rsp, r.err).Once()
			}
			card := &Apollo{smartCard: cm}

			got, err := card.ReadFile([]byte{0x01, 0x02})
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.errSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("ReadFile() error = %v", err)
				}
				if !bytes.Equal(got, tt.want) {
					t.Fatalf("ReadFile() = %x, want %x", got, tt.want)
				}
			}
			cm.AssertExpectations(t)
		})
	}
}

func TestApolloAtr(t *testing.T) {
	card := &Apollo{atr: APOLLO_ATR}
	if got := card.Atr(); !bytes.Equal(got, APOLLO_ATR) {
		t.Fatalf("Atr() = %x, want %x", got, APOLLO_ATR)
	}
}

func TestApolloGetDocumentError(t *testing.T) {
	card := &Apollo{} // no files set -> parse should fail
	if _, err := card.GetDocument(); err == nil || !strings.Contains(err.Error(), "document file") {
		t.Fatalf("expected document file error, got %v", err)
	}
}

func TestApolloTest(t *testing.T) {
	if ok := (&Apollo{}).Test(); !ok {
		t.Fatalf("Test() = %v, want true", ok)
	}
}






