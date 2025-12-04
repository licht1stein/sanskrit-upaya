# Change: Add OCR Support for Sanskrit/Devanagari Text Recognition (MCP Server)

## Why

Researchers working with Sanskrit manuscripts, printed books, and inscriptions need to digitize text before searching dictionaries. Currently this requires switching to external OCR tools, copying results, then returning to Claude. Integrating OCR into the MCP server enables a seamless LLM workflow: user provides image → Claude OCRs it → Claude searches dictionaries → user gets definitions.

Google Cloud Vision API provides 97% accuracy on printed Sanskrit/Devanagari text—the best available option based on [sanskrit-coders research](https://sanskrit-coders.github.io/content/ocr/ocr-ing/).

## What Changes

- **NEW** `pkg/ocr/` - OCR package wrapping Google Cloud Vision API
- **NEW** MCP tool `sanskrit_ocr` - OCR images from Claude Code / LLM workflows

## Impact

- Affected specs: None (new capability)
- Affected code:
  - `pkg/ocr/ocr.go` (new)
  - `cmd/mcp/main.go` (add OCR tool)

## Scope

This change focuses on MCP server integration only:

- Supports image files (PNG, JPG, TIFF) and base64-encoded images
- Credentials via `GOOGLE_APPLICATION_CREDENTIALS` environment variable (standard GCP pattern)
- No bundled API key—users provide their own GCP credentials

Future considerations (separate changes):

- Desktop app OCR UI with credential settings
- Batch OCR for multiple pages
- PDF support
- Local Tesseract fallback for offline use

## Effort Estimation (CHAI)

| Dimension         | Score | Notes                                           |
| ----------------- | ----- | ----------------------------------------------- |
| Claude Complexity | 2     | New pkg + MCP tool, reuses existing patterns    |
| Error Probability | 2     | Standard GCP auth via env var, clear API        |
| Human Attention   | 2     | Straightforward, no credential storage review   |
| Iteration Risk    | 2     | Well-defined GCP API, clear acceptance criteria |

**Summary**: Low-medium complexity. Single credential source (env var) simplifies implementation. Desktop integration can follow as separate change.
