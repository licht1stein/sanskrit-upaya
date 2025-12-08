# Change: Add OCR Feature to Desktop Application

## Why

Scholars working with Sanskrit manuscripts and printed texts currently have no way to OCR images directly from the desktop app. The MCP server already has OCR support (`sanskrit_ocr` tool), but desktop users must use external tools, copy text manually, then paste into the app for dictionary lookup. Integrating OCR into the desktop app creates a seamless workflow: drag image -> OCR -> edit/copy recognized text -> search dictionaries.

## What Changes

- **NEW** OCR window accessible via toolbar button
- **NEW** Google Cloud setup wizard (GUI version of `ocr-setup`)
- **NEW** Drag-and-drop image handling with OCR processing
- **NEW** Text editor view for recognized text with copy/search integration

## Impact

- Affected specs: None (new capability)
- Affected code:
  - `cmd/desktop/main.go` (add OCR button to toolbar, launch OCR window)
  - `cmd/desktop/ocr.go` (new - OCR window implementation)
  - `cmd/desktop/ocr_setup.go` (new - setup wizard implementation)
  - Reuses existing `pkg/ocr/` package

## Scope

This change focuses on desktop GUI integration:

- Reuses existing `pkg/ocr/` package for Vision API calls
- First-time setup wizard guides users through Google Cloud configuration
- Drag-and-drop for PNG, JPG, TIFF, PDF files
- Recognized text displayed in editable text area
- Copy button and "Search" button to search recognized text in dictionaries

Out of scope (future work):

- Batch OCR for multiple files
- Image cropping/preprocessing
- OCR history/saved results

## Effort Estimation (CHAI)

| Dimension         | Score | Notes                                            |
| ----------------- | ----- | ------------------------------------------------ |
| Claude Complexity | 3     | New Fyne windows, drag-drop, async OCR + setup   |
| Error Probability | 2     | Reuses proven pkg/ocr; Fyne patterns established |
| Human Attention   | 3     | UI review needed, setup flow UX is critical      |
| Iteration Risk    | 2     | Clear requirements, existing patterns to follow  |

**Summary**: Medium complexity. Main challenges are the setup wizard UX (guiding users through gcloud installation and browser-based auth) and drag-and-drop integration.
