## ADDED Requirements

### Requirement: OCR Window Access

The desktop application SHALL provide an OCR button in the toolbar that opens the OCR window.

#### Scenario: User opens OCR window with valid credentials

- **WHEN** Maya clicks the OCR button in the toolbar
- **AND** Google Cloud credentials are configured
- **THEN** the OCR window opens with the drop zone view

#### Scenario: User opens OCR window without credentials

- **WHEN** Maya clicks the OCR button in the toolbar
- **AND** Google Cloud credentials are not configured
- **THEN** the OCR setup wizard opens instead

#### Scenario: User opens OCR via keyboard shortcut

- **WHEN** Maya presses Ctrl+O (or Cmd+O on macOS)
- **THEN** the same credential check and window opening occurs

### Requirement: OCR Setup Wizard

The desktop application SHALL provide a setup wizard to configure Google Cloud credentials for OCR.

#### Scenario: gcloud CLI not installed

- **WHEN** Maya opens the setup wizard
- **AND** gcloud CLI is not found on the system
- **THEN** the wizard shows installation instructions with a link to cloud.google.com/sdk/docs/install
- **AND** provides a "Re-run Setup" button to check again after installation

#### Scenario: Setup wizard runs gcloud authentication

- **WHEN** Maya starts the setup process
- **AND** gcloud CLI is installed
- **THEN** the wizard runs `gcloud auth login` in the background
- **AND** displays the command output in a scrollable text area
- **AND** opens the browser automatically for authentication

#### Scenario: Setup wizard creates GCP project

- **WHEN** gcloud authentication succeeds
- **THEN** the wizard creates a new GCP project (or uses existing one)
- **AND** shows progress "Step 2 of 5: Creating project..."
- **AND** displays command output

#### Scenario: Setup wizard handles terms acceptance error

- **WHEN** project creation fails because Google Cloud terms have not been accepted
- **THEN** the wizard shows an explanation about accepting terms at console.cloud.google.com
- **AND** provides a "Re-run Setup" button

#### Scenario: Setup wizard enables Vision API

- **WHEN** project creation succeeds
- **THEN** the wizard enables the Vision API for the project
- **AND** shows progress "Step 3 of 5: Enabling Vision API..."

#### Scenario: Setup wizard configures Application Default Credentials

- **WHEN** Vision API is enabled
- **THEN** the wizard runs `gcloud auth application-default login`
- **AND** opens browser for authentication
- **AND** shows progress "Step 4 of 5: Setting up credentials..."

#### Scenario: Setup wizard handles billing requirement

- **WHEN** credential verification indicates billing is required
- **THEN** the wizard opens the billing enablement URL in the browser
- **AND** shows a message explaining free tier (1000 images/month)
- **AND** provides a "Continue" button to verify after billing is enabled

#### Scenario: Setup wizard completes successfully

- **WHEN** all setup steps complete
- **AND** credential verification passes
- **THEN** the wizard shows "Setup Complete!" message
- **AND** closes the wizard
- **AND** opens the OCR window automatically

### Requirement: OCR Drop Zone

The OCR window SHALL provide a drop zone for image files.

#### Scenario: User sees drop zone on window open

- **WHEN** Maya opens the OCR window
- **THEN** she sees a centered drop zone with text "Drop image here or click to browse"
- **AND** supported formats listed: "PNG, JPG, TIFF, PDF"

#### Scenario: User drags valid image file

- **WHEN** Maya drags a PNG file onto the drop zone
- **AND** the file is under 20MB
- **THEN** the window transitions to processing state
- **AND** shows a spinner with "Processing manuscript.png..."

#### Scenario: User drags invalid file type

- **WHEN** Maya drags a .doc file onto the drop zone
- **THEN** an error dialog appears: "Unsupported file type. Please use PNG, JPG, TIFF, or PDF."

#### Scenario: User drags oversized file

- **WHEN** Maya drags a 25MB image onto the drop zone
- **THEN** an error dialog appears: "File too large. Maximum size is 20MB."

#### Scenario: User clicks to browse

- **WHEN** Maya clicks on the drop zone
- **THEN** a file browser dialog opens filtered to supported image formats

### Requirement: OCR Processing

The OCR window SHALL show progress during OCR processing.

#### Scenario: OCR processing in progress

- **WHEN** an image is submitted for OCR
- **THEN** the window shows a spinner
- **AND** displays "Processing filename.jpg..."
- **AND** provides a Cancel button

#### Scenario: User cancels OCR processing

- **WHEN** Maya clicks Cancel during processing
- **THEN** the OCR request is cancelled
- **AND** the window returns to the drop zone state

#### Scenario: OCR completes successfully

- **WHEN** OCR processing completes
- **THEN** the window transitions to the editor state
- **AND** displays the recognized text

#### Scenario: OCR fails with error

- **WHEN** OCR processing fails (e.g., timeout, API error)
- **THEN** an error message is displayed
- **AND** a "Retry" button is provided
- **AND** a "New Image" button returns to drop zone

### Requirement: OCR Result Editor

The OCR window SHALL display recognized text in an editable text area.

#### Scenario: Recognized text displayed

- **WHEN** OCR completes successfully
- **THEN** the recognized text appears in a multi-line text entry
- **AND** the text is editable by the user
- **AND** a header shows the filename and confidence score (e.g., "manuscript.jpg - Confidence: 94.2%")

#### Scenario: User copies recognized text

- **WHEN** Maya clicks the "Copy" button
- **THEN** the text content is copied to the system clipboard
- **AND** brief feedback is shown (button text changes to "Copied!" briefly)

#### Scenario: User searches recognized text

- **WHEN** Maya selects some text in the editor
- **AND** clicks the "Search" button
- **THEN** the selected text is placed in the main window's search box
- **AND** a search is performed
- **AND** focus moves to the main window

#### Scenario: User searches without selection

- **WHEN** Maya clicks "Search" without selecting text
- **AND** the text area contains content
- **THEN** all the recognized text is used as the search query (trimmed)

#### Scenario: User wants to OCR another image

- **WHEN** Maya clicks "New Image" button
- **THEN** the window returns to the drop zone state
- **AND** the previous OCR result is cleared

### Requirement: Setup OCR Button

The OCR window SHALL provide access to the setup wizard.

#### Scenario: User re-runs setup from OCR window

- **WHEN** Maya clicks "Setup OCR..." button in the OCR window
- **THEN** the setup wizard opens
- **AND** she can re-configure credentials
