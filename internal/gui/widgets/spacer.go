package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Spacer represents a flexible spacing widget with configurable minimum width.
type Spacer struct {
	widget.BaseWidget
	minWidth float32
}

// SpacerRenderer implements the fyne.WidgetRenderer interface for the Spacer.
type SpacerRenderer struct {
	spacer *Spacer
}

// NewSpacer creates a new Spacer widget.
func NewSpacer() *Spacer {
	spacer := &Spacer{}
	spacer.ExtendBaseWidget(spacer)
	return spacer
}

// SetMinWidth sets the minimum width for the spacer.
func (s *Spacer) SetMinWidth(width float32) {
	if width < 0 {
		width = 0
	}

	s.minWidth = width
}

// CreateRenderer creates a new renderer for the Spacer.
func (s *Spacer) CreateRenderer() fyne.WidgetRenderer {
	return &SpacerRenderer{
		spacer: s,
	}
}

// Refresh is a no-op for SpacerRenderer.
func (r *SpacerRenderer) Refresh() {}

// Layout is a no-op for SpacerRenderer.
func (r *SpacerRenderer) Layout(_ fyne.Size) {}

// MinSize returns the minimum size required for the spacer.
func (r *SpacerRenderer) MinSize() fyne.Size {
	width := r.spacer.minWidth

	if r.spacer.minWidth == 0 {
		width = theme.Padding()
	}

	return fyne.NewSize(width, 5*theme.Padding())
}

// Objects returns an empty slice as the spacer has no visual objects.
func (r *SpacerRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{}
}

// Destroy is a no-op for SpacerRenderer.
func (r *SpacerRenderer) Destroy() {}
