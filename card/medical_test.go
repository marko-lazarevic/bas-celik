package card

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/mock"
	"github.com/ubavic/bas-celik/v2/document"
	testhelpers "github.com/ubavic/bas-celik/v2/test_helpers"
)

func Test_descramble(t *testing.T) {
	// Some raw bytes from different cards
	testCases := []struct {
		data, expectedData string
	}{
		{
			"200435043f04430431043b04380447043a043804200044043e043d04340420003704300420003704340440043004320441044204320435043d043e0420003e044104380433044304400430045a043504",
			"Републички фонд за здравствено осигурање",
		},
		{
			"210440043104380458043004",
			"Србија",
		},
		{
			"110415041e041304200410041404",
			"БЕОГРАД",
		},
		{
			"170430043f043e0441043b0435043d0438042000430420003f044004380432044004350434043d043e043c04200034044004430448044204320443042c00200034044004430433043e043c0420003f044004300432043d043e043c0420003b043804460443042c0020003a043e04340420003f0440043504340443043704350442043d0438043a0430042c00200046043804320438043b043d04300420003b0438044604300420003d043004200041043b04430436043104380420004304200032043e045804410446043804",
			"Запослени у привредном друштву, другом правном лицу, код предузетника, цивилна лица на служби у војсци",
		},
		{
			"110443045f04350442042000200435043f04430431043b0438043a0435042000210440043104380458043504",
			"Буџет Републике Србије",
		},
	}

	for i, testCase := range testCases {
		t.Run(
			fmt.Sprintf("Case %d", i),
			func(t *testing.T) {
				if i < 0 {
					t.Errorf("Invalid negative test case index %d", i)
				}
				idx := uint(i)
				decoded, err := hex.DecodeString(testCase.data)
				if err != nil {
					t.Errorf("Unexpected error %v", err)
				}

				fields := make(map[uint][]byte, 0)
				fields[idx] = decoded

				descramble(fields, idx)
				if !slices.Equal(fields[idx], []byte(testCase.expectedData)) {
					t.Errorf("Got %s, but expected %s", string(fields[idx]), testCase.expectedData)
				}
			},
		)
	}

	t.Run(
		"Empty case",
		func(t *testing.T) {
			fields := make(map[uint][]byte, 0)
			descramble(fields, 1)
			if !slices.Equal(fields[1], []byte{}) {
				t.Errorf("Got %v, but expected empty slice", fields[1])
			}
		},
	)
}

func TestMedicalInitCard(t *testing.T) {
	s1 := []byte{0xF3, 0x81, 0x00, 0x00, 0x02, 0x53, 0x45, 0x52, 0x56, 0x53, 0x5A, 0x4B, 0x01}
	expectedAPDU := buildAPDU(0x00, 0xA4, 0x04, 0x00, s1, 0)

	tests := []struct {
		name      string
		response  []byte
		respErr   error
		wantErr   bool
		errSubstr string
	}{
		{name: "success", response: []byte{0x90, 0x00}},
		{name: "transmit error", respErr: errors.New("tx fail"), wantErr: true, errSubstr: "tx fail"},
		{name: "bad status", response: []byte{0x6A, 0x82}, wantErr: true, errSubstr: "response"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			cm.On("Transmit", mock.Anything).Return(tt.response, tt.respErr).Once()

			card := &MedicalCard{smartCard: cm}

			err := card.InitCard()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.errSubstr, err)
				}
			} else if err != nil {
				t.Fatalf("InitCard() error = %v", err)
			}

			if len(cm.Calls) == 0 {
				t.Fatalf("Transmit was not called")
			}
			apdu := cm.Calls[0].Arguments.Get(0).([]byte)
			if !bytes.Equal(apdu, expectedAPDU) {
				t.Fatalf("InitCard APDU mismatch: %x", apdu)
			}

			cm.AssertExpectations(t)
		})
	}
}

func TestMedicalReadCard(t *testing.T) {
	okSelect := func() []byte { return []byte{0x90, 0x00} }
	okHeader := func(body []byte) []byte {
		h := []byte{0x00, 0x00, byte(len(body)), 0x00}
		return append(h, 0x90, 0x00)
	}
	okBody := func(body []byte) []byte { return append(body, 0x90, 0x00) }

	makeSuccess := func(cm *testhelpers.CardMock, body []byte) {
		cm.On("Transmit", mock.Anything).Return(okSelect(), nil).Once()
		cm.On("Transmit", mock.Anything).Return(okHeader(body), nil).Once()
		cm.On("Transmit", mock.Anything).Return(okBody(body), nil).Once()
	}

	tests := []struct {
		name      string
		setup     func(cm *testhelpers.CardMock)
		wantErr   bool
		errSubstr string
		assertOK  func(t *testing.T, card *MedicalCard)
	}{
		{
			name: "success",
			setup: func(cm *testhelpers.CardMock) {
				makeSuccess(cm, []byte{0x11})
				makeSuccess(cm, []byte{0x22})
				makeSuccess(cm, []byte{0x33})
				makeSuccess(cm, []byte{0x44})
			},
			assertOK: func(t *testing.T, card *MedicalCard) {
				if !bytes.Equal(card.medicalDocumentFile, []byte{0x11}) {
					t.Fatalf("medicalDocumentFile = %x", card.medicalDocumentFile)
				}
				if !bytes.Equal(card.fixedPersonalFile, []byte{0x22}) {
					t.Fatalf("fixedPersonalFile = %x", card.fixedPersonalFile)
				}
				if !bytes.Equal(card.variablePersonalFile, []byte{0x33}) {
					t.Fatalf("variablePersonalFile = %x", card.variablePersonalFile)
				}
				if !bytes.Equal(card.variableAdminFile, []byte{0x44}) {
					t.Fatalf("variableAdminFile = %x", card.variableAdminFile)
				}
			},
		},
		{
			name: "document error",
			setup: func(cm *testhelpers.CardMock) {
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail")).Once()
			},
			wantErr:   true,
			errSubstr: "document file",
		},
		{
			name: "fixed personal error",
			setup: func(cm *testhelpers.CardMock) {
				makeSuccess(cm, []byte{0x11})
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail"))
			},
			wantErr:   true,
			errSubstr: "fixed personal file",
		},
		{
			name: "variable personal error",
			setup: func(cm *testhelpers.CardMock) {
				makeSuccess(cm, []byte{0x11})
				makeSuccess(cm, []byte{0x22})
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail"))
			},
			wantErr:   true,
			errSubstr: "variable personal file",
		},
		{
			name: "variable admin error",
			setup: func(cm *testhelpers.CardMock) {
				makeSuccess(cm, []byte{0x11})
				makeSuccess(cm, []byte{0x22})
				makeSuccess(cm, []byte{0x33})
				cm.On("Transmit", mock.Anything).Return(nil, errors.New("select fail"))
			},
			wantErr:   true,
			errSubstr: "variable administrative file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			if tt.setup != nil {
				tt.setup(cm)
			}

			card := &MedicalCard{smartCard: cm}
			err := card.ReadCard()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.errSubstr, err)
				}
			} else if err != nil {
				t.Fatalf("ReadCard() error = %v", err)
			}

			if tt.assertOK != nil {
				tt.assertOK(t, card)
			}
			cm.AssertExpectations(t)
		})
	}
}

func TestMedicalGetDocument(t *testing.T) {
	utf16le := func(s string) []byte {
		runes := utf16.Encode([]rune(s))
		out := make([]byte, 2+len(runes)*2)
		out[0], out[1] = 0xFF, 0xFE
		for i, r := range runes {
			binary.LittleEndian.PutUint16(out[2+i*2:], r)
		}
		return out
	}
	makeTLV := func(tag uint16, val []byte) []byte {
		buf := make([]byte, 4+len(val))
		binary.LittleEndian.PutUint16(buf[0:], tag)
		binary.LittleEndian.PutUint16(buf[2:], uint16(len(val)))
		copy(buf[4:], val)
		return buf
	}

	buildFiles := func(gender string) (docFile, fixedFile, varPersFile, varAdmFile []byte) {
		docFile = append(docFile, makeTLV(1553, utf16le("Insurer"))...)
		docFile = append(docFile, makeTLV(1554, []byte("ID123"))...)
		docFile = append(docFile, makeTLV(1555, []byte("CARD"))...)
		docFile = append(docFile, makeTLV(1557, []byte("01012024"))...)
		docFile = append(docFile, makeTLV(1558, []byte("02012025"))...)
		docFile = append(docFile, makeTLV(1560, []byte("SR"))...)

		fixedFile = append(fixedFile, makeTLV(1570, utf16le("Prezime"))...)
		fixedFile = append(fixedFile, makeTLV(1571, utf16le("PrezimeLat"))...)
		fixedFile = append(fixedFile, makeTLV(1572, utf16le("Ime"))...)
		fixedFile = append(fixedFile, makeTLV(1573, utf16le("ImeLat"))...)
		fixedFile = append(fixedFile, makeTLV(1574, []byte("03031990"))...)
		fixedFile = append(fixedFile, makeTLV(1569, []byte("LBO123"))...)

		varPersFile = append(varPersFile, makeTLV(1586, []byte("04042030"))...)
		varPersFile = append(varPersFile, makeTLV(1587, []byte{0x31})...)

		varAdmFile = append(varAdmFile, makeTLV(1601, utf16le("Otac"))...)
		varAdmFile = append(varAdmFile, makeTLV(1602, utf16le("OtacLat"))...)
		varAdmFile = append(varAdmFile, makeTLV(1603, []byte(gender))...)
		varAdmFile = append(varAdmFile, makeTLV(1604, []byte("JMBG"))...)
		varAdmFile = append(varAdmFile, makeTLV(1605, utf16le("Street"))...)
		varAdmFile = append(varAdmFile, makeTLV(1607, utf16le("Municip"))...)
		varAdmFile = append(varAdmFile, makeTLV(1608, utf16le("Place"))...)
		varAdmFile = append(varAdmFile, makeTLV(1610, utf16le("No"))...)
		varAdmFile = append(varAdmFile, makeTLV(1612, utf16le("Apt"))...)
		varAdmFile = append(varAdmFile, makeTLV(1614, []byte("Basis"))...)
		varAdmFile = append(varAdmFile, makeTLV(1615, utf16le("Descr"))...)
		varAdmFile = append(varAdmFile, makeTLV(1616, utf16le("Rel"))...)
		varAdmFile = append(varAdmFile, makeTLV(1617, []byte{0x31})...)
		varAdmFile = append(varAdmFile, makeTLV(1618, []byte("CarrID"))...)
		varAdmFile = append(varAdmFile, makeTLV(1619, []byte("CarrLBO"))...)
		varAdmFile = append(varAdmFile, makeTLV(1620, utf16le("CarrFam"))...)
		varAdmFile = append(varAdmFile, makeTLV(1621, utf16le("CarrFamLat"))...)
		varAdmFile = append(varAdmFile, makeTLV(1622, utf16le("CarrGiven"))...)
		varAdmFile = append(varAdmFile, makeTLV(1623, utf16le("CarrGivenLat"))...)
		varAdmFile = append(varAdmFile, makeTLV(1624, []byte("05052015"))...)
		varAdmFile = append(varAdmFile, makeTLV(1626, utf16le("Country"))...)
		varAdmFile = append(varAdmFile, makeTLV(1630, utf16le("TaxName"))...)
		varAdmFile = append(varAdmFile, makeTLV(1631, utf16le("TaxRes"))...)
		varAdmFile = append(varAdmFile, makeTLV(1632, []byte("TaxID"))...)
		varAdmFile = append(varAdmFile, makeTLV(1634, []byte("TaxAct"))...)

		return
	}

	successDoc, successFixed, successVar, successAdmMale := buildFiles("01")
	_, _, _, successAdmFemale := buildFiles("02")

	tests := []struct {
		name      string
		files     func() *MedicalCard
		wantErr   bool
		errSubstr string
		assertOK  func(t *testing.T, doc *document.MedicalDocument)
	}{
		{
			name: "success male",
			files: func() *MedicalCard {
				return &MedicalCard{
					medicalDocumentFile:  successDoc,
					fixedPersonalFile:    successFixed,
					variablePersonalFile: successVar,
					variableAdminFile:    successAdmMale,
				}
			},
			assertOK: func(t *testing.T, doc *document.MedicalDocument) {
				if doc.InsurerName != "Insurer" || doc.InsurerID != "ID123" {
					t.Fatalf("unexpected insurer data: %+v", doc)
				}
				if doc.DateOfIssue != "01.01.2024." || doc.DateOfExpiry != "02.01.2025." {
					t.Fatalf("date formatting failed: %s %s", doc.DateOfIssue, doc.DateOfExpiry)
				}
				if doc.Gender != "Mушко" {
					t.Fatalf("expected male gender, got %s", doc.Gender)
				}
				if !doc.PermanentlyValid || !doc.CarrierFamilyMember {
					t.Fatalf("boolean fields not parsed")
				}
				if doc.InsuranceStartDate != "05.05.2015." || doc.ValidUntil != "04.04.2030." {
					t.Fatalf("date formatting in admin/var files failed")
				}
			},
		},
		{
			name: "success female",
			files: func() *MedicalCard {
				return &MedicalCard{
					medicalDocumentFile:  successDoc,
					fixedPersonalFile:    successFixed,
					variablePersonalFile: successVar,
					variableAdminFile:    successAdmFemale,
				}
			},
			assertOK: func(t *testing.T, doc *document.MedicalDocument) {
				if doc.Gender != "Женско" {
					t.Fatalf("expected female gender, got %s", doc.Gender)
				}
			},
		},
		{
			name: "document parse error",
			files: func() *MedicalCard {
				return &MedicalCard{medicalDocumentFile: []byte{}}
			},
			wantErr:   true,
			errSubstr: "document file",
		},
		{
			name: "fixed parse error",
			files: func() *MedicalCard {
				return &MedicalCard{medicalDocumentFile: successDoc, fixedPersonalFile: []byte{}}
			},
			wantErr:   true,
			errSubstr: "fixed personal file",
		},
		{
			name: "variable personal parse error",
			files: func() *MedicalCard {
				return &MedicalCard{medicalDocumentFile: successDoc, fixedPersonalFile: successFixed, variablePersonalFile: []byte{}}
			},
			wantErr:   true,
			errSubstr: "variable personal file",
		},
		{
			name: "variable admin parse error",
			files: func() *MedicalCard {
				return &MedicalCard{medicalDocumentFile: successDoc, fixedPersonalFile: successFixed, variablePersonalFile: successVar, variableAdminFile: []byte{}}
			},
			wantErr:   true,
			errSubstr: "variable administrative file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.files()
			d, err := card.GetDocument()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.errSubstr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetDocument() error = %v", err)
			}
			dc, ok := d.(*document.MedicalDocument)
			if !ok {
				t.Fatalf("unexpected document type %T", d)
			}
			if tt.assertOK != nil {
				tt.assertOK(t, dc)
			}
		})
	}
}

func TestMedicalAtr(t *testing.T) {
	atr := Atr([]byte{0x01, 0x02})
	card := &MedicalCard{atr: atr}
	if got := card.Atr(); !bytes.Equal(got, atr) {
		t.Fatalf("Atr() = %x, want %x", got, atr)
	}
}

func TestMedicalTestSuccess(t *testing.T) {
	utf16le := func(s string) []byte {
		runes := utf16.Encode([]rune(s))
		out := make([]byte, 2+len(runes)*2)
		out[0], out[1] = 0xFF, 0xFE
		for i, r := range runes {
			binary.LittleEndian.PutUint16(out[2+i*2:], r)
		}
		return out
	}
	makeTLV := func(tag uint16, val []byte) []byte {
		buf := make([]byte, 4+len(val))
		binary.LittleEndian.PutUint16(buf[0:], tag)
		binary.LittleEndian.PutUint16(buf[2:], uint16(len(val)))
		copy(buf[4:], val)
		return buf
	}

	payload := makeTLV(1553, utf16le("Републички фонд за здравствено осигурање"))
	header := append([]byte{0x00, 0x00, byte(len(payload)), byte(len(payload) >> 8)}, 0x90, 0x00)
	body := append(append([]byte{}, payload...), 0x90, 0x00)

	cm := &testhelpers.CardMock{}
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once() // app select in Test()
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once() // selectFile in ReadFile
	cm.On("Transmit", mock.Anything).Return(header, nil).Once()             // header read
	cm.On("Transmit", mock.Anything).Return(body, nil).Once()               // body read

	card := &MedicalCard{smartCard: cm}
	if ok := card.Test(); !ok {
		t.Fatalf("expected Test() to return true")
	}

	cm.AssertExpectations(t)
}
