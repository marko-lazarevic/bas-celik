package document

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"github.com/ubavic/bas-celik/v2/localization"
)

const rfzoServiceURL = "https://www.rfzo.rs/proveraUplateDoprinosa2.php"

// ErrInvalidCardNo is returned when the card number doesn't have exactly 11 digits.
var ErrInvalidCardNo = errors.New("invalid card number length")

// ErrInvalidInsuranceNo is returned when the insurance number doesn't have exactly 11 digits.
var ErrInvalidInsuranceNo = errors.New("invalid insurance number length")

// ErrNoSubmatchFound is returned when the date ValidUntil could not be extracted from RFZO response.
var ErrNoSubmatchFound = errors.New("no submatch found")

// MedicalDocument represents a document stored on a Serbian public medical insurance card.
type MedicalDocument struct {
	InsurerName            string
	InsurerID              string
	CardID                 string
	DateOfIssue            string
	DateOfExpiry           string
	ChipSerialNumber       string
	PrintLanguage          string
	PersonalNumber         string
	FamilyNameLatin        string
	GivenNameLatin         string
	ParentNameLatin        string
	FamilyName             string
	GivenName              string
	ParentName             string
	Gender                 string
	InsurantNumber         string
	DateOfBirth            string
	Apartment              string
	Number                 string
	Street                 string
	Place                  string
	Municipality           string
	Country                string
	ValidUntil             string
	PermanentlyValid       bool
	CarrierGivenNameLatin  string
	CarrierFamilyNameLatin string
	CarrierGivenName       string
	CarrierFamilyName      string
	CarrierIDNumber        string
	CarrierInsurantNumber  string
	CarrierFamilyMember    bool
	CarrierRelationship    string
	InsuranceBasisRZZO     string
	InsuranceStartDate     string
	InsuranceDescription   string
	TaxpayerName           string
	TaxpayerResidence      string
	TaxpayerNumber         string
	TaxpayerIDNumber       string
	TaxpayerActivityCode   string
}

// GetFullName returns the full name of the medical document holder.
func (doc *MedicalDocument) GetFullName() string {
	return localization.JoinWithComma(doc.GivenNameLatin, doc.ParentNameLatin, doc.FamilyNameLatin)
}

// GetFullStreetAddress returns the full street address.
func (doc *MedicalDocument) GetFullStreetAddress() string {
	var address strings.Builder

	address.WriteString(doc.Street)
	if len(doc.Number) > 0 {
		address.WriteString(", Број: ")
		address.WriteString(doc.Number)
	}

	if len(doc.Apartment) > 0 {
		address.WriteString(" Стан: ")
		address.WriteString(doc.Apartment)
	}

	return address.String()
}

// GetFullPlaceAddress returns the full place address.
func (doc *MedicalDocument) GetFullPlaceAddress() string {
	return localization.JoinWithComma(doc.Place, doc.Municipality, doc.Country)
}

func putMedicalData(pdf *gopdf.GoPdf, textLeftMargin float64, label string, data string) {
		cell(pdf, label)
		pdf.SetXY(textLeftMargin+144, pdf.GetY())

		texts, err := pdf.SplitTextWithWordWrap(data, 350)
		if err != nil && err != gopdf.ErrEmptyString {
			panic(fmt.Errorf("splitting text: %w", err))
		}

		for i, text := range texts {
			cell(pdf, text)
			if i < len(texts)-1 {
				pdf.SetXY(textLeftMargin+144, pdf.GetY()+11)
			}
		}

		pdf.SetXY(textLeftMargin, pdf.GetY()+14)
}

// BuildPdf creates a PDF representation of the MedicalDocument.
func (doc *MedicalDocument) BuildPdf() (data []byte, fileName string, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case error:
				retErr = x
			default:
				retErr = errors.New("unknown panic")
			}
		}
	}()

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	pdf.AddPage()

	err := pdf.AddTTFFontData("liberationsans", fontRegular)
	if err != nil {
		panic(fmt.Errorf("loading font: %w", err))
	}

	err = pdf.SetFont("liberationsans", "", 11)
	if err != nil {
		panic(fmt.Errorf("setting font: %w", err))
	}

	const leftMargin = 30.464
	const rightMargin = 535
	const textLeftMargin = 38.243

	section := func(name string) {
		y := pdf.GetY() + 8
		pdf.Line(leftMargin, y, rightMargin, y)
		pdf.SetXY(textLeftMargin, y+12)
		cell(&pdf, name)
		pdf.Line(leftMargin, y+32, rightMargin, y+32)
		pdf.SetXY(textLeftMargin, y+41)
	}

	

	pdf.SetLineWidth(0.58)
	pdf.SetLineType("solid")

	rfzoLogoImage, _, err := image.Decode(bytes.NewReader(rfzoLogo))
	if err != nil {
		panic(fmt.Errorf("decoding photo file: %w", err))
	}

	err = pdf.ImageFrom(rfzoLogoImage, 36.0, 14.0, &gopdf.Rect{W: 144, H: 51})
	if err != nil {
		panic(fmt.Errorf("inserting logo: %w", err))
	}

	pdf.SetXY(218.6, 49.6)
	cell(&pdf, "ПРЕГЛЕД КАРТИЦЕ ЗДРАВСТВЕНОГ ОСИГУРАЊА (КЗО)")

	pdf.SetY(68)
	section("Општи подаци о осигуранику")

	putMedicalData(&pdf,textLeftMargin, "Име:", doc.GivenName+" ("+doc.GivenNameLatin+")")

	putMedicalData(&pdf, textLeftMargin, "Име једног родитеља:", doc.ParentName+" ("+doc.ParentNameLatin+")")

	putMedicalData(&pdf, textLeftMargin, "Презиме:", doc.FamilyName+" ("+doc.FamilyNameLatin+")")	
	putMedicalData(&pdf, textLeftMargin, "Датум рођења:", doc.DateOfBirth)

	putMedicalData(&pdf, textLeftMargin, "Место, општина и држава:", doc.GetFullPlaceAddress())

	putMedicalData(&pdf, textLeftMargin, "Улица:", doc.GetFullStreetAddress())

	putMedicalData(&pdf, textLeftMargin, "Пол:", doc.Gender)

	putMedicalData(&pdf, textLeftMargin, "Језик:", doc.PrintLanguage)

	putMedicalData(&pdf, textLeftMargin, "ЛБО:", doc.InsurantNumber)

	putMedicalData(&pdf, textLeftMargin, "ЈМБГ:", doc.PersonalNumber)

	section("Подаци о картици здравственог осигурања")

	putMedicalData(&pdf, textLeftMargin, "Датум издавања:", doc.DateOfIssue)

	putMedicalData(&pdf, textLeftMargin, "Датум важења:", doc.DateOfExpiry)

	putMedicalData(&pdf, textLeftMargin, "Оверена до:", doc.ValidUntil)

	putMedicalData(&pdf, textLeftMargin, "Трајно оверена:", localization.FormatYesNo(doc.PermanentlyValid, localization.SrCyrillic))

	section("Подаци о носиоцу осигурања")

	putMedicalData(&pdf, textLeftMargin, "Име:", doc.CarrierGivenName+" ("+doc.CarrierGivenNameLatin+")")

	putMedicalData(&pdf, textLeftMargin, "Презиме:", doc.CarrierFamilyName+" ("+doc.CarrierFamilyName+")")

	putMedicalData(&pdf, textLeftMargin, "ЛБО:", doc.CarrierInsurantNumber)

	putMedicalData(&pdf, textLeftMargin, "ЈМБГ:", doc.CarrierIDNumber)

	putMedicalData(&pdf, textLeftMargin, "Члан породице:", localization.FormatYesNo(doc.CarrierFamilyMember, localization.SrCyrillic))

	putMedicalData(&pdf, textLeftMargin, "Сродство:", doc.CarrierRelationship)

	section("Подаци о осигурању")

	putMedicalData(&pdf, textLeftMargin, "Основ осигурања:", doc.InsuranceBasisRZZO)

	putMedicalData(&pdf, textLeftMargin, "Датум почетка осигурања:", doc.InsuranceStartDate)

	putMedicalData(&pdf, textLeftMargin, "Опис:", doc.InsuranceDescription)

	section("Подаци о обвезнику плаћања доприноса")

	putMedicalData(&pdf, textLeftMargin, "Назив:", doc.TaxpayerName)

	putMedicalData(&pdf, textLeftMargin, "Седиште:", doc.TaxpayerResidence)

	putMedicalData(&pdf, textLeftMargin, "Регистарски број:", doc.TaxpayerNumber)

	putMedicalData(&pdf, textLeftMargin, "ПИБ/ЈМБГ:", doc.TaxpayerIDNumber)

	putMedicalData(&pdf, textLeftMargin, "Делатност:", doc.TaxpayerActivityCode)

	fileName = doc.formatFilename() + ".pdf"

	pdf.SetInfo(gopdf.PdfInfo{
		Title:        doc.GivenNameLatin + " " + doc.FamilyNameLatin,
		Author:       "Baš Čelik",
		Subject:      "Lična karta",
		CreationDate: time.Now(),
	})

	return pdf.GetBytesPdf(), fileName, nil
}

// BuildJson creates a JSON representation of the MedicalDocument.
func (doc *MedicalDocument) BuildJson() ([]byte, error) {
	return json.Marshal(doc)
}

// BuildExcel creates an Excel representation of the MedicalDocument.
func (doc *MedicalDocument) BuildExcel() ([]byte, string, error) {
	xlsx, err := CreateExcel(*doc)
	name := doc.formatFilename() + ".xlsx"
	return xlsx, name, err
}

func (doc *MedicalDocument) formatFilename() string {
	return strings.ToLower(doc.GivenNameLatin + "_" + doc.FamilyNameLatin)
}

// UpdateValidUntilDateFromRfzo fetches and updates the ValidUntil date from the RFZO web service.
func (doc *MedicalDocument) UpdateValidUntilDateFromRfzo() error {
	if len([]rune(doc.CardID)) != 11 {
		return ErrInvalidCardNo
	}

	if len([]rune(doc.InsurantNumber)) != 11 {
		return ErrInvalidInsuranceNo
	}

	resp, err := http.PostForm(rfzoServiceURL, url.Values{"zk": {doc.CardID}, "lbo": {doc.InsurantNumber}})
	if err != nil {
		return fmt.Errorf("posting: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	date, err := ParseValidUntilDateFromRfzoResponse(string(body))
	if err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	doc.ValidUntil = date

	return nil
}

// ParseValidUntilDateFromRfzoResponse extracts the ValidUntil date from RFZO service response.
func ParseValidUntilDateFromRfzoResponse(response string) (string, error) {
	regex, err := regexp.Compile(`оверена до: <strong>(\d+\.\d+\.\d+\.)</strong>`)
	if err != nil {
		return "", fmt.Errorf("compiling regex: %w", err)
	}

	matches := regex.FindStringSubmatch(response)
	if len(matches) < 2 {
		return "", ErrNoSubmatchFound
	}

	return matches[1], nil
}
