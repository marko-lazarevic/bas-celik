package server

import (
	"encoding/json"
	"io"
)

// GetProvidersInput represents the input parameters for retrieving available PKCS#11 providers.
type GetProvidersInput struct{}

// GetProvidersPayload represents the response payload containing a list of available providers.
type GetProvidersPayload struct {
	Providers []Provider `json:"providers"`
}

// Provider represents a PKCS#11 provider with its ID and name.
type Provider struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (s *SmartBoxServer) handleGetProviders(data []byte, w io.Writer) error {
	msg := Message[GetProvidersInput]{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	providers := make([]Provider, 0, len(s.modulePaths))
	for _, module := range s.modulePaths {
		providers = append(providers, Provider{
			ID:   int(module.Vendor),
			Name: module.Vendor.String(),
		})
	}

	rsp := Response[GetProvidersPayload]{
		Operation: operationGetProviders,
		Payload: GetProvidersPayload{
			Providers: providers,
		},
	}

	return json.NewEncoder(w).Encode(rsp)
}
