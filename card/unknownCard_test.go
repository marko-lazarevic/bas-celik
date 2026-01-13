package card

import (
	"slices"
	"testing"
)

func TestAtr(t *testing.T) {
	expectedAtr := Atr{0x3B, 0x00, 0x00, 0x00}
	card := UnknownDocumentCard{atr: expectedAtr}
	atr := card.Atr()
	if !slices.Equal(atr, expectedAtr) {
		t.Errorf("Atr() = %v; want %v", atr, expectedAtr)
	}
}

func TestReadFile(t *testing.T) {
	card := UnknownDocumentCard{}
	_, err := card.ReadFile([]byte{0x00, 0xA4, 0x04, 0x00})
	if err == nil || err.Error() != "not implemented" {
		t.Errorf("ReadFile() error = %v; want 'not implemented'", err)
	}
}

func TestInitCard(t *testing.T) {
	card := UnknownDocumentCard{}
	err := card.InitCard()
	if err != nil {
		t.Errorf("InitCard() error = %v; want nil", err)
	}
}

func TestReadCard(t *testing.T) {
	card := UnknownDocumentCard{}
	err := card.ReadCard()
	if err != nil {
		t.Errorf("ReadCard() error = %v; want nil", err)
	}
}

func TestGetDocument(t *testing.T) {
	card := UnknownDocumentCard{}
	doc, err := card.GetDocument()
	if err != nil {
		t.Errorf("GetDocument() error = %v; want nil", err)
	}
	if doc != nil {
		t.Errorf("GetDocument() = %v; want nil", doc)
	}
}

func TestTest(t *testing.T) {
	card := UnknownDocumentCard{}
	result := card.Test()
	if result != true {
		t.Errorf("Test() = %v; want true", result)
	}
}