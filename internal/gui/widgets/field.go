package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Field represents a labeled field widget with hover and copy functionality.
type Field struct {
	widget.BaseWidget
	name, value  string
	minWidth     float32
	hovered      bool
	copied       bool
	valueChanged bool
}

// FieldRenderer implements the fyne.WidgetRenderer interface for the Field.
type FieldRenderer struct {
	field      *Field
	background *canvas.Rectangle
	nameText   *canvas.Text
	valueLabel *widget.Label
}

// NewField creates a new Field with the given name, value, and minimum width.
func NewField(name, value string, minWidth float32) *Field {
	field := &Field{
		name:     name,
		value:    value,
		minWidth: minWidth,
	}
	field.ExtendBaseWidget(field)
	return field
}

// CreateRenderer creates a new renderer for the Field.
func (f *Field) CreateRenderer() fyne.WidgetRenderer {
	nameText := canvas.NewText(f.name, theme.Color(theme.ColorNameForeground))
	nameText.TextSize = 11

	valueText := widget.NewLabel(f.value)
	valueText.Alignment = fyne.TextAlignLeading
	valueText.Wrapping = fyne.TextWrapWord
	valueText.Resize(fyne.NewSize(f.minWidth, valueText.MinSize().Height))

	background := canvas.NewRectangle(color.Transparent)
	background.CornerRadius = theme.InputRadiusSize()

	return &FieldRenderer{
		field:      f,
		background: background,
		nameText:   nameText,
		valueLabel: valueText,
	}
}

// Cursor returns the cursor type when hovering over the field.
func (f *Field) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// MouseIn handles mouse entering the field area.
func (f *Field) MouseIn(*desktop.MouseEvent) {
	f.copied = false
	f.hovered = true
	f.Refresh()
}

// MouseMoved handles mouse movement within the field.
func (f *Field) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut handles mouse leaving the field area.
func (f *Field) MouseOut() {
	f.hovered = false
	f.Refresh()
}

// Tapped handles tap events on the field, copying the value to clipboard.
func (f *Field) Tapped(*fyne.PointEvent) {
	if copyToClipboard(f.value) {
		f.copied = true
	}
	f.Refresh()
}

// SetValue updates the field's value.
func (f *Field) SetValue(value string) {
	f.value = value
	f.valueChanged = true
	f.Refresh()
}

// Refresh updates the visual representation of the field.
func (r *FieldRenderer) Refresh() {
	if r.field.valueChanged {
		r.valueLabel.SetText(r.field.value)
		r.field.valueChanged = false
	}

	if r.field.hovered && !r.field.copied {
		r.background.FillColor = theme.Color(theme.ColorNameButton)
	} else {
		r.background.FillColor = color.Transparent
	}

	r.valueLabel.Refresh()
	r.background.Refresh()
}

// Layout positions the field elements within the given size.
func (r *FieldRenderer) Layout(s fyne.Size) {
	r.nameText.Move(fyne.Position{X: theme.Padding(), Y: 0})
	r.valueLabel.Resize(s.SubtractWidthHeight(0, 2*theme.Padding()))
	r.valueLabel.Move(fyne.Position{X: -theme.Padding(), Y: theme.Padding()})
	r.background.Resize(s)
}

// MinSize returns the minimum size required for the field.
func (r *FieldRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.field.minWidth+2*theme.Padding(), r.valueLabel.MinSize().Height-theme.Padding())
}

// Objects returns the visual objects that make up the field.
func (r *FieldRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.valueLabel, r.nameText}
}

// Destroy is a no-op for FieldRenderer.
func (r *FieldRenderer) Destroy() {}
