// Package server implements the SmartBox server functionality.
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/ubavic/bas-celik/v2/internal/smartbox/pkcs11"
)

// GetTerminalsInput represents the input for the GetTerminals request.
type GetTerminalsInput struct {
	ProviderID stringOrInt `json:"providerId"`
}

// GetTerminalsPayload represents the payload for the GetTerminals response.
type GetTerminalsPayload struct {
	Terminals []Terminal `json:"terminals"`
}

// Terminal represents a smart card terminal.
type Terminal struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *SmartBoxServer) handleGetTerminals(session *SmartboxSession, data []byte, w io.Writer) error {
	msg := Message[GetTerminalsInput]{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	providerID := msg.Input.ProviderID

	if providerID < 0 || int(providerID) > int(pkcs11.CardVendorPks) {
		return fmt.Errorf("invalid provider id")
	}

	moduleIndex := slices.IndexFunc(s.modulePaths, func(m ModulePath) bool {
		return m.Vendor == pkcs11.CardVendor(providerID)
	})

	if moduleIndex < 0 {
		return fmt.Errorf("module not found")
	}

	module, err := pkcs11.NewPkcsExternalModule(s.modulePaths[moduleIndex].Path)
	if err != nil {
		return err
	}

	session.vendor = pkcs11.CardVendor(providerID)
	session.module = &module

	slotIDs, slotNames, err := module.ListSlots()
	if err != nil {
		return err
	}

	terminals := make([]Terminal, 0, len(slotIDs))
	for i, id := range slotIDs {
		terminals = append(terminals, Terminal{
			ID:   fmt.Sprintf("%d", id),
			Name: slotNames[i],
		})
	}

	rsp := Response[GetTerminalsPayload]{
		Operation: operationGetTerminals,
		Payload: GetTerminalsPayload{
			Terminals: terminals,
		},
	}

	return json.NewEncoder(w).Encode(rsp)
}
