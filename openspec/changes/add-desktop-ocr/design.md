## Context

The desktop app needs OCR functionality that matches the MCP server's capabilities but with a graphical interface. Users should be able to:

1. Set up Google Cloud credentials without using command line
2. Drag and drop images to OCR them
3. Edit and use the recognized text

Key constraints:

- Must work on Windows, macOS, and Linux
- Cannot bundle gcloud CLI - user must install it themselves
- Must handle the multi-step setup process gracefully (install gcloud -> accept TOS -> enable billing -> authenticate)

## Goals / Non-Goals

**Goals:**

- Seamless first-time setup experience with clear instructions
- Drag-and-drop OCR with immediate feedback
- Recognized text editable and searchable
- Re-runnable setup if user needs to fix credentials

**Non-Goals:**

- Bundling gcloud CLI with the app
- Image preprocessing (cropping, rotation, enhancement)
- Saving OCR history to database
- Batch processing multiple images

## Decisions

### D1: Setup Wizard Architecture

**Choice**: Modal dialog with step-by-step progress and console output area

The setup wizard will:

1. Check for gcloud CLI installation
2. Run gcloud commands in sequence
3. Show command output in a scrollable text area
4. Provide "Re-run" button to retry after user fixes issues
5. Open browser for auth/billing automatically

```
┌─────────────────────────────────────────────────────┐
│  OCR Setup Wizard                              [X]  │
├─────────────────────────────────────────────────────┤
│  Step 2 of 5: Authenticating gcloud...              │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │ $ gcloud auth login                           │  │
│  │ Your browser has been opened to visit:        │  │
│  │ https://accounts.google.com/...               │  │
│  │                                               │  │
│  │ Waiting for authentication...                 │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  A browser window should open for login.            │
│                                                     │
│  [ Re-run Setup ]              [ Cancel ]           │
└─────────────────────────────────────────────────────┘
```

**Alternatives considered:**

- External terminal window: Would require platform-specific code, harder to capture output
- Web-based setup: Adds complexity, doesn't integrate with app

### D2: OCR Window Design

**Choice**: Separate window with drop zone that transforms into editor

```
Initial state (drop zone):
┌─────────────────────────────────────────────────────┐
│  OCR                                           [X]  │
├─────────────────────────────────────────────────────┤
│                                                     │
│                                                     │
│          ┌─────────────────────────┐                │
│          │                         │                │
│          │   Drop image here       │                │
│          │   or click to browse    │                │
│          │                         │                │
│          │   PNG, JPG, TIFF, PDF   │                │
│          └─────────────────────────┘                │
│                                                     │
│                                                     │
│  [ Setup OCR... ]                                   │
└─────────────────────────────────────────────────────┘

After OCR (editor state):
┌─────────────────────────────────────────────────────┐
│  OCR                                           [X]  │
├─────────────────────────────────────────────────────┤
│  File: manuscript.jpg    Confidence: 94.2%          │
│  ┌───────────────────────────────────────────────┐  │
│  │ योगश्चित्तवृत्तिनिरोधः                        │  │
│  │                                               │  │
│  │ yogaś citta-vṛtti-nirodhaḥ                    │  │
│  │                                               │  │
│  │                                               │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  [ Copy ]  [ Search ]  [ New Image ]                │
└─────────────────────────────────────────────────────┘
```

**State transitions:**

- Drop zone -> Processing (show spinner + status)
- Processing -> Editor (on success)
- Processing -> Error (show error message + retry)
- Editor -> Drop zone ("New Image" button)

### D3: Credential Check Flow

**Choice**: Check credentials on OCR button click, show setup wizard if missing

```go
func onOCRButtonClick() {
    ctx := context.Background()
    if err := ocr.CheckCredentials(ctx); err != nil {
        showOCRSetupWizard()
        return
    }
    showOCRWindow()
}
```

This lazy approach:

- Doesn't slow down app startup
- Only prompts setup when user actually wants to use OCR
- Allows re-checking after setup completes

### D4: Drag-and-Drop Implementation

**Choice**: Use Fyne's built-in drag-drop with file path extraction

Fyne supports drag-and-drop via `SetOnDropped` on containers. The handler receives URIs which can be converted to file paths.

```go
dropZone := container.NewVBox(...)
dropZone.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
    if len(uris) == 0 {
        return
    }
    filePath := uris[0].Path()
    startOCR(filePath)
})
```

**Validation:**

- Check file extension (png, jpg, jpeg, tiff, tif, pdf)
- Check file size (< 20MB)
- Show error dialog for invalid files

### D5: Setup Wizard Command Execution

**Choice**: Run gcloud commands via exec.Command with output capture

The wizard runs the same commands as `ocr-setup`:

1. `gcloud auth login` - Opens browser
2. `gcloud projects create <project-id>` - Creates GCP project
3. `gcloud services enable vision.googleapis.com` - Enables API
4. `gcloud auth application-default login` - Sets up ADC
5. `gcloud auth application-default set-quota-project` - Sets quota project

Output is streamed to a multi-line text widget in real-time using a goroutine that reads from stdout/stderr pipes.

**Error handling:**

- If gcloud not found: Show installation instructions with link
- If auth fails: Show "Re-run" button
- If project creation fails: Explain terms acceptance
- If billing needed: Open billing URL + show continue button

## Risks / Trade-offs

### R1: gcloud CLI Dependency

**Risk**: Users may not have gcloud installed
**Mitigation**: Clear installation instructions with platform-specific links. First step of wizard checks for gcloud and shows helpful message.

### R2: Browser-based Authentication

**Risk**: Browser might not open automatically on some Linux distros
**Mitigation**: Show the URL in the console output area so user can copy/paste manually.

### R3: Setup Takes Multiple Steps

**Risk**: Users might abandon setup if it's too complex
**Mitigation**:

- Show clear progress (step N of M)
- Explain why each step is needed
- Allow re-running from any point
- Save progress (project ID) so partial setup can resume

### R4: OCR Latency

**Risk**: Large images may take several seconds to process
**Mitigation**: Show clear "Processing..." indicator with spinner. Cancel option for stuck requests.

## Migration Plan

No migration needed - this is a new feature. Existing users can continue using the app without OCR; the feature is opt-in via toolbar button.

## Open Questions

None - requirements are clear.
