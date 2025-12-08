## 1. Setup Wizard

- [x] 1.1 Create `cmd/desktop/ocr_setup.go` with setup wizard window
- [x] 1.2 Implement gcloud CLI detection (`exec.LookPath("gcloud")`)
- [x] 1.3 Create "gcloud not found" view with platform-specific install links
- [x] 1.4 Implement command execution with real-time output streaming to text widget
- [x] 1.5 Implement step-by-step flow: auth -> project create -> enable API -> ADC login -> set quota
- [x] 1.6 Add progress indicator ("Step N of 5: Description...")
- [x] 1.7 Implement browser opening for auth URLs (reuse `openBrowser` from cmd/mcp)
- [x] 1.8 Add "Re-run Setup" button for retry after user fixes issues
- [x] 1.9 Handle billing-required error: open billing URL, show "Continue" button
- [x] 1.10 Verify credentials work at end of setup via `ocr.CheckCredentials`

## 2. OCR Window - Drop Zone

- [x] 2.1 Create `cmd/desktop/ocr.go` with OCR window
- [x] 2.2 Implement drop zone UI with centered instruction text
- [x] 2.3 Add "click to browse" functionality via file dialog
- [x] 2.4 Implement drag-and-drop handler via `window.SetOnDropped`
- [x] 2.5 Validate dropped files (extension, size < 20MB)
- [x] 2.6 Show error dialog for invalid files
- [x] 2.7 Add "Setup OCR..." button that opens setup wizard

## 3. OCR Window - Processing State

- [x] 3.1 Implement processing state with spinner/progress indicator
- [x] 3.2 Show filename being processed
- [x] 3.3 Call `pkg/ocr` client with file path
- [x] 3.4 Handle timeout (30s) gracefully with error message
- [x] 3.5 Implement cancel button during processing

## 4. OCR Window - Editor State

- [x] 4.1 Create editor view with multi-line text entry (editable)
- [x] 4.2 Display filename and confidence score in header
- [x] 4.3 Add "Copy" button to copy text to clipboard
- [x] 4.4 Add "Search" button to search recognized text in main dictionary window
- [x] 4.5 Add "New Image" button to return to drop zone state
- [x] 4.6 Style text area for Sanskrit/Devanagari (appropriate font size)

## 5. Main Window Integration

- [x] 5.1 Add OCR button to main window toolbar (next to settings button)
- [x] 5.2 On OCR button click: check credentials, show setup wizard if missing
- [x] 5.3 If credentials valid: open OCR window
- [x] 5.4 Add keyboard shortcut for OCR window (Ctrl/Cmd+O)

## 6. Shared Utilities

- [x] 6.1 Extract `openBrowser` function to shared package or reuse from mcp
- [x] 6.2 Extract `getOrCreateOCRProjectID` to shared package for reuse
- [x] 6.3 Extract gcloud helper functions (isGcloudInstalled, isGcloudAuthenticated, etc.)

## 7. Testing

- [x] 7.1 Test setup wizard with no gcloud installed
- [x] 7.2 Test setup wizard with gcloud installed but no credentials
- [x] 7.3 Test setup wizard complete flow (manual)
- [x] 7.4 Test drag-and-drop with valid image
- [x] 7.5 Test drag-and-drop with invalid file type
- [x] 7.6 Test drag-and-drop with oversized file
- [x] 7.7 Test OCR result display and copy functionality
- [x] 7.8 Test "Search" button integration with main window

## 8. Documentation

- [x] 8.1 Update CLAUDE.md with desktop OCR feature description
- [ ] 8.2 Add OCR section to future README/user guide
