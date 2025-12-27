package card

import (
	"encoding/hex"
	"slices"
)

// Atr represents the Answer To Reset bytes from a smart card.
type Atr []byte

// String returns the hexadecimal string representation of the ATR.
func (atr Atr) String() string {
	return hex.EncodeToString([]byte(atr))
}

// Is compares this ATR with another ATR and returns true if they match.
func (atr Atr) Is(otherAtr Atr) bool {
	return slices.Equal(atr, otherAtr)
}

// DetectCardDocumentByAtr detects the type of card document based on its ATR.
func DetectCardDocumentByAtr(atr Atr) []CardDocumentType {
	if atr.Is(GEMALTO_ATR_1) {
		return []CardDocumentType{GemaltoIDDocumentCardType, VehicleDocumentCardType}
	} else if atr.Is(GEMALTO_ATR_2) || atr.Is(GEMALTO_ATR_3) {
		return []CardDocumentType{GemaltoIDDocumentCardType, MedicalDocumentCardType, VehicleDocumentCardType}
	} else if atr.Is(GEMALTO_ATR_4) {
		return []CardDocumentType{GemaltoIDDocumentCardType, VehicleDocumentCardType}
	} else if atr.Is(MEDICAL_ATR_1) || atr.Is(MEDICAL_ATR_2) {
		return []CardDocumentType{MedicalDocumentCardType}
	} else if atr.Is(VEHICLE_ATR_0) || atr.Is(VEHICLE_ATR_2) || atr.Is(VEHICLE_ATR_3) || atr.Is(VEHICLE_ATR_4) {
		return []CardDocumentType{VehicleDocumentCardType}
	} else if atr.Is(APOLLO_ATR) {
		return []CardDocumentType{ApolloIDDocumentCardType}
	}
	return []CardDocumentType{UnknownDocumentCardType}
}
