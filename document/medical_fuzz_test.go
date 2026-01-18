package document_test

import (
	"testing"

	"github.com/ubavic/bas-celik/v2/document"
)

// FuzzMedicalDocumentBuildPdf ensures MedicalDocument.BuildPdf handles varied inputs.
func FuzzMedicalDocumentBuildPdf(f *testing.F) {
	setDocumentConfigFromLocalFilesFuzz(f)

	f.Add("Ana", "Popovic", "Belgrade", "2000-01-01", "12345678901")
	f.Add("Mila", "Markovic", "Novi Sad", "1990-12-12", "09876543210")

	f.Fuzz(func(t *testing.T, givenName, familyName, place, dob, cardID string) {
		doc := document.MedicalDocument{
			GivenName:              givenName,
			FamilyName:             familyName,
			GivenNameLatin:         givenName,
			FamilyNameLatin:        familyName,
			ParentName:             place,
			ParentNameLatin:        place,
			Place:                  place,
			Municipality:           place,
			Country:                place,
			Street:                 place,
			Number:                 "1",
			Apartment:              "2",
			DateOfBirth:            dob,
			InsuranceStartDate:     "2020-01-01",
			InsuranceDescription:   "desc",
			CardID:                 cardID,
			InsurantNumber:         "11111111111",
			PersonalNumber:         "22222222222",
			InsuranceBasisRZZO:     "basis",
			ValidUntil:             "2030-01-01",
			DateOfIssue:            "2020-01-01",
			DateOfExpiry:           "2030-01-01",
			Gender:                 "F",
			CarrierFamilyMember:    true,
			CarrierRelationship:    "member",
			CarrierIDNumber:        "33333333333",
			CarrierInsurantNumber:  "44444444444",
			CarrierGivenName:       givenName,
			CarrierFamilyName:      familyName,
			CarrierGivenNameLatin:  givenName,
			CarrierFamilyNameLatin: familyName,
			TaxpayerName:           place,
			TaxpayerResidence:      place,
			TaxpayerNumber:         "55555555555",
			TaxpayerIDNumber:       "66666666666",
			TaxpayerActivityCode:   "123",
		}

		if _, _, err := doc.BuildPdf(); err != nil {
			t.Fatalf("BuildPdf returned error: %v", err)
		}
	})
}
