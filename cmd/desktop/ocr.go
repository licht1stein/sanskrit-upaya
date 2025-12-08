package main

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/licht1stein/sanskrit-upaya/pkg/ocr"
)

// Supported image extensions for OCR
var ocrSupportedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".tiff": true,
	".tif":  true,
	".pdf":  true,
}

// MaxOCRFileSize is the maximum file size for OCR (20MB)
const MaxOCRFileSize = 20 * 1024 * 1024

// OCRWindowState represents the current state of the OCR window
type OCRWindowState int

const (
	OCRStateDropZone OCRWindowState = iota
	OCRStateProcessing
	OCRStateResult
	OCRStateError
)

// OCRWindow manages the OCR UI
type OCRWindow struct {
	window      fyne.Window
	app         fyne.App
	mainWindow  fyne.Window
	searchEntry *widget.Entry // Reference to main window search entry
	doSearch    func(string)  // Reference to main window search function

	// State
	state        OCRWindowState
	currentFile  string
	resultText   string
	confidence   float64
	errorMessage string

	// UI containers for each state
	dropZoneContent   *fyne.Container
	processingContent *fyne.Container
	resultContent     *fyne.Container
	errorContent      *fyne.Container
	mainContainer     *fyne.Container

	// Result view widgets (for direct update)
	resultHeaderLabel *widget.Label
	resultTextEntry   *widget.Entry

	// Error view widgets
	errorMessageLabel *widget.Label

	// Processing state
	cancelFunc context.CancelFunc
	mu         sync.Mutex
}

// NewOCRWindow creates a new OCR window
func NewOCRWindow(app fyne.App, mainWindow fyne.Window, searchEntry *widget.Entry, doSearch func(string)) *OCRWindow {
	w := &OCRWindow{
		app:         app,
		mainWindow:  mainWindow,
		searchEntry: searchEntry,
		doSearch:    doSearch,
		state:       OCRStateDropZone,
	}

	w.window = app.NewWindow("OCR")
	w.window.Resize(fyne.NewSize(500, 400))

	w.buildUI()
	return w
}

func (w *OCRWindow) buildUI() {
	// Build all state UIs
	w.buildDropZoneUI()
	w.buildProcessingUI()
	w.buildResultUI()
	w.buildErrorUI()

	// Main container that holds current state
	w.mainContainer = container.NewStack(w.dropZoneContent)

	w.window.SetContent(container.NewPadded(w.mainContainer))
}

func (w *OCRWindow) buildDropZoneUI() {
	// Drop zone visual
	dropLabel := widget.NewLabel("Drop image here\nor click to browse")
	dropLabel.Alignment = fyne.TextAlignCenter
	dropLabel.TextStyle = fyne.TextStyle{Bold: true}

	formatLabel := widget.NewLabel("PNG, JPG, TIFF, PDF")
	formatLabel.Alignment = fyne.TextAlignCenter
	formatLabel.Importance = widget.LowImportance

	// Create a bordered drop zone
	dropContent := container.NewVBox(
		dropLabel,
		formatLabel,
	)

	// Make it look like a drop zone with a border
	dropZoneBorder := canvas.NewRectangle(theme.Color(theme.ColorNameInputBorder))
	dropZoneBorder.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	dropZoneBorder.StrokeWidth = 2
	dropZoneBorder.CornerRadius = 10
	dropZoneBorder.FillColor = theme.Color(theme.ColorNameInputBackground)

	dropZoneStack := container.NewStack(
		dropZoneBorder,
		container.NewCenter(dropContent),
	)

	// Make clickable for file browser
	browseBtn := widget.NewButton("Browse Files...", func() {
		w.showFileBrowser()
	})

	// Setup button
	setupBtn := widget.NewButton("Setup OCR...", func() {
		ShowOCRSetupDialog(w.window, w.app, nil)
	})
	setupBtn.Importance = widget.LowImportance

	w.dropZoneContent = container.NewBorder(
		nil,
		container.NewVBox(
			container.NewCenter(browseBtn),
			widget.NewSeparator(),
			container.NewCenter(setupBtn),
		),
		nil, nil,
		container.NewPadded(dropZoneStack),
	)

	// Set up drag and drop on the window
	w.window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 {
			return
		}
		w.handleDroppedFile(uris[0])
	})
}

func (w *OCRWindow) buildProcessingUI() {
	spinner := widget.NewActivity()
	spinner.Start()

	statusLabel := widget.NewLabel("Processing...")
	statusLabel.Alignment = fyne.TextAlignCenter

	fileLabel := widget.NewLabel("")
	fileLabel.Alignment = fyne.TextAlignCenter
	fileLabel.Importance = widget.LowImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		w.cancelProcessing()
	})

	w.processingContent = container.NewCenter(
		container.NewVBox(
			spinner,
			statusLabel,
			fileLabel,
			widget.NewSeparator(),
			cancelBtn,
		),
	)
}

func (w *OCRWindow) buildResultUI() {
	// Header with file info - store reference for updates
	w.resultHeaderLabel = widget.NewLabel("File: - | Confidence: -")
	w.resultHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Editable text area for result - store reference for updates
	w.resultTextEntry = widget.NewMultiLineEntry()
	w.resultTextEntry.Wrapping = fyne.TextWrapWord
	w.resultTextEntry.SetMinRowsVisible(12)

	// White background for text area
	textBg := canvas.NewRectangle(color.White)
	textWithBg := container.NewStack(textBg, w.resultTextEntry)

	// Buttons
	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		w.window.Clipboard().SetContent(w.resultTextEntry.Text)
	})

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		w.saveText()
	})

	searchBtn := widget.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		w.searchText(w.resultTextEntry.Text)
	})
	searchBtn.Importance = widget.HighImportance

	newImageBtn := widget.NewButton("New Image", func() {
		// Open a new OCR window instead of clearing this one
		newWindow := NewOCRWindow(w.app, w.mainWindow, w.searchEntry, w.doSearch)
		newWindow.Show()
	})

	buttonRow := container.NewHBox(copyBtn, saveBtn, searchBtn, newImageBtn)

	w.resultContent = container.NewBorder(
		container.NewVBox(w.resultHeaderLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), container.NewCenter(buttonRow)),
		nil, nil,
		container.NewScroll(textWithBg),
	)
}

func (w *OCRWindow) buildErrorUI() {
	errorLabel := widget.NewLabel("Error")
	errorLabel.TextStyle = fyne.TextStyle{Bold: true}
	errorLabel.Alignment = fyne.TextAlignCenter

	// Store reference for updates
	w.errorMessageLabel = widget.NewLabel("")
	w.errorMessageLabel.Wrapping = fyne.TextWrapWord
	w.errorMessageLabel.Alignment = fyne.TextAlignCenter

	retryBtn := widget.NewButton("Retry", func() {
		if w.currentFile != "" {
			w.startOCR(w.currentFile)
		} else {
			w.showDropZone()
		}
	})

	newImageBtn := widget.NewButton("New Image", func() {
		// Open a new OCR window instead of clearing this one
		newWindow := NewOCRWindow(w.app, w.mainWindow, w.searchEntry, w.doSearch)
		newWindow.Show()
	})

	buttonRow := container.NewHBox(retryBtn, newImageBtn)

	w.errorContent = container.NewCenter(
		container.NewVBox(
			errorLabel,
			w.errorMessageLabel,
			widget.NewSeparator(),
			container.NewCenter(buttonRow),
		),
	)
}

func (w *OCRWindow) showFileBrowser() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}
		if reader == nil {
			return // User cancelled
		}
		defer reader.Close()

		uri := reader.URI()
		w.handleDroppedFile(uri)
	}, w.window)

	// Filter to supported image types
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".tiff", ".tif", ".pdf"}))
	fd.Show()
}

func (w *OCRWindow) handleDroppedFile(uri fyne.URI) {
	path := uri.Path()

	// Validate extension
	ext := strings.ToLower(filepath.Ext(path))
	if !ocrSupportedExtensions[ext] {
		dialog.ShowError(
			fmt.Errorf("Unsupported file type '%s'.\nPlease use PNG, JPG, TIFF, or PDF.", ext),
			w.window,
		)
		return
	}

	// Validate file size
	info, err := os.Stat(path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Could not read file: %v", err), w.window)
		return
	}
	if info.Size() > MaxOCRFileSize {
		dialog.ShowError(
			fmt.Errorf("File too large (%d MB).\nMaximum size is 20 MB.", info.Size()/(1024*1024)),
			w.window,
		)
		return
	}

	w.startOCR(path)
}

func (w *OCRWindow) startOCR(filePath string) {
	w.mu.Lock()
	w.currentFile = filePath
	w.state = OCRStateProcessing
	w.mu.Unlock()

	// Show processing UI
	w.showProcessing(filePath)

	// Create cancellable context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	w.mu.Lock()
	w.cancelFunc = cancel
	w.mu.Unlock()

	go func() {
		defer cancel()

		// Create OCR client
		client, err := ocr.NewClient(ctx)
		if err != nil {
			w.handleOCRError(err)
			return
		}
		defer client.Close()

		// Perform OCR
		result, err := client.RecognizeFile(ctx, filePath)
		if err != nil {
			w.handleOCRError(err)
			return
		}

		w.mu.Lock()
		w.resultText = result.Text
		w.confidence = result.Confidence
		w.state = OCRStateResult
		w.mu.Unlock()

		fyne.Do(func() {
			w.showResult(filepath.Base(filePath), result.Text, result.Confidence)
		})
	}()
}

func (w *OCRWindow) cancelProcessing() {
	w.mu.Lock()
	if w.cancelFunc != nil {
		w.cancelFunc()
	}
	w.mu.Unlock()
	w.showDropZone()
}

func (w *OCRWindow) handleOCRError(err error) {
	w.mu.Lock()
	w.state = OCRStateError
	w.errorMessage = err.Error()
	w.mu.Unlock()

	fyne.Do(func() {
		w.showError(err.Error())
	})
}

func (w *OCRWindow) showDropZone() {
	w.mu.Lock()
	w.state = OCRStateDropZone
	w.currentFile = ""
	w.mu.Unlock()

	w.mainContainer.RemoveAll()
	w.mainContainer.Add(w.dropZoneContent)
	w.mainContainer.Refresh()
}

func (w *OCRWindow) showProcessing(filePath string) {
	filename := filepath.Base(filePath)

	// Update processing content
	// Find and update the labels
	for _, obj := range w.processingContent.Objects {
		if vbox, ok := obj.(*fyne.Container); ok {
			for _, child := range vbox.Objects {
				if label, ok := child.(*widget.Label); ok {
					if label.Text == "Processing..." || strings.HasPrefix(label.Text, "Processing") {
						label.SetText(fmt.Sprintf("Processing %s...", filename))
					}
				}
			}
		}
	}

	w.mainContainer.RemoveAll()
	w.mainContainer.Add(w.processingContent)
	w.mainContainer.Refresh()
}

func (w *OCRWindow) showResult(filename, text string, confidence float64) {
	// Update widgets directly using stored references
	w.resultHeaderLabel.SetText(fmt.Sprintf("File: %s | Confidence: %.1f%%", filename, confidence*100))
	w.resultTextEntry.SetText(text)

	w.mainContainer.RemoveAll()
	w.mainContainer.Add(w.resultContent)
	w.mainContainer.Refresh()
}

func (w *OCRWindow) showError(message string) {
	// Update error message directly using stored reference
	w.errorMessageLabel.SetText(message)

	w.mainContainer.RemoveAll()
	w.mainContainer.Add(w.errorContent)
	w.mainContainer.Refresh()
}

func (w *OCRWindow) searchText(fullText string) {
	// Use selected text if available, otherwise show hint
	searchQuery := w.resultTextEntry.SelectedText()
	searchQuery = strings.TrimSpace(searchQuery)

	if searchQuery == "" {
		dialog.ShowInformation("Select Text", "Please select the word or phrase you want to search.", w.window)
		return
	}

	// Limit search query length
	if len(searchQuery) > 100 {
		searchQuery = searchQuery[:100]
	}

	// Set search text in main window and perform search
	if w.searchEntry != nil {
		w.searchEntry.SetText(searchQuery)
	}
	if w.doSearch != nil {
		w.doSearch(searchQuery)
	}

	// Focus main window
	w.mainWindow.RequestFocus()
}

func (w *OCRWindow) saveText() {
	text := w.resultTextEntry.Text
	if strings.TrimSpace(text) == "" {
		return
	}

	// Generate default filename from source file
	defaultName := "ocr-result.txt"
	if w.currentFile != "" {
		base := filepath.Base(w.currentFile)
		ext := filepath.Ext(base)
		defaultName = strings.TrimSuffix(base, ext) + ".txt"
	}

	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}
		if writer == nil {
			return // User cancelled
		}
		defer writer.Close()

		_, err = writer.Write([]byte(text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save: %v", err), w.window)
			return
		}
	}, w.window)

	fd.SetFileName(defaultName)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
	fd.Show()
}

// Show displays the OCR window
func (w *OCRWindow) Show() {
	w.showDropZone()
	w.window.Show()
}

// GetWindow returns the underlying fyne.Window
func (w *OCRWindow) GetWindow() fyne.Window {
	return w.window
}
