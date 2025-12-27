package server

import (
	"encoding/json"
	"fmt"
	"io"
)

// GetInfoInput represents the input parameters for the getInfo operation.
type GetInfoInput struct {
	SbSession string `json:"sbSession"`
	Language  string `json:"language"`
	Host      string `json:"host"`
}

// GetInfoPayload represents the response payload for the getInfo operation.
type GetInfoPayload struct {
	TerminalID    int    `json:"terminalId"`
	ProviderID    int    `json:"providerId"`
	CertificateID string `json:"certificateId"`
}

func (s *SmartBoxServer) handleGetInfo(sessionID *string, data []byte, w io.Writer) error {
	msg := Message[GetInfoInput]{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	if msg.Input.SbSession == "" {
		return fmt.Errorf("invalid session id")
	}

	session, ok := s.sessions[msg.Input.SbSession]
	if !ok {
		s.sessions[msg.Input.SbSession] = SmartboxSession{
			id: msg.Input.SbSession,
		}
	}

	*sessionID = msg.Input.SbSession

	rsp := Response[GetInfoPayload]{
		Operation: operationGetInfo,
		Payload: GetInfoPayload{
			TerminalID:    session.terminalID,
			ProviderID:    int(session.vendor),
			CertificateID: session.certificateID,
		},
	}

	return json.NewEncoder(w).Encode(rsp)
}
