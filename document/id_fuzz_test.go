package document_test

import (
	"hash/crc32"
	"image"
	"os"
	"testing"

	"github.com/ubavic/bas-celik/v2/document"
)

// setDocumentConfigFromLocalFilesFuzz ensures BuildPdf handles varied text inputs without panicking.
func FuzzIDDocumentBuildPdf(f *testing.F) {
	setDocumentConfigFromLocalFilesFuzz(f)

	seedImage := image.NewRGBA(image.Rect(0, 0, 16, 16))

	f.Add("ID", "Ana", "Popovic", "Belgrade")
	f.Add("IF", "Λάμπρος", "Παπαδόπουλος", "Thessaloniki")
	f.Add("RP", "Petar", "Petrović", "Novi Sad")

	docTypes := []string{
		document.ID_TYPE_APOLLO,
		document.ID_TYPE_ID,
		document.ID_TYPE_IDENTITY_FOREIGNER,
		document.ID_TYPE_RESIDENCE_PERMIT,
	}

	f.Fuzz(func(t *testing.T, docType, givenName, surname, place string) {
		idx := crc32.ChecksumIEEE([]byte(docType)) % uint32(len(docTypes))
		normalizedType := docTypes[idx]

		doc := document.IDDocument{
			Portrait:             seedImage,
			DocumentType:         normalizedType,
			GivenName:            givenName,
			ParentGivenName:      place,
			Surname:              surname,
			Place:                place,
			Community:            place,
			Street:               place,
			HouseNumber:          "1",
			DocumentSerialNumber: place,
		}

		if _, _, err := doc.BuildPdf(); err != nil {
			t.Fatalf("BuildPdf returned error: %v", err)
		}
	})
}

// setDocumentConfigFromLocalFiles configures fonts/assets required by BuildPdf for fuzzing.
func setDocumentConfigFromLocalFilesFuzz(tb testing.TB) {
	tb.Helper()

	config := document.DocumentConfig{}

	var err error
	config.FontRegular, err = os.ReadFile("../embed/liberationSansRegular.ttf")
	if err != nil {
		tb.Fatal(err.Error())
	}

	config.FontBold, err = os.ReadFile("../embed/liberationSansBold.ttf")
	if err != nil {
		tb.Fatal(err.Error())
	}

	config.RfzoLogo, err = os.ReadFile("../embed/rfzo.png")
	if err != nil {
		tb.Fatal(err.Error())
	}

	if err := document.Configure(config); err != nil {
		tb.Fatalf("setting document config: %v", err)
	}
}
