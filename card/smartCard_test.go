package card

import (
	"slices"
	"testing"

	"github.com/ebfe/scard"
)

func TestMakeVirtualCard(t *testing.T) {
	atr := []byte{0x00, 0x01, 0x02, 0x0}
	fs := map[uint32][]byte{
		0x00010001: {0x01, 0x02, 0x03, 0x04},
		0x00010002: {0x0A, 0x0B},
	}

	virtualCard := MakeVirtualCard(atr, fs)

	if !slices.Equal(virtualCard.atr, atr) {
		t.Errorf("ATR mismatch: expected %v, got %v", atr, virtualCard.atr)
	}
	
	for fileID, expectedData := range fs {
		actualData, exists := virtualCard.files[fileID]
		if !exists {
			t.Errorf("File ID %08X not found in virtual card files", fileID)
			continue
		}
		if !slices.Equal(actualData, expectedData) {
			t.Errorf("Data mismatch for File ID %08X: expected %v, got %v", fileID, expectedData, actualData)
		}
	}
}

func TestVirtualCard_Status(t *testing.T) {
	atr := []byte{0x3B, 0x9F, 0x96, 0x00, 0xFF, 0x91, 0xFE, 0x54, 0x43, 0x4B, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36}
	virtualCard := MakeVirtualCard(atr, nil)

	status, err := virtualCard.Status()
	if err != nil {
		t.Fatalf("Unexpected error from Status(): %v", err)
	}

	if !slices.Equal(status.Atr, atr) {
		t.Errorf("ATR mismatch: expected %v, got %v", atr, status.Atr)
	}
	if status.Reader != "Virtual" {
		t.Errorf("Reader mismatch: expected 'Virtual', got '%s'", status.Reader)
	}
	if status.State != scard.Powered {
		t.Errorf("State mismatch: expected %v, got %v", scard.Powered, status.State)
	}
}

func TestTransmit(t *testing.T) {
	response, err := Transmit([]byte{0x00, 0xA4, 0x04, 0x00})
	if err != nil {
		t.Fatalf("Unexpected error from Transmit(): %v", err)
	}

	expectedResponse := []byte{0x90, 0x00}
	if !slices.Equal(response, expectedResponse) {
		t.Errorf("Response mismatch: expected %v, got %v", expectedResponse, response)
	}
}