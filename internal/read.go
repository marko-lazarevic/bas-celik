package internal

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/ebfe/scard"
	"github.com/ubavic/bas-celik/v2/card"
	"github.com/ubavic/bas-celik/v2/document"
)

// LaunchConfig contains configuration options for launching the application.
type LaunchConfig struct {
	PdfPath               string
	JSONPath              string
	ExcelPath             string
	Verbose               bool
	GetValidUntilFromRfzo bool
	Reader                uint
	EmbedDirectory        embed.FS
}

func checkFile(path string) error {
	if len(path) == 0 {
		return nil
	}

	_, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking file %s: %w", path, err)
	}

	return nil
}

func writeFileIfNotEmpty(path string, generate func() ([]byte, string, error)) error {
	if len(path) > 0 {
		pdf, _, err := generate()
		if err != nil {
			return fmt.Errorf("generating pdf: %w", err)
		}

		err = os.WriteFile(path, pdf, 0600)
		if err != nil {
			return fmt.Errorf("writing file %s: %w", path, err)
		}
	}
	return nil
}

func checkReaders(ctx *scard.Context, cfg LaunchConfig) ([]string, error) {
	readersNames, err := ctx.ListReaders()
	if err != nil {
		return nil, fmt.Errorf("listing readers: %w", err)
	}

	if len(readersNames) == 0 {
		return nil, fmt.Errorf("no reader found")
	}

	if cfg.Reader >= uint(len(readersNames)) {
		return nil, fmt.Errorf("only %d readers found", len(readersNames))
	}

	return readersNames, nil
}

func updateMedical(doc *document.MedicalDocument, cfg LaunchConfig) error {
	if cfg.GetValidUntilFromRfzo {
		err := doc.UpdateValidUntilDateFromRfzo()
		if err != nil {
			return fmt.Errorf("updating `ValidUntil` date: %w", err)
		}
	}
	return nil
}

func checkFiles(cfg LaunchConfig) error {
	if err := checkFile(cfg.PdfPath); err != nil {
		return err
	}

	if err := checkFile(cfg.JSONPath); err != nil {
		return err
	}

	if err := checkFile(cfg.ExcelPath); err != nil {
		return err
	}

	return nil;
}

func detectCardAndGetDocument(sCard *scard.Card) (document.Document, error) {
	cardDoc, err := card.DetectCardDocument(sCard)
	if err != nil {
		return nil, fmt.Errorf("detecting card type: %w", err)
	}

	err = cardDoc.InitCard()
	if err != nil {
		return nil, fmt.Errorf("initializing card: %w", err)
	}

	err = cardDoc.ReadCard()
	if err != nil {
		return nil, fmt.Errorf("reading card: %w", err)
	}

	doc, err := cardDoc.GetDocument()
	if err != nil {
		return nil,	 fmt.Errorf("getting document: %w", err)
	}
	return doc, nil
}

func writeFilesIfNotEmpty(cfg LaunchConfig, doc document.Document) error {
	if err := writeFileIfNotEmpty(cfg.PdfPath, doc.BuildPdf); err != nil {
    	return fmt.Errorf("pdf: %w", err)
	}

	if err := writeFileIfNotEmpty(cfg.JSONPath, func() ([]byte, string, error) {json, err := doc.BuildJson(); return json, "", err}); err != nil {
    	return fmt.Errorf("pdf: %w", err)
	}

	if err := writeFileIfNotEmpty(cfg.ExcelPath, doc.BuildExcel); err != nil {
		return fmt.Errorf("excel: %w", err)
	}
	return nil

}

func readAndSave(cfg LaunchConfig) error {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return fmt.Errorf("establishing context: %w", err)
	}

	defer ctx.Release()

	// Check output files
	if err := checkFiles(cfg); err != nil {
		return err
	}

	readersNames, err := checkReaders(ctx, cfg)
	if err != nil {
		return err
	}

	sCard, err := ctx.Connect(readersNames[cfg.Reader], scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return fmt.Errorf("connecting reader %s: %w", readersNames[cfg.Reader], err)
	}

	defer sCard.Disconnect(scard.LeaveCard)

	doc, err := detectCardAndGetDocument(sCard)
	if err != nil {
		return err
	}

	switch doc := doc.(type) {
	case *document.MedicalDocument:
		if err := updateMedical(doc, cfg); err != nil {
			return err
		}
	}

	if err := writeFilesIfNotEmpty(cfg, doc); err != nil {
    	return err
	}

	return nil
}
