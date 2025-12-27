package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// StatusBar represents a status bar widget that displays status messages.
type StatusBar struct {
	widget.BaseWidget
	status string
	err    bool
}

// StatusBarRenderer implements the fyne.WidgetRenderer interface for the StatusBar.
type StatusBarRenderer struct {
	bar        *StatusBar
	statusText *canvas.Text
}

// NewStatusBar creates a new StatusBar instance.
func NewStatusBar() *StatusBar {
	statusBar := &StatusBar{
		status: "",
		err:    true,
	}
	statusBar.ExtendBaseWidget(statusBar)
	return statusBar
}

// SetStatus updates the status bar with a new message and error state.
func (sb *StatusBar) SetStatus(status string, err bool) {
	sb.status = status
	sb.err = err
}

// GetStatus returns the current status message.
func (sb *StatusBar) GetStatus() string {
	return sb.status
}

// CreateRenderer creates a new renderer for the StatusBar.
func (sb *StatusBar) CreateRenderer() fyne.WidgetRenderer {
	statusText := canvas.NewText(sb.status, theme.Color(theme.ColorNameForeground))
	statusText.TextSize = 11
	statusText.Color = theme.Color(theme.ColorNameError)

	return &StatusBarRenderer{
		bar:        sb,
		statusText: statusText,
	}
}

// Refresh updates the visual representation of the status bar.
func (r *StatusBarRenderer) Refresh() {
	r.statusText.Text = r.bar.status

	if r.bar.err {
		r.statusText.Color = theme.Color(theme.ColorNameError)
	} else {
		r.statusText.Color = theme.Color(theme.ColorNameForeground)
	}

	r.statusText.Refresh()
}

// Layout positions the status bar elements within the given size.
func (r *StatusBarRenderer) Layout(_ fyne.Size) {
	r.statusText.Move(fyne.Position{X: theme.Padding(), Y: 2 * theme.Padding()})
}

// MinSize returns the minimum size required for the status bar.
func (r *StatusBarRenderer) MinSize() fyne.Size {
	ts1 := fyne.MeasureText(r.statusText.Text, r.statusText.TextSize, r.statusText.TextStyle)
	return fyne.NewSize(ts1.Width+theme.Padding(), 2*ts1.Height)
}

// Objects returns the visual objects that make up the status bar.
func (r *StatusBarRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.statusText}
}

// Destroy is a no-op for StatusBarRenderer.
func (r *StatusBarRenderer) Destroy() {}
