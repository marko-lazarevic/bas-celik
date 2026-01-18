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

var (
	linuxVendorPaths = map[CardVendor]string{
		CardVendorPosta:  "/usr/lib/libaetpkss.so",
		CardVendorEsmart: "/usr/lib/libeToken.so",
	}

	darwinVendorPaths = map[CardVendor]string{
		CardVendorPosta:  "/Applications/tokenadmin.app/Contents/Frameworks/libaetpkss.dylib",
		CardVendorEsmart: "/Library/Frameworks/eToken.framework/Versions/A/libIDPrimePKCS11.dylib",
		CardVendorHalcom: "/Applications/Personal.app/Contents/Frameworks/libtokenapi.dylib",
	}

	windowsVendorPaths = map[CardVendor]string{
		CardVendorHalcom: "C:\\Program Files (x86)\\Personal\\bin64\\personal64.dll",
		CardVendorPosta:  "C:\\Windows\\System32\\aetpkss1.dll",
		CardVendorEsmart: "C:\\Program Files\\SafeNet\\Authentication\\SAC\\x64\\IDPrimePKCS1164.dll",
		CardVendorMup:    "C:\\Program Files\\TrustEdgeID\\netsetpkcs11_x64.dll",
		CardVendorPks:    "C:\\Program Files\\TrustEdgeID\\netsetpkcs11_x64.dll",
	}
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
	switch os {
	case "linux":
		if path, ok := linuxVendorPaths[cv]; ok {
			return path
		}
	case "darwin":
		if path, ok := darwinVendorPaths[cv]; ok {
			return path
		}
	case "windows":
		if path, ok := windowsVendorPaths[cv]; ok {
			return path
		}
	}

	return ""
}
