package card

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ebfe/scard"
	"github.com/stretchr/testify/mock"
	doc "github.com/ubavic/bas-celik/v2/document"
	testhelpers "github.com/ubavic/bas-celik/v2/test_helpers"
)

func TestGemaltoInitCard(t *testing.T) {
	type call struct {
		rsp []byte
		err error
	}

	tests := []struct {
		name   string
		calls  []call
		errSub string
	}{
		{name: "first applet", calls: []call{{rsp: []byte{0x90, 0x00}}}},
		{name: "second applet", calls: []call{{rsp: []byte{0x6A, 0x82}}, {rsp: []byte{0x90, 0x00}}}},
		{name: "fallback applet", calls: []call{{rsp: []byte{0x6A, 0x82}}, {rsp: []byte{0x6A, 0x82}}, {rsp: []byte{0x90, 0x00}}}},
		{name: "all fail", calls: []call{{rsp: []byte{0x6A, 0x82}}, {rsp: []byte{0x6A, 0x82}}, {rsp: []byte{0x6A, 0x82}}}, errSub: "unknown"},
		{name: "first transmit error", calls: []call{{err: errors.New("tx1")}}, errSub: "initializing ID card"},
		{name: "second transmit error", calls: []call{{rsp: []byte{0x6A, 0x82}}, {err: errors.New("tx2")}}, errSub: "initializing IF card"},
		{name: "third transmit error", calls: []call{{rsp: []byte{0x6A, 0x82}}, {rsp: []byte{0x6A, 0x82}}, {err: errors.New("tx3")}}, errSub: "initializing RP card"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			for _, c := range tt.calls {
				cm.On("Transmit", mock.Anything).Return(c.rsp, c.err).Once()
			}

			g := Gemalto{smartCard: cm}

			err := g.InitCard()
			if tt.errSub == "" {
				if err != nil {
					t.Fatalf("InitCard() unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("InitCard() expected error containing %q, got %v", tt.errSub, err)
				}
			}

			cm.AssertExpectations(t)
		})
	}
}

func TestGemaltoReadCard(t *testing.T) {
	type call struct {
		rsp []byte
		err error
	}

	appendFileCalls := func(payload []byte, calls *[]call) {
		header := []byte{0x00, 0x00, byte(len(payload)), 0x00, 0x90, 0x00}
		*calls = append(*calls,
			call{rsp: []byte{0x90, 0x00}}, // select ok
			call{rsp: header},             // header
			call{rsp: append(payload, 0x90, 0x00)},
		)
	}

	docPayload := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	personalPayload := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	resPayload := []byte{0x01, 0x02, 0x03, 0x04}
	photoPayload := []byte{0x00, 0x01, 0x02, 0x03, 0x99}

	tests := []struct {
		name      string
		calls     []call
		errSub    string
		wantPhoto []byte
	}{
		{
			name: "success",
			calls: func() []call {
				var calls []call
				appendFileCalls(docPayload, &calls)
				appendFileCalls(personalPayload, &calls)
				appendFileCalls(resPayload, &calls)
				appendFileCalls(photoPayload, &calls)
				return calls
			}(),
			wantPhoto: []byte{0x99},
		},
		{
			name:   "document read error",
			calls:  []call{{err: errors.New("doc fail")}},
			errSub: "reading document file",
		},
		{
			name: "personal read error",
			calls: func() []call {
				var calls []call
				appendFileCalls(docPayload, &calls)
				calls = append(calls, call{err: errors.New("personal fail")})
				return calls
			}(),
			errSub: "reading personal file",
		},
		{
			name: "residence read error",
			calls: func() []call {
				var calls []call
				appendFileCalls(docPayload, &calls)
				appendFileCalls(personalPayload, &calls)
				calls = append(calls, call{err: errors.New("res fail")})
				return calls
			}(),
			errSub: "reading residence file",
		},
		{
			name: "photo read error",
			calls: func() []call {
				var calls []call
				appendFileCalls(docPayload, &calls)
				appendFileCalls(personalPayload, &calls)
				appendFileCalls(resPayload, &calls)
				calls = append(calls, call{err: errors.New("photo fail")})
				return calls
			}(),
			errSub: "reading photo file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			for _, c := range tt.calls {
				cm.On("Transmit", mock.Anything).Return(c.rsp, c.err).Once()
			}

			g := Gemalto{smartCard: cm}
			err := g.ReadCard()

			if tt.errSub == "" {
				if err != nil {
					t.Fatalf("ReadCard() unexpected error: %v", err)
				}
				if tt.wantPhoto != nil && !bytes.Equal(g.photoFile, tt.wantPhoto) {
					t.Fatalf("photoFile = %x, want %x", g.photoFile, tt.wantPhoto)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("ReadCard() expected error containing %q, got %v", tt.errSub, err)
				}
			}

			cm.AssertExpectations(t)
		})
	}
}

func makeTLV(tag uint16, value []byte) []byte {
	out := make([]byte, 0, 4+len(value))
	out = append(out, byte(tag), byte(tag>>8))
	out = append(out, byte(len(value)), byte(len(value)>>8))
	out = append(out, value...)
	return out
}

func TestGemaltoGetDocument(t *testing.T) {
	tests := []struct {
		name         string
		docData      []byte
		personalData []byte
		resData      []byte
		errSub       string
	}{

		{
			name:         "document parse error",
			docData:      []byte{},
			personalData: makeTLV(1558, []byte("pn")),
			resData:      makeTLV(1568, []byte("rs")),
			errSub:       "parsing document file",
		},
		{
			name:         "personal parse error",
			docData:      makeTLV(1546, []byte("doc")),
			personalData: []byte{},
			resData:      makeTLV(1568, []byte("rs")),
			errSub:       "parsing personal file",
		},
		{
			name:         "residence parse error",
			docData:      makeTLV(1546, []byte("doc")),
			personalData: makeTLV(1558, []byte("pn")),
			resData:      []byte{},
			errSub:       "parsing residence file",
		},
		{
			name:         "photo decode error",
			docData:      makeTLV(1546, []byte("doc")),
			personalData: makeTLV(1558, []byte("pn")),
			resData:      makeTLV(1568, []byte("rs")),
			errSub:       "parsing photo file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Gemalto{
				documentFile:  tt.docData,
				personalFile:  tt.personalData,
				residenceFile: tt.resData,
			}

			d, err := g.GetDocument()
			if tt.errSub == "" {
				if err != nil {
					t.Fatalf("GetDocument() unexpected error: %v", err)
				}
				id, ok := d.(*doc.IDDocument)
				if !ok || id == nil {
					t.Fatalf("GetDocument() returned %T, want *IDDocument", d)
				}
				if id.Portrait == nil {
					t.Fatalf("GetDocument() missing portrait image")
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("GetDocument() expected error containing %q, got %v", tt.errSub, err)
				}
				if d != nil {
					t.Fatalf("GetDocument() expected nil document on error, got %T", d)
				}
			}
		})
	}
}

func TestGemaltoAtr(t *testing.T) {
	atr := []byte{0xA}
	g := Gemalto{atr: atr}

	res := g.Atr()

	if !bytes.Equal(res, atr) {
		t.Fatalf("Atr() = %x, want %x", res, atr)
	}
}

func TestGemaltoReadFile(t *testing.T) {
	type call struct {
		rsp []byte
		err error
	}

	tests := []struct {
		name   string
		calls  []call
		errSub string
		want   []byte
	}{
		{
			name: "success",
			calls: []call{
				{rsp: []byte{0x90, 0x00}},
				{rsp: []byte{0x00, 0x00, 0x04, 0x00, 0x90, 0x00}},
				{rsp: []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x90, 0x00}},
			},
			want: []byte{0xDE, 0xAD, 0xBE, 0xEF},
		},
		{
			name:   "select transmit error",
			calls:  []call{{err: errors.New("select fail")}},
			errSub: "selecting file",
		},
		{
			name:   "select bad status",
			calls:  []call{{rsp: []byte{0x6A, 0x82}}},
			errSub: "selecting file: response",
		},
		{
			name: "header read error",
			calls: []call{
				{rsp: []byte{0x90, 0x00}},
				{err: errors.New("hdr fail")},
			},
			errSub: "reading file header",
		},
		{
			name: "header too short",
			calls: []call{
				{rsp: []byte{0x90, 0x00}},
				{rsp: []byte{0x01, 0x02, 0x90, 0x00}},
			},
			errSub: "file too short",
		},
		{
			name: "body read error",
			calls: []call{
				{rsp: []byte{0x90, 0x00}},
				{rsp: []byte{0x00, 0x00, 0x04, 0x00, 0x90, 0x00}},
				{err: errors.New("body fail")},
			},
			errSub: "reading file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			for _, c := range tt.calls {
				cm.On("Transmit", mock.Anything).Return(c.rsp, c.err).Once()
			}

			g := Gemalto{smartCard: cm}
			data, err := g.ReadFile([]byte{0x01, 0x02})

			if tt.errSub == "" {
				if err != nil {
					t.Fatalf("ReadFile() unexpected error: %v", err)
				}
				if !bytes.Equal(data, tt.want) {
					t.Fatalf("ReadFile() = %x, want %x", data, tt.want)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("ReadFile() expected error containing %q, got %v", tt.errSub, err)
				}
			}

			cm.AssertExpectations(t)
		})
	}
}

func TestGemaltoReadCertificateFile(t *testing.T) {
	type call struct {
		rsp []byte
		err error
	}

	tests := []struct {
		name   string
		calls  []call
		errSub string
		want   []byte
	}{
		{
			name: "success",
			calls: []call{
				{rsp: []byte{0x90, 0x00}},                                     // select ok
				{rsp: []byte{0x06, 0x00, 0x90, 0x00}},                         // header length=6 (total read=8)
				{rsp: []byte{0xBA, 0xDC, 0x0F, 0xFE, 0xCA, 0xFE, 0x90, 0x00}}, // body part 1 (6 bytes)
				{rsp: []byte{0xFA, 0xCE, 0x90, 0x00}},                         // body part 2 (2 bytes)
			},
			want: []byte{0xBA, 0xDC, 0x0F, 0xFE, 0xCA, 0xFE, 0xFA, 0xCE},
		},
		{
			name:   "select transmit error",
			calls:  []call{{err: errors.New("select fail")}},
			errSub: "selecting file",
		},
		{
			name:   "select bad status",
			calls:  []call{{rsp: []byte{0x6A, 0x82}}},
			errSub: "selecting file: response",
		},
		{
			name: "header read error",
			calls: []call{
				{rsp: []byte{0x90, 0x00}},
				{err: errors.New("hdr fail")},
			},
			errSub: "reading file header",
		},
		{
			name: "body read error",
			calls: []call{
				{rsp: []byte{0x90, 0x00}},
				{rsp: []byte{0x02, 0x00, 0x90, 0x00}},
				{err: errors.New("body fail")},
			},
			errSub: "reading file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &testhelpers.CardMock{}
			for _, c := range tt.calls {
				cm.On("Transmit", mock.Anything).Return(c.rsp, c.err).Once()
			}

			g := Gemalto{smartCard: cm}
			data, err := g.readCertificateFile([]byte{0x71, 0x02})

			if tt.errSub == "" {
				if err != nil {
					t.Fatalf("readCertificateFile() unexpected error: %v", err)
				}
				if !bytes.Equal(data, tt.want) {
					t.Fatalf("readCertificateFile() = %x, want %x", data, tt.want)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("readCertificateFile() expected error containing %q, got %v", tt.errSub, err)
				}
			}

			cm.AssertExpectations(t)
		})
	}
}

func TestGemaltoInitCrypto(t *testing.T) {
	cm := &testhelpers.CardMock{}
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()

	g := Gemalto{smartCard: cm}

	if err := g.InitCrypto(); err != nil {
		t.Fatalf("InitCrypto() unexpected error: %v", err)
	}

	cm.AssertExpectations(t)
}

func TestGemaltoChangePinSuccess(t *testing.T) {
	cm := &testhelpers.CardMock{}
	cm.On("BeginTransaction").Return(nil).Once()
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()
	cm.On("EndTransaction", scard.LeaveCard).Return(nil).Once()

	g := Gemalto{smartCard: cm}

	tries, err := g.ChangePin("5678", "1234")
	if err != nil {
		t.Fatalf("ChangePin() unexpected error: %v", err)
	}
	if tries != -1 {
		t.Fatalf("ChangePin() tries = %d, want -1", tries)
	}

	cm.AssertExpectations(t)
}

func TestGemaltoChangePinInvalidNew(t *testing.T) {
	cm := &testhelpers.CardMock{}
	cm.On("BeginTransaction").Return(nil).Once()
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()

	g := Gemalto{smartCard: cm}

	if _, err := g.ChangePin("12", "1234"); err == nil || !strings.Contains(err.Error(), "new pin") {
		t.Fatalf("ChangePin() expected validation error, got %v", err)
	}

	cm.AssertExpectations(t)
}

func TestGemaltoReadSignatures(t *testing.T) {
	cm := &testhelpers.CardMock{}
	// first signature
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()                                     // select
	cm.On("Transmit", mock.Anything).Return([]byte{0x00, 0x00, 0x06, 0x00, 0x90, 0x00}, nil).Once()             // header len=6
	cm.On("Transmit", mock.Anything).Return([]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x90, 0x00}, nil).Once() // body
	// second signature
	cm.On("Transmit", mock.Anything).Return([]byte{0x90, 0x00}, nil).Once()                                     // select
	cm.On("Transmit", mock.Anything).Return([]byte{0x00, 0x00, 0x06, 0x00, 0x90, 0x00}, nil).Once()             // header len=6
	cm.On("Transmit", mock.Anything).Return([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x90, 0x00}, nil).Once() // body

	card := Gemalto{smartCard: cm}

	if err := card.ReadSignatures(); err != nil {
		t.Fatalf("ReadSignatures() unexpected error: %v", err)
	}

	cm.AssertExpectations(t)
}

func TestGemaltoLoadCertificates_Existing(t *testing.T) {
	cm := &testhelpers.CardMock{}
	existing := &x509.Certificate{Raw: []byte{0x01, 0x02}}

	g := Gemalto{smartCard: cm, certificates: []*x509.Certificate{existing}}

	if err := g.LoadCertificates(); err != nil {
		t.Fatalf("LoadCertificates() unexpected error: %v", err)
	}

	if len(g.certificates) != 1 || g.certificates[0] != existing {
		t.Fatalf("LoadCertificates() mutated preloaded certificates")
	}

	cm.AssertExpectations(t)
}

func TestGemaltoGetCertificates(t *testing.T) {
	// build a minimal valid certificate
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("CreateCertificate() error: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("ParseCertificate() error: %v", err)
	}

	g := Gemalto{certificates: []*x509.Certificate{parsed, nil}}

	certs := g.GetCertificates()
	if len(certs) != 1 {
		t.Fatalf("GetCertificates() len=%d, want 1", len(certs))
	}
	if !bytes.Equal(certs[0].Raw, parsed.Raw) {
		t.Fatalf("GetCertificates() raw mismatch")
	}
}
