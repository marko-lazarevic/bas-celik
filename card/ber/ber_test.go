package ber

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/ubavic/bas-celik/v2/card/carderrors"
)

func Test_parseBerLength(t *testing.T) {
	testCases := []struct {
		data                []byte
		expectedLength      uint32
		expectedParsedBytes uint32
		expectedError       error
	}{
		{
			data:          []byte{},
			expectedError: carderrors.ErrInvalidLength,
		},
		{
			data:                []byte{0x79},
			expectedLength:      0x79,
			expectedParsedBytes: 1,
			expectedError:       nil,
		},
		{
			data:          []byte{0x80, 0x91},
			expectedError: carderrors.ErrInvalidFormat,
		},
		{
			data:                []byte{0x81, 0x01},
			expectedLength:      0x01,
			expectedParsedBytes: 2,
			expectedError:       nil,
		},
		{
			data:          []byte{0x81},
			expectedError: carderrors.ErrInvalidLength,
		},
		{
			data:                []byte{0x82, 0x01, 0x02},
			expectedLength:      uint32(0x01)<<8 + uint32(0x02),
			expectedParsedBytes: 3,
			expectedError:       nil,
		},
		{
			data:          []byte{0x82},
			expectedError: carderrors.ErrInvalidLength,
		},
		{
			data:                []byte{0x83, 0x01, 0x02, 0x03},
			expectedLength:      uint32(0x01)<<16 + uint32(0x02)<<8 + uint32(0x03),
			expectedParsedBytes: 4,
			expectedError:       nil,
		},
		{
			data:          []byte{0x83},
			expectedError: carderrors.ErrInvalidLength,
		},
		{
			data:                []byte{0x84, 0x01, 0x02, 0x03, 0x04},
			expectedLength:      uint32(0x01)<<24 + uint32(0x02)<<16 + uint32(0x03)<<8 + uint32(0x04),
			expectedParsedBytes: 5,
			expectedError:       nil,
		},
		{
			data:          []byte{0x84},
			expectedError: carderrors.ErrInvalidLength,
		},
	}

	for _, testCase := range testCases {
		length, parsedBytes, err := ParseLength(testCase.data)

		if err == nil && testCase.expectedError == nil {
			if length != testCase.expectedLength {
				t.Errorf("Expected parsed length to be %d, but it is %d", testCase.expectedLength, length)
			}
			if parsedBytes != testCase.expectedParsedBytes {
				t.Errorf("Expected %d bytes to be parsed, but %d bytes were parsed", testCase.expectedParsedBytes, parsedBytes)
			}
		} else {
			if err != testCase.expectedError {
				t.Errorf("Expected error '%v', but error is '%v'", testCase.expectedError, err)
			}
		}
	}
}

func Test_parseBerTag(t *testing.T) {
	testCases := []struct {
		data                []byte
		expectedTag         uint32
		expectedPrimitive   bool
		expectedParsedBytes uint32
		expectedError       error
	}{
		{
			data:          []byte{},
			expectedError: carderrors.ErrInvalidLength,
		},
		{
			data:                []byte{0b000001},
			expectedTag:         0b000001,
			expectedPrimitive:   true,
			expectedParsedBytes: 1,
			expectedError:       nil,
		},
		{
			data:                []byte{0b00100001},
			expectedTag:         0b00100001,
			expectedPrimitive:   false,
			expectedParsedBytes: 1,
			expectedError:       nil,
		},
		{
			data:                []byte{0b10111111, 0b00101111},
			expectedTag:         uint32(binary.BigEndian.Uint16([]byte{0b10111111, 0b00101111})),
			expectedPrimitive:   false,
			expectedParsedBytes: 2,
			expectedError:       nil,
		},
		{
			data:                []byte{0b10111111, 0b10101111},
			expectedTag:         uint32(binary.BigEndian.Uint16([]byte{0b10111111, 0b10101111})),
			expectedPrimitive:   false,
			expectedParsedBytes: 2,
			expectedError:       carderrors.ErrInvalidLength,
		},
		{
			data:                []byte{0b10111111, 0b10101111, 0b011010101},
			expectedTag:         uint32(binary.BigEndian.Uint32([]byte{0, 0b10111111, 0b10101111, 0b011010101})),
			expectedPrimitive:   false,
			expectedParsedBytes: 3,
			expectedError:       nil,
		},
	}

	for _, testCase := range testCases {
		tag, primitive, parsedBytes, err := ParseTag(testCase.data)
		if err == nil && testCase.expectedError == nil {
			if tag != testCase.expectedTag {
				t.Errorf("Expected tag be %d, but it is %d", testCase.expectedTag, tag)
			}
			if primitive != testCase.expectedPrimitive {
				t.Errorf("Expected primitive flag to be %t, but it is %t", testCase.expectedPrimitive, primitive)
			}
			if parsedBytes != testCase.expectedParsedBytes {
				t.Errorf("Expected %d bytes to be parsed, but %d bytes ware parsed", testCase.expectedParsedBytes, parsedBytes)
			}
		} else {
			if err != testCase.expectedError {
				t.Errorf("Expected error '%v', but error is '%v'", testCase.expectedError, err)
			}
		}
	}
}

func TestParseBER(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"ok primitive child", []byte{0x01, 0x01, 0xAA}, false},
		{"invalid data", []byte{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := ParseBER(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseBER() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(tree.children) != 1 || tree.children[0].tag != 0x01 || !bytes.Equal(tree.children[0].data, []byte{0xAA}) {
				t.Fatalf("unexpected tree %+v", tree.children)
			}
		})
	}
}

func Test_parseBERLayer(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"single primitive", []byte{0x01, 0x01, 0xAA}, false},
		{"length overflow", []byte{0x02, 0x02}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prim, cons, err := parseBERLayer(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseBERLayer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(prim) != 1 || len(cons) != 0 || !bytes.Equal(prim[0x01], []byte{0xAA}) {
				t.Fatalf("unexpected maps prim=%v cons=%v", prim, cons)
			}
		})
	}
}

func TestBERAccess(t *testing.T) {
	tree := BER{
		tag: 0, primitive: false,
		children: []BER{
			{tag: 1, primitive: false, children: []BER{
				{tag: 2, primitive: true, data: []byte("val")},
			}},
		},
	}
	tests := []struct {
		name    string
		address []uint32
		want    []byte
		wantErr bool
	}{
		{"found", []uint32{1, 2}, []byte("val"), false},
		{"missing", []uint32{3}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tree.access(tt.address...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("access() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("access() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBERAdd(t *testing.T) {
	tests := []struct {
		name       string
		tree       BER
		newNode    BER
		wantErr    bool
		checkChild uint32
	}{
		{"into primitive", BER{primitive: true}, BER{tag: 1}, true, 0},
		{"new child", BER{}, BER{tag: 1, primitive: true, data: []byte{0x01}}, false, 1},
		{"type mismatch", BER{children: []BER{{tag: 1, primitive: true}}}, BER{tag: 1, primitive: false}, true, 0},
		{"merge constructed", BER{children: []BER{{tag: 1, primitive: false, children: []BER{{tag: 2, primitive: true}}}}}, BER{tag: 1, primitive: false, children: []BER{{tag: 3, primitive: true}}}, false, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tree.add(tt.newNode)
			if (err != nil) != tt.wantErr {
				t.Fatalf("add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr || tt.checkChild == 0 {
				return
			}
			found := false
			for _, c := range tt.tree.children {
				if c.tag == tt.newNode.tag {
					for _, g := range c.children {
						if g.tag == tt.checkChild {
							found = true
						}
					}
					if c.primitive && tt.checkChild == tt.newNode.tag {
						found = bytes.Equal(c.data, tt.newNode.data)
					}
				}
			}
			if !found {
				t.Fatalf("expected child %d to exist after add", tt.checkChild)
			}
		})
	}
}

func TestBERMerge(t *testing.T) {
	tests := []struct {
		name    string
		base    BER
		other   BER
		wantErr bool
	}{
		{"tag mismatch", BER{tag: 1}, BER{tag: 2}, true},
		{"merge child", BER{tag: 1}, BER{tag: 1, children: []BER{{tag: 5, primitive: true, data: []byte{0x05}}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := tt.base
			err := base.Merge(tt.other)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Merge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(base.children) != 1 || base.children[0].tag != 5 {
				t.Fatalf("Merge() failed, children: %+v", base.children)
			}
		})
	}
}

func TestAssignFromAndString(t *testing.T) {
	tree := BER{tag: 0, primitive: false, children: []BER{{tag: 1, primitive: true, data: []byte("abc")}}}

	tests := []struct {
		name     string
		address  []uint32
		start    string
		expected string
	}{
		{"assign success", []uint32{1}, "", "abc"},
		{"assign missing", []uint32{2}, "keep", "keep"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.start
			tree.AssignFrom(&out, tt.address...)
			if out != tt.expected {
				t.Fatalf("AssignFrom() = %s, want %s", out, tt.expected)
			}
		})
	}

	if s := tree.String(); s != "0:\n  1: abc" {
		t.Fatalf("String() = %q, want %q", s, "0:\n  1: abc")
	}
}
