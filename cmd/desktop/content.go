package main

import (
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/licht1stein/sanskrit-upaya/pkg/transliterate"
)

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

// createArticleContent creates content with highlighted search terms (both IAST and Devanagari)
// Uses regex matching like the Clojure web version for proper Unicode handling
func createArticleContent(content string, searchTerm string) fyne.CanvasObject {
	cleanContent := cleanHTML(content)

	// If no search term, just use a simple label
	if searchTerm == "" {
		label := widget.NewLabel(cleanContent)
		label.Wrapping = fyne.TextWrapWord
		label.Selectable = true
		return label
	}

	// Get both IAST and Devanagari versions of search term
	searchTerms := transliterate.ToSearchTerms(searchTerm)

	// Build regex pattern like Clojure: (?i)(term1|term2)
	// Escape regex special chars and join with |
	var escapedTerms []string
	for _, term := range searchTerms {
		escapedTerms = append(escapedTerms, regexp.QuoteMeta(term))
	}
	pattern := "(?i)(" + strings.Join(escapedTerms, "|") + ")"

	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fallback to simple label if regex fails
		label := widget.NewLabel(cleanContent)
		label.Wrapping = fyne.TextWrapWord
		label.Selectable = true
		return label
	}

	// Find all matches
	matches := re.FindAllStringIndex(cleanContent, -1)

	// If no matches, return simple label
	if len(matches) == 0 {
		label := widget.NewLabel(cleanContent)
		label.Wrapping = fyne.TextWrapWord
		label.Selectable = true
		return label
	}

	// Build segments from matches
	var segments []widget.RichTextSegment
	lastEnd := 0

	for _, m := range matches {
		start, end := m[0], m[1]

		// Add text before match
		if start > lastEnd {
			segments = append(segments, &widget.TextSegment{
				Text:  cleanContent[lastEnd:start],
				Style: widget.RichTextStyleInline,
			})
		}

		// Add highlighted match with bold + color
		segments = append(segments, &widget.TextSegment{
			Text: cleanContent[start:end],
			Style: widget.RichTextStyle{
				Inline:    true,
				ColorName: theme.ColorNamePrimary, // Use theme's primary color
				TextStyle: fyne.TextStyle{Bold: true},
			},
		})

		lastEnd = end
	}

	// Add remaining text
	if lastEnd < len(cleanContent) {
		segments = append(segments, &widget.TextSegment{
			Text:  cleanContent[lastEnd:],
			Style: widget.RichTextStyleInline,
		})
	}

	richText := widget.NewRichText(segments...)
	richText.Wrapping = fyne.TextWrapWord
	return richText
}
