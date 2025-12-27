package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ubavic/bas-celik/v2/document"
	"github.com/ubavic/bas-celik/v2/internal/gui/reader"
	"github.com/ubavic/bas-celik/v2/internal/gui/translation"
	"github.com/ubavic/bas-celik/v2/internal/gui/widgets"
	"github.com/ubavic/bas-celik/v2/internal/logger"
)

func startCardReaderUI() {
	widgets.SetClipboard(copyToClipboard)

	spacer := widgets.NewSpacer()

	poller, pollerErr := reader.NewPoller(state.toolbar, connectToCard)

	rows := container.New(layout.NewVBoxLayout(), state.toolbar, spacer, state.startPage, state.documentUIMainContainer, state.cryptoUIContainer)
	columns := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), rows, layout.NewSpacer())

	state.documentUI = columns

	state.mainContainer.Add(state.documentUI)

	state.cryptoUIContainer.Hide()

	if pollerErr == nil {
		poller.StartPoller()
	} else {
		setStartPage("error.contextFail", "", pollerErr)
	}

	state.window.ShowAndRun()
}

func setUI(doc document.Document) {
	state.mu.Lock()
	defer state.mu.Unlock()

	var page *fyne.Container
	buttonBarObjects := []fyne.CanvasObject{state.statusBar, layout.NewSpacer()}

	switch doc := doc.(type) {
	case *document.IDDocument:
		page = pageID(doc)
	case *document.MedicalDocument:
		updateButton := widget.NewButton(t("ui.update"), updateMedicalDocHandler(doc))
		buttonBarObjects = append(buttonBarObjects, updateButton)
		page = pageMedical(doc)
	case *document.VehicleDocument:
		page = pageVehicle(doc)
	}

	savePdfButton := widget.NewButton(t("ui.savePdf"), savePdf(doc))
	saveXlsxButton := widget.NewButton(t("ui.saveXlsx"), saveXlsx(doc))
	buttonBarObjects = append(buttonBarObjects, saveXlsxButton, savePdfButton)

	buttonBar := container.New(layout.NewHBoxLayout(), buttonBarObjects...)

	state.documentUIMainContainer.RemoveAll()
	state.documentUIMainContainer.Add(page)
	state.documentUIMainContainer.Add(buttonBar)

	state.startPage.Hide()
	state.documentUIMainContainer.Show()

	resizeWindow(false)
}

func setStartPage(statusID, explanationID string, err error) {
	state.mu.Lock()
	defer state.mu.Unlock()

	status := t(statusID)
	explanation := t(explanationID)

	isError := err != nil

	if isError {
		logger.Error(err)
	} else {
		logger.Info(translation.EnglishTranslation(statusID) + " " + translation.EnglishTranslation(explanationID))
	}

	state.startPage.SetStatus(status, explanation, isError)
	state.startPage.Refresh()

	state.documentUIMainContainer.RemoveAll()

	state.documentUIMainContainer.Hide()
	state.startPage.Show()

	resizeWindow(true)
}

func setStatus(statusID string, err error) {
	isError := err != nil

	if isError {
		logger.Error(err)
	} else {
		logger.Info(translation.EnglishTranslation(statusID))
	}

	status := t(statusID)
	state.statusBar.SetStatus(status, isError)
	state.statusBar.Refresh()
}

func updateMedicalDocHandler(doc *document.MedicalDocument) func() {
	return func() {
		err := doc.UpdateValidUntilDateFromRfzo()
		if err != nil {
			logger.Error(fmt.Errorf("updating medical information: %w", err))
			dialog.ShowInformation(t("error.error"), t("error.dataUpdate"), state.window)
			return
		}

		setStatus("ui.updateSuccessful", nil)
		setUI(doc)
	}
}

func copyToClipboard(str string) bool {
	if state.window == nil {
		return false
	}

	clipboard := state.window.Clipboard()
	if clipboard == nil {
		return false
	}

	label := t("ui.contentCopied")

	clipboard.SetContent(str)
	setTimedStatus(label)

	return true
}

func setTimedStatus(label string) {
	state.statusBar.SetStatus(label, false)
	state.statusBar.Refresh()
	go func() {
		time.Sleep(2 * time.Second)
		if state.statusBar.GetStatus() == label {
			state.statusBar.SetStatus("", false)
			state.statusBar.Refresh()
		}
	}()
}

func showDocumentUI() {
	state.mainContainer.RemoveAll()
	state.mainContainer.Add(state.documentUI)
}

func resizeWindow(keepCurrentSize bool) {
	minSize := state.documentUI.MinSize()

	if keepCurrentSize {
		currentSize := state.mainContainer.Size()

		if currentSize.Height > minSize.Height {
			minSize.Height = currentSize.Height
		}

		if currentSize.Width > minSize.Width {
			minSize.Width = currentSize.Width
		}
	}

	state.window.Resize(minSize)
}
