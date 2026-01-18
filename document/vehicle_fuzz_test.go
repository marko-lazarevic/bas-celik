package document_test

import (
	"testing"

	"github.com/ubavic/bas-celik/v2/document"
)

// FuzzVehicleDocumentBuildPdf ensures VehicleDocument.BuildPdf handles varied inputs.
func FuzzVehicleDocumentBuildPdf(f *testing.F) {
	setDocumentConfigFromLocalFilesFuzz(f)

	f.Add("BG123AA", "Petar", "Main St 1", "VW Golf")
	f.Add("NS987BB", "Ana", "Second St 2", "Tesla Model 3")

	f.Fuzz(func(t *testing.T, reg, owner, address, vehicle string) {
		doc := document.VehicleDocument{
			RegistrationNumberOfVehicle: reg,
			IssuingDate:                 "2020-01-01",
			ExpiryDate:                  "2030-01-01",
			StateIssuing:                "RS",
			AuthorityIssuing:            owner,
			CompetentAuthority:          address,
			UnambiguousNumber:           reg + "-U",
			SerialNumber:                reg + "-S",
			OwnersSurnameOrBusinessName: owner,
			OwnerName:                   owner,
			OwnerAddress:                address,
			OwnersPersonalNo:            "1234567890123",
			UsersSurnameOrBusinessName:  owner,
			UsersName:                   owner,
			UsersAddress:                address,
			UsersPersonalNo:             "3210987654321",
			VehicleMake:                 vehicle,
			VehicleType:                 vehicle,
			VehicleIDNumber:             reg + "VIN",
			VehicleCategory:             "M1",
			VehicleMass:                 "1500",
			MaximumPermissibleLadenMass: "2000",
			NumberOfSeats:               "5",
			NumberOfStandingPlaces:      "0",
			TypeOfFuel:                  "E",
			DateOfFirstRegistration:     "2019-01-01",
			EngineCapacity:              "1600",
			EngineIDNumber:              "ENG" + reg,
			MaximumNetPower:             "100",
			ColourOfVehicle:             "Blue",
			VehicleLoad:                 "500",
			YearOfProduction:            "2019",
		}

		if _, _, err := doc.BuildPdf(); err != nil {
			t.Fatalf("BuildPdf returned error: %v", err)
		}
	})
}
