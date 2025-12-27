package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ubavic/bas-celik/v2/internal/smartbox/pkcs11"
)

// GetCertificatesInput represents the input parameters for retrieving certificates from a smart card terminal.
type GetCertificatesInput struct {
	TerminalID int    `json:"terminalId"`
	Pin        string `json:"pin"`
}

// GetCertificatesPayload represents the response payload containing a list of certificate aliases.
type GetCertificatesPayload struct {
	Certificates []CertificateAlias `json:"certificates"`
}

// CertificateAlias represents a certificate with its alias and common name.
type CertificateAlias struct {
	Alias string `json:"alias"`
	Name  string `json:"name"`
}

func (s *SmartBoxServer) handleGetCertificates(session *SmartboxSession, data []byte, w io.Writer) error {
	msg := Message[GetCertificatesInput]{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	if session.module == nil {
		return fmt.Errorf("pkcs11 module not loaded")
	}

	err := session.module.OpenSessionAndLogin(msg.Input.Pin, msg.Input.TerminalID)
	if err != nil {
		return err
	}

	certs, err := session.module.GetCertificates()
	if err != nil {
		return err
	}

	rsp := Response[GetCertificatesPayload]{
		Operation: operationGetCertificates,
		Payload: GetCertificatesPayload{
			Certificates: GetCertificateAliases(GetValidCertificates(certs)),
		},
	}

	return json.NewEncoder(w).Encode(rsp)
}

// GetValidCertificates filters a list of certificates and returns only those that are currently valid based on their validity period.
func GetValidCertificates(namedCerts []pkcs11.NamedCert) []pkcs11.NamedCert {
	now := time.Now()

	validNamedCertificates := make([]pkcs11.NamedCert, 0, len(namedCerts))
	for _, namedCert := range namedCerts {
		if now.Before(namedCert.Certificate.NotBefore) {
			continue
		}

		if now.After(namedCert.Certificate.NotAfter) {
			continue
		}

		validNamedCertificates = append(validNamedCertificates, namedCert)
	}

	return validNamedCertificates
}

// GetCertificateAliases converts a list of named certificates into a list of certificate aliases containing ID and common name.
func GetCertificateAliases(namedCerts []pkcs11.NamedCert) []CertificateAlias {
	aliases := make([]CertificateAlias, 0, len(namedCerts))
	for _, namedCert := range namedCerts {
		alias := CertificateAlias{
			Alias: hex.EncodeToString(namedCert.ID),
			Name:  namedCert.Certificate.Subject.CommonName,
		}
		aliases = append(aliases, alias)
	}

	return aliases
}
