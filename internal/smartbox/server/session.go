// Package server implements the Smartbox server functionality.
package server

import (
	"os"

	"github.com/ubavic/bas-celik/v2/internal/smartbox/pkcs11"
)

// PkcsModuleSession defines the interface for PKCS#11 module sessions.
type PkcsModuleSession interface {
	ListSlots() ([]uint, []string, error)
	OpenSessionAndLogin(pin string, terminalIndex int) error
	GetCertificates() ([]pkcs11.NamedCert, error)
	Sign(certID []byte, message []byte) ([]byte, error)
	CloseSession() error
}

// SmartboxSession represents a session with the Smartbox server.
type SmartboxSession struct {
	id            string
	module        PkcsModuleSession
	terminalID    int
	vendor        pkcs11.CardVendor
	certificateID string
}

// ModulePath represents the path to a PKCS#11 module along with its vendor.
type ModulePath struct {
	Vendor pkcs11.CardVendor
	Path   string
}

func (s *SmartBoxServer) setModulePaths(paths []ModulePath) int {
	s.modulePaths = make([]ModulePath, 0, len(paths))

	for _, mPath := range paths {
		if mPath.Path == "" {
			continue
		}

		fInfo, err := os.Stat(mPath.Path)
		if err != nil {
			continue
		}

		if fInfo.IsDir() {
			continue
		}

		s.modulePaths = append(s.modulePaths, mPath)
	}

	return len(s.modulePaths)
}
