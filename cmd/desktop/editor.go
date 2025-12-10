package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/licht1stein/sanskrit-upaya/pkg/state"
	"github.com/licht1stein/sanskrit-upaya/pkg/transliterate"
)

// EditorWindow manages the transliteration editor UI
type EditorWindow struct {
	window     fyne.Window
	app        fyne.App
	mainWindow fyne.Window
	settings   *state.Store

	// Text entries
	iastEntry *widget.Entry
	devaEntry *widget.Entry

	// Track which field is being edited to avoid infinite loops
	updating bool
	closed   bool
}

// NewEditorWindow creates a new transliteration editor window
func NewEditorWindow(app fyne.App, mainWindow fyne.Window, settings *state.Store) *EditorWindow {
	w := &EditorWindow{
		app:        app,
		mainWindow: mainWindow,
		settings:   settings,
	}

	w.window = app.NewWindow("Transliteration Editor")
	w.window.Resize(fyne.NewSize(700, 400))

	w.window.SetOnClosed(func() {
		w.closed = true
		// Save content on close
		w.saveContent()
	})

	w.buildUI()
	w.loadContent()
	return w
}

func (w *EditorWindow) buildUI() {
	// IAST panel (left side)
	iastLabel := widget.NewLabel("IAST")
	iastLabel.TextStyle = fyne.TextStyle{Bold: true}
	iastLabel.Alignment = fyne.TextAlignCenter

	w.iastEntry = widget.NewMultiLineEntry()
	w.iastEntry.Wrapping = fyne.TextWrapWord
	w.iastEntry.SetMinRowsVisible(12)
	w.iastEntry.SetPlaceHolder("Type IAST here...")

	iastBg := canvas.NewRectangle(color.White)
	iastWithBg := container.NewStack(iastBg, w.iastEntry)

	iastCopyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		w.window.Clipboard().SetContent(w.iastEntry.Text)
	})

	iastClearBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		w.updating = true
		w.iastEntry.SetText("")
		w.devaEntry.SetText("")
		w.updating = false
	})

	iastPanel := container.NewBorder(
		iastLabel,
		container.NewCenter(container.NewHBox(iastCopyBtn, iastClearBtn)),
		nil, nil,
		container.NewScroll(iastWithBg),
	)

	// Devanagari panel (right side)
	devaLabel := widget.NewLabel("Devanagari")
	devaLabel.TextStyle = fyne.TextStyle{Bold: true}
	devaLabel.Alignment = fyne.TextAlignCenter

	w.devaEntry = widget.NewMultiLineEntry()
	w.devaEntry.Wrapping = fyne.TextWrapWord
	w.devaEntry.SetMinRowsVisible(12)
	w.devaEntry.SetPlaceHolder("Type Devanagari here...")

	devaBg := canvas.NewRectangle(color.White)
	devaWithBg := container.NewStack(devaBg, w.devaEntry)

	devaCopyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		w.window.Clipboard().SetContent(w.devaEntry.Text)
	})

	devaClearBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		w.updating = true
		w.iastEntry.SetText("")
		w.devaEntry.SetText("")
		w.updating = false
	})

	devaPanel := container.NewBorder(
		devaLabel,
		container.NewCenter(container.NewHBox(devaCopyBtn, devaClearBtn)),
		nil, nil,
		container.NewScroll(devaWithBg),
	)

	// Set up bidirectional transliteration
	w.iastEntry.OnChanged = func(text string) {
		if w.updating {
			return
		}
		w.updating = true
		w.devaEntry.SetText(transliterate.IASTToDevanagari(text))
		w.updating = false
	}

	w.devaEntry.OnChanged = func(text string) {
		if w.updating {
			return
		}
		w.updating = true
		w.iastEntry.SetText(transliterate.DevanagariToIAST(text))
		w.updating = false
	}

	// Split view
	splitView := container.NewHSplit(iastPanel, devaPanel)
	splitView.SetOffset(0.5)

	w.window.SetContent(container.NewPadded(splitView))
}

// Show displays the editor window
func (w *EditorWindow) Show() {
	if w.closed {
		return
	}
	w.window.Show()
	w.window.RequestFocus()
}

// IsClosed returns true if the window was closed
func (w *EditorWindow) IsClosed() bool {
	return w.closed
}

// GetWindow returns the underlying fyne.Window
func (w *EditorWindow) GetWindow() fyne.Window {
	return w.window
}

// saveContent persists the editor content to settings
func (w *EditorWindow) saveContent() {
	if w.settings != nil && w.iastEntry != nil {
		w.settings.Set("editor_content", w.iastEntry.Text)
	}
}

// loadContent restores the editor content from settings
func (w *EditorWindow) loadContent() {
	if w.settings != nil {
		if saved := w.settings.Get("editor_content"); saved != "" {
			w.updating = true
			w.iastEntry.SetText(saved)
			w.devaEntry.SetText(transliterate.IASTToDevanagari(saved))
			w.updating = false
		}
	}
}
