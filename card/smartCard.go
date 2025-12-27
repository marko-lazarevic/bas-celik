package card

import "github.com/ebfe/scard"

// VirtualCard represents a virtual smart card for testing purposes.
type VirtualCard struct {
	atr   []byte
	files map[uint32][]byte
}

// MakeVirtualCard creates a new virtual card with the given ATR and file system.
func MakeVirtualCard(atr []byte, fs map[uint32][]byte) *VirtualCard {
	vc := VirtualCard{
		atr:   atr,
		files: fs,
	}

	return &vc
}

// Status returns the status of the virtual card.
func (card *VirtualCard) Status() (*scard.CardStatus, error) {
	status := scard.CardStatus{Atr: card.atr, Reader: "Virtual", State: scard.Powered}
	return &status, nil
}

// Transmit simulates transmitting a command to the virtual card.
func Transmit(_ []byte) ([]byte, error) {
	return []byte{0x90, 0x00}, nil
}
