package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Group represents a titled group container for grouping related widgets.
type Group struct {
	widget.BaseWidget
	name    string
	objects []fyne.CanvasObject
}

// GroupRenderer implements the fyne.WidgetRenderer interface for the Group.
type GroupRenderer struct {
	group    *Group
	nameText *canvas.Text
	column   *fyne.Container
}

// NewGroup creates a new Group with the given name and objects.
func NewGroup(name string, objects ...fyne.CanvasObject) *Group {
	group := &Group{
		name:    name,
		objects: objects,
	}
	group.ExtendBaseWidget(group)
	return group
}

// CreateRenderer creates a new renderer for the Group.
func (g *Group) CreateRenderer() fyne.WidgetRenderer {
	nameText := canvas.NewText(g.name, theme.Color(theme.ColorNameForeground))
	nameText.TextStyle.Bold = true
	nameText.TextSize = 14

	nameText.Move(fyne.NewPos(2*theme.Padding(), 0))

	column := container.New(layout.NewVBoxLayout(), g.objects...)
	column.Move(fyne.NewPos(theme.Padding(), 6*theme.Padding()))

	return &GroupRenderer{
		group:    g,
		nameText: nameText,
		column:   column,
	}
}

// Refresh updates the visual representation of the group.
func (r *GroupRenderer) Refresh() {
	r.column.Refresh()
	r.nameText.Refresh()
}

// Layout positions the group elements within the given size.
func (r *GroupRenderer) Layout(s fyne.Size) {
	r.column.Move(fyne.Position{X: theme.Padding(), Y: 6 * theme.Padding()})
	r.column.Layout.Layout(r.group.objects, s.SubtractWidthHeight(2*theme.Padding(), 0))
}

// MinSize returns the minimum size required for the group.
func (r *GroupRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.column.MinSize().Width+2*theme.Padding(), r.column.MinSize().Height+10*theme.Padding())
}

// Objects returns the visual objects that make up the group.
func (r *GroupRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.column, r.nameText}
}

// Destroy is a no-op for GroupRenderer.
func (r *GroupRenderer) Destroy() {}
