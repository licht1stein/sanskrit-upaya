## ADDED Requirements

### Requirement: OCR Text Recognition

The system SHALL recognize Sanskrit and Devanagari text from images using Google Cloud Vision API.

#### Scenario: Successful OCR of printed Sanskrit text

- **GIVEN** Priya is a Sanskrit researcher with valid GCP credentials configured
- **AND** Priya has an image containing printed Devanagari text
- **WHEN** Priya submits the image for OCR via the MCP tool
- **THEN** the system returns the recognized text preserving line breaks
- **AND** returns a confidence score between 0.0 and 1.0

#### Scenario: OCR of mixed script image

- **GIVEN** Priya has an image containing both Devanagari and IAST transliteration
- **WHEN** Priya submits the image for OCR
- **THEN** the system recognizes both scripts
- **AND** returns text in the order it appears in the image

#### Scenario: No text detected in image

- **GIVEN** Priya has an image with no recognizable text
- **WHEN** Priya submits the image for OCR
- **THEN** the system returns empty text with zero confidence
- **AND** does not return an error

#### Scenario: OCR request timeout

- **GIVEN** the Google Cloud Vision API is slow or unresponsive
- **WHEN** Priya submits an image for OCR
- **AND** the API does not respond within 30 seconds
- **THEN** the system returns a timeout error

### Requirement: GCP Credential Authentication

The system SHALL support Google Cloud Platform authentication via gcloud CLI OAuth or environment variable.

#### Scenario: Authenticate via gcloud application-default credentials

- **GIVEN** Arun has run `gcloud auth application-default login`
- **AND** credentials exist at the default location
- **WHEN** Arun uses the OCR tool
- **THEN** OCR requests use those credentials automatically

#### Scenario: Authenticate via environment variable

- **GIVEN** Arun sets the `GOOGLE_APPLICATION_CREDENTIALS` environment variable
- **WHEN** Arun uses the OCR tool
- **THEN** OCR requests use the service account credentials

#### Scenario: Missing credentials error

- **GIVEN** a user has no GCP credentials configured
- **WHEN** the user attempts to use OCR
- **THEN** the system returns an error directing them to run `sanskrit-mcp ocr-setup`

#### Scenario: Invalid credentials error

- **GIVEN** a user has credentials that are expired or invalid
- **WHEN** the user attempts to use OCR
- **THEN** the system returns an error suggesting re-authentication

### Requirement: OCR Setup Command

The MCP server SHALL provide an `ocr-setup` subcommand to help users configure credentials.

#### Scenario: Check credentials when not configured

- **GIVEN** Maya has no GCP credentials configured
- **WHEN** Maya runs `sanskrit-mcp ocr-setup`
- **THEN** the system displays instructions for installing gcloud CLI
- **AND** shows the command to run for authentication

#### Scenario: Check credentials when configured

- **GIVEN** Maya has valid GCP credentials
- **WHEN** Maya runs `sanskrit-mcp ocr-setup`
- **THEN** the system confirms credentials are found
- **AND** tests that the Vision API is accessible

#### Scenario: Check credentials when invalid

- **GIVEN** Maya has expired or invalid credentials
- **WHEN** Maya runs `sanskrit-mcp ocr-setup`
- **THEN** the system reports the credentials are invalid
- **AND** shows the command to re-authenticate

### Requirement: MCP OCR Tool

The MCP server SHALL expose a `sanskrit_ocr` tool for LLM workflows to recognize text from images.

#### Scenario: OCR via base64-encoded image

- **GIVEN** Claude is assisting Priya with manuscript analysis
- **AND** Priya provides a base64-encoded image with data URI prefix
- **WHEN** Claude calls the `sanskrit_ocr` tool with the image data
- **THEN** the tool returns the recognized text and confidence score

#### Scenario: OCR via file path

- **GIVEN** Claude is running locally with file system access
- **AND** Priya has an image file at a known path
- **WHEN** Claude calls the `sanskrit_ocr` tool with the file path
- **THEN** the tool reads the image and returns recognized text

#### Scenario: Unsupported image format

- **GIVEN** a user provides an image in an unsupported format
- **WHEN** the `sanskrit_ocr` tool is called
- **THEN** the tool returns an error indicating supported formats

### Requirement: Image Format Support

The system SHALL support common image formats for OCR input.

#### Scenario: PNG image support

- **GIVEN** an image in PNG format
- **WHEN** submitted for OCR
- **THEN** the system processes it successfully

#### Scenario: JPEG image support

- **GIVEN** an image in JPEG format
- **WHEN** submitted for OCR
- **THEN** the system processes it successfully

#### Scenario: TIFF image support

- **GIVEN** an image in TIFF format
- **WHEN** submitted for OCR
- **THEN** the system processes it successfully

#### Scenario: Image size limit checked before reading

- **GIVEN** an image larger than 20MB
- **WHEN** submitted for OCR
- **THEN** the system returns an error indicating the size limit
- **AND** does not attempt to read the full image into memory
