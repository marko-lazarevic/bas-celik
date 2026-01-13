package card

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/ubavic/bas-celik/v2/card/carderrors"
	"github.com/ubavic/bas-celik/v2/document"
	testhelpers "github.com/ubavic/bas-celik/v2/test_helpers"
)

func Test_parseVehicleCardFileSize(t *testing.T) {
	testCases := []struct {
		data           []byte
		expectedLength uint
		expectedOffset uint
		expectedError  error
	}{
		{
			data:          []byte{},
			expectedError: carderrors.ErrInvalidLength,
		},
		{
			data:          []byte{0x01, 0x02, 0x03, 0x04},
			expectedError: carderrors.ErrInvalidLength,
		},
		{
			data: []byte{
				0x78, 0x0E, 0x4F, 0x0C, 0xA0, 0x00, 0x00, 0x00,
				0x18, 0x65, 0x56, 0x4C, 0x2D, 0x30, 0x30, 0x31,
				0x72,
			},
			expectedError: carderrors.ErrInvalidLength,
		},
		{
			data: []byte{
				0x78, 0x0E, 0x4F, 0x0C, 0xA0, 0x00, 0x00, 0x00,
				0x18, 0x65, 0x56, 0x4C, 0x2D, 0x30, 0x30, 0x31,
				0x72, 0x27,
			},
			expectedLength: 41,
			expectedOffset: 16,
			expectedError:  nil,
		},
		{
			data:          []byte{0x01, 0x01, 0x01, 0x00, 0x80},
			expectedError: carderrors.ErrInvalidFormat,
		},
	}

	for _, testCase := range testCases {
		length, offset, err := parseVehicleCardFileSize(testCase.data)
		if err == nil && testCase.expectedError == nil {
			if length != testCase.expectedLength {
				t.Errorf("Expected length to be %d, but it is %d", testCase.expectedLength, length)
			}
			if offset != testCase.expectedOffset {
				t.Errorf("Expected offset to be %d, but it is %d", testCase.expectedOffset, offset)
			}
		} else {
			if !errors.Is(err, testCase.expectedError) {
				t.Errorf("Expected error to be '%v', but it is '%v'", testCase.expectedError, err)
			}
		}
	}
}

func TestVehicleCardReadCard(t *testing.T) {
	validHeader := []byte{
		0x78, 0x0E, 0x4F, 0x0C, 0xA0, 0x00, 0x00, 0x00,
		0x18, 0x65, 0x56, 0x4C, 0x2D, 0x30, 0x30, 0x31,
		0x72, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	// parseVehicleCardFileSize expects 41 bytes based on this header
	fileData := make([]byte, 41)
	for i := range fileData {
		fileData[i] = byte(i)
	}

	okSelect := func() []byte { return []byte{0x90, 0x00} }
	okHeader := func() []byte { return append(validHeader, 0x90, 0x00) }
	okBody := func() []byte { return append(fileData, 0x90, 0x00) }

	makeSuccess := func(cm *testhelpers.CardMock) {
		cm.On("Transmit", mock.Anything).Return(okSelect(), nil).Once()
		cm.On("Transmit", mock.Anything).Return(okHeader(), nil).Once()
		cm.On("Transmit", mock.Anything).Return(okBody(), nil).Once()
	}

	tests := []struct {
		name      string
		setup     func(cm *testhelpers.CardMock)
		wantErr   bool
		errSubstr string
		assertOK  func(t *testing.T, card *VehicleCard)
	}{
		{
			name: "success",
			setup: func(cm *testhelpers.CardMock) {
				makeSuccess(cm)
				makeSuccess(cm)
				makeSuccess(cm)
				makeSuccess(cm)
			},
			assertOK: func(t *testing.T, card *VehicleCard) {
				for i := 0; i < 4; i++ {
					if len(card.files[i]) == 0 {
						t.Fatalf("File %d was not populated", i)
					}
				}
			},
		},
		{
			name: "document 0 error",
			setup: func(cm *testhelpers.CardMock) {
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail")).Once()
			},
			wantErr:   true,
			errSubstr: "reading document 0 file",
		},
		{
			name: "document 1 error",
			setup: func(cm *testhelpers.CardMock) {
				makeSuccess(cm)
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail")).Once()
			},
			wantErr:   true,
			errSubstr: "reading document 1 file",
		},
		{
			name: "document 2 error",
			setup: func(cm *testhelpers.CardMock) {
				makeSuccess(cm)
				makeSuccess(cm)
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail")).Once()
			},
			wantErr:   true,
			errSubstr: "reading document 2 file",
		},
		{
			name: "document 3 error",
			setup: func(cm *testhelpers.CardMock) {
				makeSuccess(cm)
				makeSuccess(cm)
				makeSuccess(cm)
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail")).Once()
			},
			wantErr:   true,
			errSubstr: "reading document 3 file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			if tt.setup != nil {
				tt.setup(cm)
			}

			card := &VehicleCard{
				atr:       VEHICLE_ATR_2,
				smartCard: cm,
			}

			err := card.ReadCard()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.errSubstr, err)
				}
			} else if err != nil {
				t.Fatalf("ReadCard() error = %v", err)
			}

			if tt.assertOK != nil {
				tt.assertOK(t, card)
			}
			cm.AssertExpectations(t)
		})
	}
}

func TestVehicleCardGetDocument(t *testing.T) {
	// Helper to create simple BER-encoded primitive data
	makePrimitive := func(tag byte, data []byte) []byte {
		length := byte(len(data))
		result := []byte{tag, length}
		result = append(result, data...)
		return result
	}

	// Helper to create constructed BER data
	makeConstructed := func(tag byte, children []byte) []byte {
		length := byte(len(children))
		result := []byte{tag | 0x20, length} // Set constructed bit (bit 6)
		result = append(result, children...)
		return result
	}

	// Create minimal valid BER data for vehicle files
	// File 0 with constructed 0x71 tag
	inner0 := []byte{}
	inner0 = append(inner0, makePrimitive(0x81, []byte("BG123AB"))...)          // RegistrationNumberOfVehicle
	inner0 = append(inner0, makePrimitive(0x82, []byte("20200101"))...)         // DateOfFirstRegistration
	inner0 = append(inner0, makePrimitive(0x8A, []byte("VIN123456789"))...)     // VehicleIDNumber
	inner0 = append(inner0, makePrimitive(0x8D, []byte("20251231"))...)         // ExpiryDate
	inner0 = append(inner0, makePrimitive(0x8E, []byte("20200115"))...)         // IssuingDate
	file0 := makeConstructed(0x71, inner0)

	// File 1 with constructed 0x72 tag
	inner1 := []byte{}
	inner1 = append(inner1, makePrimitive(0x98, []byte("M1"))...)               // VehicleCategory
	inner1 = append(inner1, makePrimitive(0xC5, []byte("2019"))...)             // YearOfProduction
	file1 := makeConstructed(0x72, inner1)

	// Minimal BER-encoded empty files to satisfy parser
	file2 := []byte{0x01, 0x00} // primitive tag 0x01, length 0
	file3 := []byte{0x02, 0x00} // primitive tag 0x02, length 0

	tests := []struct {
		name      string
		files     func() *VehicleCard
		wantErr   bool
		errSubstr string
		assertOK  func(t *testing.T, doc *document.VehicleDocument)
	}{
		{
			name: "success",
			files: func() *VehicleCard {
				return &VehicleCard{
					files: [4][]byte{file0, file1, file2, file3},
				}
			},
			assertOK: func(t *testing.T, doc *document.VehicleDocument) {
				if doc.RegistrationNumberOfVehicle != "BG123AB" {
					t.Fatalf("expected RegistrationNumberOfVehicle=BG123AB, got %s", doc.RegistrationNumberOfVehicle)
				}
				if doc.VehicleIDNumber != "VIN123456789" {
					t.Fatalf("expected VehicleIDNumber=VIN123456789, got %s", doc.VehicleIDNumber)
				}
				if doc.VehicleCategory != "M1" {
					t.Fatalf("expected VehicleCategory=M1, got %s", doc.VehicleCategory)
				}
				if doc.YearOfProduction != "2019" {
					t.Fatalf("expected YearOfProduction=2019, got %s", doc.YearOfProduction)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.files()
			d, err := card.GetDocument()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.errSubstr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetDocument() error = %v", err)
			}
			dc, ok := d.(*document.VehicleDocument)
			if !ok {
				t.Fatalf("unexpected document type %T", d)
			}
			if tt.assertOK != nil {
				tt.assertOK(t, dc)
			}
		})
	}
}



