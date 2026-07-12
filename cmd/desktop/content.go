package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/licht1stein/sanskrit-upaya/pkg/transliterate"
)

// formatWordHeader renders a headword for display, appending its Devanagari
// transliteration (separated by sep) when the word is given in IAST. Non-IAST
// (already-Devanagari) or empty words are returned unchanged.
func formatWordHeader(word, sep string) string {
	if word == "" || transliterate.IsDevanagari(word) {
		return word
	}
	deva := transliterate.IASTToDevanagari(word)
	if deva == "" {
		return word
	}
	return word + sep + deva
}

// cleanHTML removes HTML tags and converts breaks to newlines
func cleanHTML(content string) string {
	result := content
	result = strings.ReplaceAll(result, "<BR>", "\n\n")
	result = strings.ReplaceAll(result, "<br>", "\n\n")
	result = strings.ReplaceAll(result, "<P>", "\n\n")
	result = strings.ReplaceAll(result, "<p>", "\n\n")
	// Remove other HTML tags
	for _, tag := range []string{"<b>", "</b>", "<i>", "</i>", "<B>", "</B>", "<I>", "</I>"} {
		result = strings.ReplaceAll(result, tag, "")
	}
	return result
}

// createSelectableArticleContent creates a selectable Label for article content
func createSelectableArticleContent(content string) fyne.CanvasObject {
	cleanContent := cleanHTML(content)
	label := widget.NewLabel(cleanContent)
	label.Wrapping = fyne.TextWrapWord
	label.Selectable = true
	return label
}
