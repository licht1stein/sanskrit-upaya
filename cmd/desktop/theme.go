package main

import (
	"image/color"

	"fyne.io/fyne/v2"
)

// scaledTheme wraps a base theme and scales all sizes
type scaledTheme struct {
	base  fyne.Theme
	scale float32
}

func newScaledTheme(base fyne.Theme, scale float32) *scaledTheme {
	return &scaledTheme{base: base, scale: scale}
}

func (t *scaledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *scaledTheme) Font(style fyne.TextStyle) fyne.Resource {
	// Use bundled Noto Sans (Latin + Devanagari) for all text
	return resourceNotoSansMergedRegularTtf
}

func (t *scaledTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *scaledTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name) * t.scale
}
