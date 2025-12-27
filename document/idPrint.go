package document

import (
	"fmt"
	"math"
	"time"

	"github.com/signintech/gopdf"
	"github.com/ubavic/bas-celik/v2/localization"
)

// IDPdfWriter handles writing ID document data to a PDF.
type IDPdfWriter struct {
	pdf            *gopdf.GoPdf
	leftMargin     float64
	rightMargin    float64
	textLeftMargin float64
	doc            *IDDocument
}

func (idw *IDPdfWriter) line(width float64) {
	if width > 0 {
		idw.pdf.SetLineWidth(width)
	}

	y := idw.pdf.GetY()
	idw.pdf.Line(idw.leftMargin, y, idw.rightMargin, y)
}

func (idw *IDPdfWriter) moveY(y float64) {
	idw.pdf.SetXY(idw.pdf.GetX(), idw.pdf.GetY()+y)
}

func (idw *IDPdfWriter) cell(s string) {
	err := idw.pdf.Cell(nil, s)
	if err != nil {
		panic(fmt.Errorf("putting text: %w", err))
	}
}

func (idw *IDPdfWriter) putData(label, data string) {
	y := idw.pdf.GetY()

	idw.pdf.SetX(idw.textLeftMargin)
	texts, err := idw.pdf.SplitTextWithWordWrap(label, 120)
	if err != nil && err != gopdf.ErrEmptyString {
		panic(err)
	}

	for i, text := range texts {
		idw.cell(text)
		if i < len(texts)-1 {
			idw.pdf.SetXY(idw.textLeftMargin, idw.pdf.GetY()+12)
		}
	}

	y1 := idw.pdf.GetY()

	idw.pdf.SetXY(idw.textLeftMargin+128, y)
	texts, err = idw.pdf.SplitTextWithWordWrap(data, 350)
	if err != nil && err != gopdf.ErrEmptyString {
		panic(err)
	}

	for i, text := range texts {
		idw.cell(text)
		if i < len(texts)-1 {
			idw.pdf.SetXY(idw.textLeftMargin+128, idw.pdf.GetY()+12)
		}
	}

	y2 := idw.pdf.GetY()

	idw.pdf.SetXY(idw.textLeftMargin, math.Max(y1, y2)+24.67)
}

func (idw *IDPdfWriter) printRegularID() {
	idw.pdf.SetLineType("solid")
	idw.pdf.SetY(59.041)
	idw.line(0.83)

	idw.pdf.SetXY(idw.textLeftMargin+1.0, 68.5)

	err := idw.pdf.SetCharSpacing(-0.2)
	if err != nil {
		panic(err)
	}
	idw.cell("ČITAČ ELEKTRONSKE LIČNE KARTE: ŠTAMPA PODATAKA")

	err = idw.pdf.SetCharSpacing(-0.1)
	if err != nil {
		panic(err)
	}

	idw.pdf.SetY(88)

	idw.line(0)

	imageY := 102.8
	imageHeight := 159.0

	err = idw.pdf.ImageFrom(idw.doc.Portrait, idw.leftMargin, imageY, &gopdf.Rect{W: 119.9, H: imageHeight})
	if err != nil {
		panic(err)
	}

	idw.pdf.SetLineWidth(0.48)
	idw.pdf.SetFillColor(255, 255, 255)
	err = idw.pdf.Rectangle(idw.leftMargin, imageY, 179, imageY+imageHeight, "D", 0, 0)
	if err != nil {
		panic(err)
	}

	idw.pdf.SetFillColor(0, 0, 0)

	idw.pdf.SetY(276)

	idw.line(1.08)
	idw.moveY(8)
	idw.pdf.SetX(idw.textLeftMargin)
	err = idw.pdf.SetFontSize(11.1)
	if err != nil {
		panic(err)
	}

	idw.cell("Podaci o građaninu")

	idw.moveY(16)
	idw.line(0)
	idw.moveY(9)

	idw.putData("Prezime:", idw.doc.Surname)
	idw.putData("Ime:", idw.doc.GivenName)
	idw.putData("Ime jednog roditelja:", idw.doc.ParentGivenName)
	idw.putData("Datum rođenja:", idw.doc.DateOfBirth)
	idw.putData("Mesto rođenja,\nopština i država:", idw.doc.GetFullPlaceOfBirth())
	addressLabel := "Prebivalište\ni adresa stana:"
	if idw.doc.AddressLabel == "prebivalište" {
		addressLabel = "Prebivalište:"
	}
	idw.putData(addressLabel, idw.doc.GetFullAddress(true))
	idw.putData("Datum promene adrese:", idw.doc.AddressDate)
	idw.putData("JMBG:", idw.doc.PersonalNumber)
	idw.putData("Pol:", idw.doc.Sex)

	idw.moveY(-8.67)
	idw.line(0)
	idw.moveY(9)
	idw.cell("Podaci o dokumentu")
	idw.moveY(16)

	idw.line(0)
	idw.moveY(9)
	idw.putData("Dokument izdaje:", idw.doc.IssuingAuthority)
	idw.putData("Broj dokumenta:", idw.doc.DocRegNo)
	idw.putData("Datum izdavanja:", idw.doc.IssuingDate)
	idw.putData("Važi do:", idw.doc.ExpiryDate)

	idw.moveY(-8.67)
	idw.line(0)
	idw.moveY(3)
	idw.line(0)
	idw.moveY(9)

	idw.cell("Datum štampe: " + time.Now().Format("02.01.2006."))

	idw.moveY(19)

	if idw.pdf.GetY() < 700 {
		idw.pdf.SetY(730.6)
	}

	idw.line(0.83)

	err = idw.pdf.SetFontSize(9)
	if err != nil {
		panic(err)
	}

	idw.moveY(10)
	idw.pdf.SetX(idw.leftMargin)

	idw.cell("1. U čipu lične karte, podaci o imenu i prezimenu imaoca lične karte ispisani su na nacionalnom pismu onako kako su")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("ispisani na samom obrascu lične karte, dok su ostali podaci ispisani latiničkim pismom.")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("2. Ako se ime lica sastoji od dve reči čija je ukupna dužina između 20 i 30 karaktera ili prezimena od dve reči čija je")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("ukupna dužina između 30 i 36 karaktera, u čipu lične karte izdate pre 18.08.2014. godine, druga reč u imenu ili prezimenu")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("skraćuje se na prva dva karaktera")

	idw.moveY(15.7)
	idw.line(0)
}

func (idw *IDPdfWriter) printForeignerID() {
	idw.pdf.SetLineType("solid")
	idw.pdf.SetY(59.041)
	idw.line(0.83)

	idw.pdf.SetXY(idw.textLeftMargin+1.0, 64.95)

	err := idw.pdf.SetCharSpacing(-0.2)
	if err != nil {
		panic(err)
	}
	idw.cell("ČITAČ ELEKTRONSKE LIČNE KARTE: ŠTAMPA PODATAKA")

	err = idw.pdf.SetCharSpacing(-0.1)
	if err != nil {
		panic(err)
	}

	idw.pdf.SetY(79.8)

	idw.line(0)

	imageY := 86.0
	imageHeight := 159.0

	err = idw.pdf.ImageFrom(idw.doc.Portrait, idw.leftMargin, imageY, &gopdf.Rect{W: 119.9, H: imageHeight})
	if err != nil {
		panic(err)
	}

	idw.pdf.SetLineWidth(0.48)
	idw.pdf.SetFillColor(255, 255, 255)
	err = idw.pdf.Rectangle(idw.leftMargin, imageY, 179, imageY+imageHeight, "D", 0, 0)
	if err != nil {
		panic(err)
	}

	idw.pdf.SetFillColor(0, 0, 0)

	idw.pdf.SetY(250)

	idw.line(1.08)
	idw.moveY(8)
	idw.pdf.SetX(idw.textLeftMargin)
	err = idw.pdf.SetFontSize(11.1)
	if err != nil {
		panic(err)
	}

	idw.cell("Podaci o strancu")

	idw.moveY(16)
	idw.line(0)
	idw.moveY(9)

	idw.putData("Prezime:", idw.doc.Surname)
	idw.putData("Ime:", idw.doc.GivenName)
	idw.putData("Državljanstvo:", idw.doc.NationalityFull)
	idw.putData("Datum rođenja:", idw.doc.DateOfBirth)
	idw.putData("Osnov boravka:", idw.doc.PurposeOfStay)
	addressLabel := "Prebivalište\ni adresa stana:"
	if idw.doc.AddressLabel == "prebivalište" {
		addressLabel = "Prebivalište:"
	}
	idw.putData(addressLabel, localization.JoinWithComma(idw.doc.State, idw.doc.GetFullAddress(true)))
	idw.putData("Datum promene adrese:", idw.doc.AddressDate)
	idw.putData("Evidencijski broj\nstranca:", idw.doc.PersonalNumber)
	idw.putData("Pol:", idw.doc.Sex)

	idw.moveY(-8.67)
	idw.line(0)
	idw.moveY(9)
	idw.cell("Podaci o dokumentu")
	idw.moveY(16)

	idw.line(0)
	idw.moveY(9)
	idw.putData("Dokument izdaje:", idw.doc.IssuingAuthority)
	idw.putData("Broj dokumenta:", idw.doc.DocRegNo)
	idw.putData("Datum izdavanja:", idw.doc.IssuingDate)
	idw.putData("Važi do:", idw.doc.ExpiryDate)

	idw.moveY(-8.67)
	idw.line(0)
	idw.moveY(3)
	idw.line(0)
	idw.moveY(9)

	idw.cell("Datum štampe: " + time.Now().Format("02.01.2006."))

	idw.moveY(19)

	idw.line(0.83)

	err = idw.pdf.SetFontSize(9)
	if err != nil {
		panic(err)
	}

	idw.moveY(4)

	idw.pdf.SetX(idw.leftMargin)

	idw.cell("1. U čipu lične karte za strance, podaci o imenu i prezimenu stranca ispisani su onako kako su ispisani na samom")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("obrascu lične karte za stranca latiničnim pismom.")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("2. Ako se ime ili prezime stranca sastoji od dve ili više reči čija dužina prelazi 30 karaktera za ime, odnosno 36")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("karaktera za prezime, u čip se upisuje puno ime i prezime stranca, a na obrascu lične karte za stranca se upisuje do")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("30 karaktera za ime, odnosno 36 karaktera za prezime.")

	idw.moveY(9.7)

	idw.line(0)
}

func (idw *IDPdfWriter) printResidencePermit() {
	idw.pdf.SetLineType("solid")
	idw.pdf.SetY(59.041)
	idw.line(0.83)

	idw.pdf.SetXY(idw.textLeftMargin+1.0, 64.95)

	err := idw.pdf.SetCharSpacing(-0.2)
	if err != nil {
		panic(err)
	}
	idw.cell("ČITAČ ELEKTRONSKE LIČNE KARTE: ŠTAMPA PODATAKA")

	err = idw.pdf.SetCharSpacing(-0.1)
	if err != nil {
		panic(err)
	}

	idw.pdf.SetY(79.8)

	idw.line(0)

	imageY := 86.0
	imageHeight := 159.0

	err = idw.pdf.ImageFrom(idw.doc.Portrait, idw.leftMargin, imageY, &gopdf.Rect{W: 119.9, H: imageHeight})
	if err != nil {
		panic(err)
	}

	idw.pdf.SetLineWidth(0.48)
	idw.pdf.SetFillColor(255, 255, 255)
	err = idw.pdf.Rectangle(idw.leftMargin, imageY, 179, imageY+imageHeight, "D", 0, 0)
	if err != nil {
		panic(err)
	}

	idw.pdf.SetFillColor(0, 0, 0)

	idw.pdf.SetY(250)

	idw.line(1.08)
	idw.moveY(8)
	idw.pdf.SetX(idw.textLeftMargin)
	err = idw.pdf.SetFontSize(11.1)
	if err != nil {
		panic(err)
	}

	idw.cell("Podaci o strancu")

	idw.moveY(16)
	idw.line(0)
	idw.moveY(9)

	idw.putData("Prezime:", idw.doc.Surname)
	idw.putData("Ime:", idw.doc.GivenName)
	idw.putData("Državljanstvo:", idw.doc.NationalityFull)
	idw.putData("Datum rođenja:", idw.doc.DateOfBirth)
	idw.putData("Mesto rođenja,\nopština i država:", idw.doc.GetFullPlaceOfBirth())
	addressLabel := "Prebivalište\ni adresa stana:"
	if idw.doc.AddressLabel == "prebivalište" {
		addressLabel = "Prebivalište:"
	}
	idw.putData(addressLabel, idw.doc.GetFullAddress(true))
	idw.putData("Datum promene adrese:", idw.doc.AddressDate)
	idw.putData("Evidencijski broj\nstranca:", idw.doc.PersonalNumber)
	idw.putData("Pol:", idw.doc.Sex)
	idw.putData("Osnov boravka:", idw.doc.PurposeOfStay)
	idw.putData("Napomena:", idw.doc.ENote)

	idw.moveY(-8.67)
	idw.line(0)
	idw.moveY(9)
	idw.cell("Podaci o dokumentu")
	idw.moveY(16)

	idw.line(0)
	idw.moveY(9)
	idw.putData("Naziv dokumenta:", idw.doc.DocumentName)
	idw.putData("Dokument izdaje:", idw.doc.IssuingAuthority)
	idw.putData("Broj dokumenta:", idw.doc.DocRegNo)
	idw.putData("Datum izdavanja:", idw.doc.IssuingDate)
	idw.putData("Važi do:", idw.doc.ExpiryDate)

	idw.moveY(-8.67)
	idw.line(0)
	idw.moveY(3)
	idw.line(0)
	idw.moveY(9)

	idw.cell("Datum štampe: " + time.Now().Format("02.01.2006."))

	idw.moveY(19)

	idw.line(0.83)

	err = idw.pdf.SetFontSize(9)
	if err != nil {
		panic(err)
	}

	idw.moveY(4)

	idw.pdf.SetX(idw.leftMargin)

	idw.cell("1. U čipu dozvole za privremeni boravak i rad, podaci o imenu i prezimenu imaoca dozvole ispisani su onako")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("kako su ispisani na samom obrascu dozvole za privremeni boravak latiničnim pismom.")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("2. Ako se ime ili prezime stranca sastoji od dve ili više reči čija dužina prelazi 30 karaktera za ime,")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("odnosno 36 karaktera za prezime u čip se upisuje puno ime stranca, a na obrascu dozvole za privremeni boravak")
	idw.pdf.SetX(idw.leftMargin)
	idw.moveY(9.7)
	idw.cell("se upisuje do 30 karaktera za ime, odnosno 36 karaktera za prezime.")

	idw.moveY(9.7)

	idw.line(0)
}
