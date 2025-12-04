## Context

Sanskrit Upaya MCP server needs OCR to recognize text from manuscript images, enabling LLMs to perform scan-to-search workflows. Google Cloud Vision API is the recommended choice based on:

- 97% accuracy on printed Sanskrit (highest among tested options)
- Auto-detects language (no hints required)
- Official Go SDK available
- Free tier: 1000 images/month

## Goals / Non-Goals

**Goals:**

- Enable OCR of Sanskrit/Devanagari text from images via MCP tool
- Provide easy setup via `gcloud` CLI OAuth flow
- Provide clear feedback when credentials are missing or invalid

**Non-Goals:**

- Desktop app UI for OCR (future change)
- Bundling GCP credentials with the app
- Supporting offline OCR (future: Tesseract fallback)
- Batch processing / PDF OCR (future enhancement)

## Decisions

### Decision 1: Google Cloud Vision API

**Choice**: Use `cloud.google.com/go/vision/apiv1` Go SDK

**Rationale**:

- Best accuracy for Sanskrit (97% on printed text)
- Official Go SDK, well-maintained
- Handles both `TEXT_DETECTION` and `DOCUMENT_TEXT_DETECTION`
- Auto-detects Sanskrit without language hints

**Alternatives considered**:

- Tesseract: Lower accuracy for Sanskrit, but free/offline
- Microsoft Document AI: Good accuracy but less Go support
- Claude Vision: Good but adds API dependency, higher cost for high volume

### Decision 2: Credential Authentication

**Choice**: Browser-based OAuth via `gcloud` CLI (primary), env var fallback

**Primary method** (easiest for users):

```bash
gcloud auth application-default login
```

- Opens browser, user logs in with Google account
- Creates `~/.config/gcloud/application_default_credentials.json`
- Go SDK automatically picks up these credentials
- No service account, no JSON key file, no env var needed

**Fallback method** (advanced users, CI):

- `GOOGLE_APPLICATION_CREDENTIALS` environment variable pointing to service account JSON

**Rationale**:

- OAuth flow is familiar (like "Sign in with Google")
- No manual file management
- `gcloud` CLI is well-documented and widely used
- Fallback covers CI/automation scenarios

### Decision 3: Setup Command

**Choice**: Add `sanskrit-mcp ocr-setup` subcommand

```bash
$ sanskrit-mcp ocr-setup

Checking Google Cloud credentials...

❌ No credentials found.

To enable OCR:

1. Install Google Cloud CLI: https://cloud.google.com/sdk/docs/install

2. Run: gcloud auth application-default login

3. A browser will open. Log in with your Google account.

4. Done! OCR will work automatically.

Note: Free tier is 1000 images/month, then $1.50/1000.
```

If credentials found:

```bash
$ sanskrit-mcp ocr-setup

✓ Google Cloud credentials found
✓ Vision API accessible

OCR is ready to use.
```

**Rationale**:

- Single command to check status and get guidance
- Tests credentials actually work (not just exist)
- Clear next steps when credentials missing

### Decision 4: MCP Tool Design

**Choice**: Single `sanskrit_ocr` tool with simple interface

**Input**:

- `image_data`: base64-encoded image (with `data:image/...;base64,` prefix) OR file path

**Output**:

- `text`: recognized text (preserves line breaks)
- `confidence`: average confidence score (0.0-1.0)

**Input detection**:

- Starts with `data:image/` → base64
- Starts with `/` or drive letter → file path

**Rationale**:

- Claude Desktop always sends images as base64 with data URI prefix
- File paths for Claude Code local file access
- Prefix-based detection is unambiguous

### Decision 5: Timeouts and Limits

**Choice**: 30-second timeout, 20MB size limit

- API timeout: 30 seconds (prevents hanging on slow network)
- Image size: 20MB max (check BEFORE reading into memory)
- Size estimation for base64: `len(base64) * 3 / 4`

**Rationale**:

- GCP Vision typically responds in 2-5 seconds
- 30s covers slow connections without infinite hang
- Size check prevents memory exhaustion

### Decision 6: Confidence Score

**Choice**: Average of word-level confidences from GCP response

- GCP returns per-word confidence scores
- Aggregate as arithmetic mean of all word confidences
- If no text detected: confidence = 0.0
- Range: 0.0 (no confidence) to 1.0 (perfect)

### Decision 7: Error Handling

**Choice**: Clear, actionable error messages

| Condition                 | Error Message                                                                  |
| ------------------------- | ------------------------------------------------------------------------------ |
| No credentials configured | "No Google Cloud credentials. Run: sanskrit-mcp ocr-setup"                     |
| Invalid credentials       | "Google Cloud credentials invalid. Run: gcloud auth application-default login" |
| API quota exceeded        | "Google Cloud Vision API quota exceeded. Check your GCP console."              |
| Unsupported image format  | "Unsupported image format. Use PNG, JPG, or TIFF."                             |
| Image too large           | "Image exceeds 20MB limit."                                                    |
| Timeout                   | "OCR request timed out after 30 seconds."                                      |
| No text detected          | Returns empty text with confidence 0.0 (not an error)                          |

## Risks / Trade-offs

### Risk: gcloud CLI not installed

Users need to install `gcloud` CLI first.

**Mitigation**: Clear instructions in `ocr-setup` output with install link.

### Trade-off: No offline fallback

Requires internet + Google account.

**Accepted because**:

- GCP provides best accuracy
- Tesseract fallback can be added later
- Most researchers have internet access

## Open Questions

None—scope is intentionally minimal for this change.
