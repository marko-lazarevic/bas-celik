// Package pkcs11 provides utilities for handling PKCS#11 card vendors and their library paths.
package pkcs11

// CardVendor represents different PKCS#11 card vendors.
type CardVendor int

const (
	// CardVendorHalcom represents Halcom CA.
	CardVendorHalcom = CardVendor(iota)
	// CardVendorPosta represents Sertifikaciono telo Pošte.
	CardVendorPosta
	// CardVendorEsmart represents E-Smart Systems d.o.o. Beograd (ESS QCA).
	CardVendorEsmart
	// CardVendorMup represents Ministarstvo unutrašnjih poslova CA.
	CardVendorMup
	// CardVendorPks represents Privredna komora Srbije CA.
	CardVendorPks
)

// String returns the string representation of the CardVendor.
func (cv CardVendor) String() string {
	switch cv {
	case CardVendorHalcom:
		return "Halcom CA"
	case CardVendorPosta:
		return "Sertifikaciono telo Pošte"
	case CardVendorEsmart:
		return "E-Smart Systems d.o.o. Beograd (ESS QCA)"
	case CardVendorMup:
		return "Ministarstvo unutrašnjih poslova CA"
	case CardVendorPks:
		return "Privredna komora Srbije CA"
	default:
		return ""
	}
}

// GetDefaultPath returns the default PKCS#11 library path for the given card vendor and operating system.
func GetDefaultPath(cv CardVendor, os string) string {
	if os == "linux" {
		switch cv {
		case CardVendorPosta:
			return "/usr/lib/libaetpkss.so"
		case CardVendorEsmart:
			return "/usr/lib/libeToken.so"
		default:
			return ""
		}
	}

	if os == "darwin" {
		switch cv {
		case CardVendorPosta:
			return "/Applications/tokenadmin.app/Contents/Frameworks/libaetpkss.dylib"
		case CardVendorEsmart:
			return "/Library/Frameworks/eToken.framework/Versions/A/libIDPrimePKCS11.dylib"
		case CardVendorHalcom:
			return "/Applications/Personal.app/Contents/Frameworks/libtokenapi.dylib"
		default:
			return ""
		}
	}

	if os == "windows" {
		switch cv {
		case CardVendorHalcom:
			return "C:\\Program Files (x86)\\Personal\\bin64\\personal64.dll"
		case CardVendorPosta:
			return "C:\\Windows\\System32\\aetpkss1.dll"
		case CardVendorEsmart:
			return "C:\\Program Files\\SafeNet\\Authentication\\SAC\\x64\\IDPrimePKCS1164.dll"
		case CardVendorMup:
			return "C:\\Program Files\\TrustEdgeID\\netsetpkcs11_x64.dll"
		case CardVendorPks:
			return "C:\\Program Files\\TrustEdgeID\\netsetpkcs11_x64.dll"
		default:
			return ""
		}
	}

	return ""
}
