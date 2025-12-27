// Package widgets contains custom GUI widgets for the application.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ubavic/bas-celik/v2/internal/gui/translation"
)

// Toolbar represents the application toolbar with buttons and reader selection.
type Toolbar struct {
	widget.BaseWidget
	readers           []string
	onOpenAbout       func()
	onOpenPreferences func()
	onReaderChange    func(string)
	selectedReader    string
	showReaders       bool
}

// ToolbarRenderer implements the fyne.WidgetRenderer interface for the Toolbar.
type ToolbarRenderer struct {
	toolbar           *Toolbar
	aboutButton       *widget.Button
	preferencesButton *widget.Button
	container         *fyne.Container
	readersLabel      *widget.Label
	readersSelect     *widget.Select
}

// NewToolbar creates a new Toolbar instance.
func NewToolbar(onOpenAbout, onOpenPreferences func(), showReaders bool) *Toolbar {
	toolbar := &Toolbar{
		readers:           nil,
		onOpenAbout:       onOpenAbout,
		onOpenPreferences: onOpenPreferences,
		showReaders:       showReaders,
	}

	toolbar.ExtendBaseWidget(toolbar)
	return toolbar
}

// HookReaderChange sets the callback function for reader change events.
func (t *Toolbar) HookReaderChange(hook func(string)) {
	t.onReaderChange = hook
}

// CreateRenderer creates a new renderer for the Toolbar.
func (t *Toolbar) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabel(translation.Translate("ui.reader"))

	onChange := func(reader string) {
		if t.onReaderChange != nil {
			t.onReaderChange(reader)
		}
	}

	readersSelect := widget.NewSelect(t.readers, onChange)

	preferencesButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), t.onOpenPreferences)
	preferencesButton.Importance = widget.LowImportance

	aboutButton := widget.NewButtonWithIcon("", theme.InfoIcon(), t.onOpenAbout)
	aboutButton.Importance = widget.LowImportance

	var horizontalContainer *fyne.Container
	if t.showReaders {
		horizontalContainer = container.New(layout.NewHBoxLayout(), label, readersSelect, layout.NewSpacer(), preferencesButton, aboutButton)
	} else {
		horizontalContainer = container.New(layout.NewHBoxLayout(), layout.NewSpacer(), preferencesButton, aboutButton)
	}

	return &ToolbarRenderer{
		toolbar:           t,
		aboutButton:       aboutButton,
		preferencesButton: preferencesButton,
		container:         horizontalContainer,
		readersLabel:      label,
		readersSelect:     readersSelect,
	}
}

// Refresh updates the visual representation of the toolbar.
func (t *ToolbarRenderer) Refresh() {
	if t.toolbar.showReaders {
		t.readersSelect.SetOptions(t.toolbar.readers)
		t.readersSelect.Selected = t.toolbar.selectedReader

		if len(t.toolbar.readers) <= 1 {
			t.readersSelect.Disable()
		} else {
			t.readersSelect.Enable()
		}

		t.readersSelect.Refresh()
	}

	t.aboutButton.Refresh()
}

// Layout positions the toolbar elements within the given size.
func (t *ToolbarRenderer) Layout(s fyne.Size) {
	availableWidth := s.Width
	availableWidth -= t.aboutButton.Size().Width
	availableWidth -= t.preferencesButton.MinSize().Width
	if t.toolbar.showReaders {
		availableWidth -= t.readersLabel.MinSize().Width
	}
	availableWidth -= 2 * theme.InnerPadding()
	t.container.Resize(s)
	t.readersSelect.Resize(fyne.Size{Width: availableWidth, Height: s.Height})
}

// MinSize returns the minimum size required for the toolbar.
func (t *ToolbarRenderer) MinSize() fyne.Size {
	return t.container.MinSize()
}

// Objects returns the visual objects that make up the toolbar.
func (t *ToolbarRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{t.aboutButton, t.preferencesButton, t.container}

	if t.toolbar.showReaders {
		objects = append(objects, t.readersSelect)
	}

	return objects
}

// Destroy is a no-op for ToolbarRenderer.
func (t *ToolbarRenderer) Destroy() {}

// SetReaders updates the list of available readers and sets the selected reader.
func (t *Toolbar) SetReaders(readers []string, selectedReader string) {
	t.readers = make([]string, len(readers))
	copy(t.readers, readers)

	t.selectedReader = selectedReader

	t.Refresh()
}
