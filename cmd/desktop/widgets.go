// Custom widgets for the Sanskrit Upaya application
package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// pillLabel is a custom widget that displays text with a pill/tag background
type pillLabel struct {
	widget.BaseWidget
	text string
}

func newPillLabel(text string) *pillLabel {
	p := &pillLabel{text: text}
	p.ExtendBaseWidget(p)
	return p
}

func (p *pillLabel) CreateRenderer() fyne.WidgetRenderer {
	// Subtle background - very light gray, almost transparent
	bg := canvas.NewRectangle(color.RGBA{R: 240, G: 240, B: 242, A: 255})
	bg.CornerRadius = 3

	// Muted text color
	label := canvas.NewText(p.text, color.RGBA{R: 140, G: 145, B: 155, A: 255})
	label.TextSize = 10

	return &pillRenderer{bg: bg, label: label, pill: p}
}

type pillRenderer struct {
	bg    *canvas.Rectangle
	label *canvas.Text
	pill  *pillLabel
}

func (r *pillRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	// Center the text
	textSize := fyne.MeasureText(r.pill.text, r.label.TextSize, r.label.TextStyle)
	xOffset := (size.Width - textSize.Width) / 2
	r.label.Move(fyne.NewPos(xOffset, 1))
}

func (r *pillRenderer) MinSize() fyne.Size {
	// Fixed width for all pills (accommodate longest dict code like "mw72" or "ap90")
	return fyne.NewSize(38, 14)
}

func (r *pillRenderer) Refresh() {
	r.label.Text = r.pill.text
	r.label.Refresh()
	r.bg.Refresh()
}

func (r *pillRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.label}
}

func (r *pillRenderer) Destroy() {}
