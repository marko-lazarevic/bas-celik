package card

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ebfe/scard"
	"github.com/stretchr/testify/mock"
	testhelpers "github.com/ubavic/bas-celik/v2/test_helpers"
)

func Test_responseOK(t *testing.T) {
	testCases := []struct {
		value  []byte
		result bool
	}{
		{[]byte{0x0F, 0x0F}, false},
		{[]byte{0x90, 0x00}, true},
		{[]byte{0x01, 0xFF, 0x90, 0x00}, true},
		{[]byte{0x01, 0xFF, 0x00, 0x00}, false},
		{[]byte{0xA1}, false},
	}

	for i, testCase := range testCases {
		t.Run(
			fmt.Sprintf("Case %d", i),
			func(t *testing.T) {
				res := responseOK(testCase.value)

				if res != testCase.result {
					t.Errorf("Expected %t, but got %t", testCase.result, res)
				}
			},
		)
	}
}

func statusRsp(data []byte) []byte {
	out := append([]byte{}, data...)
	return append(out, 0x90, 0x00)
}

func Test_read(t *testing.T) {
	tests := []struct {
		name     string
		offset   uint
		length   uint
		resp     []byte
		respErr  error
		want     []byte
		wantErr  bool
		errSub   string
		wantP1P2 [2]byte
		wantLe   byte
	}{
		{
			name:     "success",
			offset:   0x0102,
			length:   2,
			resp:     statusRsp([]byte{0xAA, 0xBB}),
			want:     []byte{0xAA, 0xBB},
			wantP1P2: [2]byte{0x01, 0x02},
			wantLe:   0x02,
		},
		{
			name:    "transmit error",
			offset:  0,
			length:  1,
			respErr: errors.New("tx fail"),
			wantErr: true,
			errSub:  "reading binary",
		},
		{
			name:    "short response",
			offset:  0,
			length:  1,
			resp:    []byte{0x90},
			wantErr: true,
			errSub:  "bad status",
		},
		{
			name:     "length capped to 0xFF",
			offset:   0,
			length:   0x200,
			resp:     statusRsp([]byte{0xCC}),
			want:     []byte{0xCC},
			wantP1P2: [2]byte{0x00, 0x00},
			wantLe:   0xFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			cm.On("Transmit", mock.Anything).Return(tt.resp, tt.respErr).Once()

			got, err := read(cm, tt.offset, tt.length)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("expected error containing %q, got %v", tt.errSub, err)
				}
			} else {
				if err != nil {
					t.Fatalf("read() error = %v", err)
				}
				if string(got) != string(tt.want) {
					t.Fatalf("read() = %x, want %x", got, tt.want)
				}
				if call := cm.Calls[0]; tt.wantLe != 0 {
					apdu := call.Arguments.Get(0).([]byte)
					if apdu[2] != tt.wantP1P2[0] || apdu[3] != tt.wantP1P2[1] || apdu[len(apdu)-1] != tt.wantLe {
						t.Fatalf("APDU fields mismatch: %x", apdu)
					}
				}
			}
			cm.AssertExpectations(t)
		})
	}
}

func Test_trim4b(t *testing.T) {
	tests := []struct {
		in   []byte
		want []byte
	}{
		{[]byte{1, 2, 3, 4, 5, 6}, []byte{5, 6}},
		{[]byte{1, 2, 3}, []byte{1, 2, 3}},
	}
	for _, tt := range tests {
		if got := trim4b(tt.in); string(got) != string(tt.want) {
			t.Fatalf("trim4b(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func Test_DetectCardDocument(t *testing.T) {
	supported := []struct {
		name       string
		atr        []byte
		expectType CardDocumentType
	}{
		{"apollo", APOLLO_ATR, ApolloIDDocumentCardType},
		{"gemalto", GEMALTO_ATR_1, GemaltoIDDocumentCardType},
		{"vehicle", VEHICLE_ATR_2, VehicleDocumentCardType},
	}

	for _, tt := range supported {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if !slices.Contains(DetectCardDocumentByAtr(Atr(tt.atr)), tt.expectType) {
				t.Fatalf("ATR %x not supported in current build", tt.atr)
			}

			cm := &testhelpers.CardMock{}
			cm.On("Status").Return(&scard.CardStatus{Atr: tt.atr}, nil).Once()

			switch tt.name {
			case "gemalto":
				cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()                                       // InitCard select succeeds
				cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()                                       // selectFile in ReadFile
				cm.On("Transmit", mock.Anything).Return([]byte{0x00, 0x00, 0x04, 0x00, 0x90, 0x00}, nil).Once()               // header read: length=4
				cm.On("Transmit", mock.Anything).Return([]byte{0xDE, 0xAD, 0xBE, 0xEF, 0x90, 0x00}, nil).Once()               // body read
			case "vehicle":
				cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Times(3)
			}

			doc, err := DetectCardDocument(cm)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			switch tt.expectType {
			case ApolloIDDocumentCardType:
				if _, ok := doc.(*Apollo); !ok {
					t.Fatalf("expected Apollo card, got %T", doc)
				}
			case GemaltoIDDocumentCardType:
				if _, ok := doc.(*Gemalto); !ok {
					t.Fatalf("expected Gemalto card, got %T", doc)
				}
			case VehicleDocumentCardType:
				if _, ok := doc.(*VehicleCard); !ok {
					t.Fatalf("expected Vehicle card, got %T", doc)
				}
			}

			cm.AssertExpectations(t)
		})
	}

	t.Run("status error", func(t *testing.T) {
		cm := &testhelpers.CardMock{}
		cm.On("Status").Return((*scard.CardStatus)(nil), errors.New("status fail")).Once()

		if _, err := DetectCardDocument(cm); err == nil {
			t.Fatalf("expected error on status failure")
		}
		cm.AssertExpectations(t)
	})

	t.Run("unknown atr", func(t *testing.T) {
		cm := &testhelpers.CardMock{}
		cm.On("Status").Return(&scard.CardStatus{Atr: []byte{0x01, 0x02, 0x03}}, nil).Once()

		doc, err := DetectCardDocument(cm)
		if err == nil || !strings.Contains(err.Error(), "unknown") {
			t.Fatalf("expected unknown card error, got %v", err)
		}
		if _, ok := doc.(*UnknownDocumentCard); !ok {
			t.Fatalf("expected UnknownDocumentCard, got %T", doc)
		}
		cm.AssertExpectations(t)
	})
}