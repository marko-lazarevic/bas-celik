package card

import (
	"errors"

	"github.com/ubavic/bas-celik/v2/document"
)

// UnknownDocumentCard represents an unknown or unsupported card type.
type UnknownDocumentCard struct {
	atr       Atr
	smartCard Card
}

// Atr returns the ATR of the unknown card.
func (card *UnknownDocumentCard) Atr() Atr {
	return card.atr
}

// ReadFile is not implemented for unknown cards.
func (card *UnknownDocumentCard) ReadFile(_ []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// InitCard does nothing for unknown cards.
func (card *UnknownDocumentCard) InitCard() error {
	return nil
}

// ReadCard does nothing for unknown cards.
func (card *UnknownDocumentCard) ReadCard() error {
	return nil
}

// GetDocument returns nil for unknown cards.
func (card *UnknownDocumentCard) GetDocument() (document.Document, error) {
	return nil, nil
}

// Test always returns true for unknown cards.
func (card *UnknownDocumentCard) Test() bool {
	return true
}
