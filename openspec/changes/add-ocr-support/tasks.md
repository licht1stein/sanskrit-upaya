## 1. Core OCR Package

- [x] 1.1 Add `cloud.google.com/go/vision/v2` dependency to go.mod
- [x] 1.2 Create `pkg/ocr/ocr.go` with `NewClient(ctx)` using default credentials
- [x] 1.3 Implement `RecognizeText(ctx, imageData []byte) (text string, confidence float64, error)`
- [x] 1.4 Add 30-second context timeout for API calls
- [x] 1.5 Add image format detection via magic bytes (PNG, JPG, TIFF)
- [x] 1.6 Add image size validation BEFORE reading (check file size or estimate from base64 length)
- [x] 1.7 Calculate confidence as average of word-level confidences
- [x] 1.8 Implement `CheckCredentials(ctx) error` to verify API access
- [ ] 1.9 Write unit tests for `pkg/ocr/` with mocked GCP client

## 2. Setup Command

- [x] 2.1 Add `ocr-setup` subcommand to `cmd/mcp/main.go`
- [x] 2.2 Check for credentials at default gcloud location
- [x] 2.3 Check for `GOOGLE_APPLICATION_CREDENTIALS` env var
- [x] 2.4 Test Vision API access if credentials found
- [x] 2.5 Display appropriate message (success/missing/invalid)
- [x] 2.6 Show gcloud install URL and auth command when credentials missing
- [x] 2.7 (Added) Automated project creation and Vision API enablement
- [x] 2.8 (Added) Open browser for billing setup with user prompt

## 3. MCP Tool Integration

- [x] 3.1 Add `sanskrit_ocr` tool to MCP server
- [x] 3.2 Detect input type: `data:image/` prefix → base64, `/` or drive letter → file path
- [x] 3.3 Handle base64 input: strip data URI prefix, decode
- [x] 3.4 Handle file path input: read file
- [x] 3.5 Use sync.Once pattern for OCR client (like existing DB pattern)
- [x] 3.6 Return text and confidence in response
- [x] 3.7 Return actionable error messages per design.md

## 4. Documentation

- [x] 4.1 Update CLAUDE.md with `sanskrit_ocr` tool description
- [x] 4.2 Document `ocr-setup` command usage
- [x] 4.3 Add GCP free tier note (1000 images/month)

## 5. Validation

- [x] 5.1 Test `ocr-setup` with no credentials
- [x] 5.2 Test `ocr-setup` after `gcloud auth application-default login`
- [x] 5.3 Test OCR with printed Sanskrit text image
- [x] 5.4 Test OCR with Devanagari manuscript image
- [x] 5.5 Test error handling (no credentials, invalid format, oversized image)
- [x] 5.6 Test MCP tool end-to-end with Claude Code
