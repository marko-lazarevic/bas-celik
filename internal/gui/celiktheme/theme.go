// Package celiktheme provides a custom Fyne theme with light and dark modes.
package celiktheme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Theme implements a custom Fyne theme with light and dark modes.
type Theme struct {
	systemDecides bool
	dark          bool
}

var (
	defaultColor = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x10}

	lightColors = map[fyne.ThemeColorName]color.Color{
		theme.ColorNameBackground:       color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		theme.ColorNameButton:           color.NRGBA{R: 0xF0, G: 0xF0, B: 0xF0, A: 0xFF},
		theme.ColorNameDisabledButton:   color.NRGBA{R: 0xD0, G: 0xD0, B: 0xD0, A: 0xFF},
		theme.ColorNameDisabled:         color.NRGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xFF},
		theme.ColorNameError:            color.NRGBA{R: 0xF0, G: 0x47, B: 0x3B, A: 0xFF},
		theme.ColorNameFocus:            color.NRGBA{R: 0xDE, G: 0xEB, B: 0xFA, A: 0xFF},
		theme.ColorNameForeground:       color.NRGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xFF},
		theme.ColorNameForegroundOnPrimary: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		theme.ColorNameHeaderBackground: color.NRGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xFF},
		theme.ColorNameHover:            color.NRGBA{R: 0x00, G: 0x00, B: 0x40, A: 0x10},
		theme.ColorNameHyperlink:        color.NRGBA{R: 0x50, G: 0x50, B: 0xA0, A: 0xFF},
		theme.ColorNameInputBackground:  color.NRGBA{R: 0xF0, G: 0xF0, B: 0xF0, A: 0xFF},
		theme.ColorNameInputBorder:      color.NRGBA{R: 0xDA, G: 0xDA, B: 0xDA, A: 0xFF},
		theme.ColorNameMenuBackground:   color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		theme.ColorNameOverlayBackground: color.NRGBA{R: 0xF9, G: 0xF9, B: 0xF9, A: 0xFF},
		theme.ColorNamePlaceHolder:      color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xFF},
		theme.ColorNamePressed:          color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
		theme.ColorNamePrimary:          color.NRGBA{R: 0x5A, G: 0x73, B: 0x8F, A: 0xFF},
		theme.ColorNameScrollBar:        color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x99},
		theme.ColorNameSelection:        color.NRGBA{R: 0xDE, G: 0xEB, B: 0xFA, A: 0xFF},
		theme.ColorNameShadow:           color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x10},
	}

	darkColors = map[fyne.ThemeColorName]color.Color{
		theme.ColorNameBackground:       color.NRGBA{R: 0x10, G: 0x10, B: 0x13, A: 0xFF},
		theme.ColorNameButton:           color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xFF},
		theme.ColorNameDisabledButton:   color.NRGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xFF},
		theme.ColorNameDisabled:         color.NRGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xFF},
		theme.ColorNameError:            color.NRGBA{R: 0xF0, G: 0x47, B: 0x3B, A: 0xFF},
		theme.ColorNameFocus:            color.NRGBA{R: 0x23, G: 0x20, B: 0x24, A: 0xFF},
		theme.ColorNameForeground:       color.NRGBA{R: 0xD0, G: 0xD0, B: 0xD0, A: 0xFF},
		theme.ColorNameForegroundOnPrimary: color.NRGBA{R: 0xD9, G: 0xD0, B: 0xD0, A: 0xFF},
		theme.ColorNameHeaderBackground: color.NRGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xFF},
		theme.ColorNameHover:            color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x10},
		theme.ColorNameHyperlink:        color.NRGBA{R: 0x80, G: 0x90, B: 0xF0, A: 0xFF},
		theme.ColorNameInputBackground:  color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xFF},
		theme.ColorNameInputBorder:      color.NRGBA{R: 0xDA, G: 0xDA, B: 0xDA, A: 0x00},
		theme.ColorNameMenuBackground:   color.NRGBA{R: 0x15, G: 0x15, B: 0x15, A: 0xFF},
		theme.ColorNameOverlayBackground: color.NRGBA{R: 0x15, G: 0x15, B: 0x17, A: 0xFF},
		theme.ColorNamePlaceHolder:      color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xFF},
		theme.ColorNamePressed:          color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
		theme.ColorNamePrimary:          color.NRGBA{R: 0x41, G: 0x4D, B: 0x7A, A: 0xFF},
		theme.ColorNameScrollBar:        color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x50},
		theme.ColorNameSelection:        color.NRGBA{R: 0x23, G: 0x20, B: 0x24, A: 0xFF},
		theme.ColorNameShadow:           color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x40},
	}
)

// NewTheme creates a new Theme instance based on the user's selection.
func NewTheme(themeSelection int) Theme {
	theme := Theme{}

	if themeSelection <= 0 {
		theme.systemDecides = true
	} else if themeSelection == 2 {
		theme.dark = true
	}

	return theme
}

// Color returns the color for the given color name and variant.
func (t Theme) Color(colorName fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if t.systemDecides {
		if v == theme.VariantLight || v == 2 {
			return lightTheme(colorName)
		}
		return darkTheme(colorName)
	} else if t.dark {
		return darkTheme(colorName)
	}
	return lightTheme(colorName)
}

// Font returns the font resource for the given text style.
func (Theme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(s)
}

// Icon returns the icon resource for the given icon name.
func (Theme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

// Size returns the size for the given size name.
func (Theme) Size(s fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(s)
}

// CornerRadius returns the corner radius for UI elements.
func (Theme) CornerRadius() float32 {
	return 3
}

func lightTheme(c fyne.ThemeColorName) color.Color {
	if col, ok := lightColors[c]; ok {
		return col
	}
	return defaultColor
}

func darkTheme(c fyne.ThemeColorName) color.Color {
	if col, ok := darkColors[c]; ok {
		return col
	}
	return defaultColor
}
