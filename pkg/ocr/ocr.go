// Package ocr provides text recognition for Sanskrit/Devanagari images using Google Cloud Vision API.
package ocr

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	vision "cloud.google.com/go/vision/v2/apiv1"
	visionpb "cloud.google.com/go/vision/v2/apiv1/visionpb"
)

// Client wraps a Google Cloud Vision API client.
type Client struct {
	client *vision.ImageAnnotatorClient
}

// MaxImageSize is the maximum allowed image size (20MB).
const MaxImageSize = 20 * 1024 * 1024

// DefaultTimeout is the default timeout for OCR requests.
const DefaultTimeout = 30 * time.Second

// Result contains the OCR output.
type Result struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// ErrNoCredentials indicates Google Cloud credentials are not configured.
var ErrNoCredentials = errors.New("no Google Cloud credentials. Run: sanskrit-mcp ocr-setup")

// ErrInvalidCredentials indicates credentials exist but are invalid.
var ErrInvalidCredentials = errors.New("Google Cloud credentials invalid. Run: gcloud auth application-default login")

// ErrNoQuotaProject indicates credentials exist but no quota project is set.
var ErrNoQuotaProject = errors.New("no quota project configured - see instructions above")

// ErrAPINotEnabled indicates the Vision API is not enabled for the project.
var ErrAPINotEnabled = errors.New("Vision API not enabled for your project")

// ErrBillingDisabled indicates billing is not enabled for the project.
var ErrBillingDisabled = errors.New("billing not enabled for project")

// ErrQuotaExceeded indicates API quota has been exceeded.
var ErrQuotaExceeded = errors.New("Google Cloud Vision API quota exceeded. Check your GCP console")

// ErrUnsupportedFormat indicates the image format is not supported.
var ErrUnsupportedFormat = errors.New("unsupported image format. Use PNG, JPG, TIFF, or PDF")

// ErrImageTooLarge indicates the image exceeds the size limit.
var ErrImageTooLarge = errors.New("image exceeds 20MB limit")

// ErrTimeout indicates the OCR request timed out.
var ErrTimeout = errors.New("OCR request timed out after 30 seconds")

// NewClient creates a new OCR client using default credentials.
// It looks for credentials in the following order:
// 1. GOOGLE_APPLICATION_CREDENTIALS environment variable
// 2. gcloud application-default credentials (~/.config/gcloud/application_default_credentials.json)
func NewClient(ctx context.Context) (*Client, error) {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		if isCredentialError(err) {
			return nil, ErrNoCredentials
		}
		return nil, fmt.Errorf("failed to create Vision client: %w", err)
	}
	return &Client{client: client}, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// RecognizeText performs OCR on the given image data.
// Returns the recognized text and average confidence score.
func (c *Client) RecognizeText(ctx context.Context, imageData []byte) (*Result, error) {
	// Check image size
	if len(imageData) > MaxImageSize {
		return nil, ErrImageTooLarge
	}

	// Detect and validate image format
	if !isValidImageFormat(imageData) {
		return nil, ErrUnsupportedFormat
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	// Create request for DOCUMENT_TEXT_DETECTION (better for printed text)
	req := &visionpb.BatchAnnotateImagesRequest{
		Requests: []*visionpb.AnnotateImageRequest{
			{
				Image: &visionpb.Image{Content: imageData},
				Features: []*visionpb.Feature{
					{Type: visionpb.Feature_DOCUMENT_TEXT_DETECTION},
				},
			},
		},
	}

	resp, err := c.client.BatchAnnotateImages(ctx, req)
	if err != nil {
		return nil, classifyError(err)
	}

	// Check for errors in response
	if len(resp.Responses) == 0 {
		return &Result{Text: "", Confidence: 0.0}, nil
	}

	imageResp := resp.Responses[0]
	if imageResp.Error != nil {
		return nil, fmt.Errorf("OCR failed: %s", imageResp.Error.Message)
	}

	// Handle case where no text was detected
	if imageResp.FullTextAnnotation == nil || imageResp.FullTextAnnotation.Text == "" {
		return &Result{Text: "", Confidence: 0.0}, nil
	}

	// Calculate average confidence from word-level confidences
	confidence := calculateConfidence(imageResp.FullTextAnnotation)

	return &Result{
		Text:       imageResp.FullTextAnnotation.Text,
		Confidence: confidence,
	}, nil
}

// RecognizeFile performs OCR on an image file.
func (c *Client) RecognizeFile(ctx context.Context, path string) (*Result, error) {
	// Check file size before reading
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if info.Size() > MaxImageSize {
		return nil, ErrImageTooLarge
	}

	imageData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return c.RecognizeText(ctx, imageData)
}

// RecognizeBase64 performs OCR on a base64-encoded image.
// Accepts both raw base64 and data URI format (data:image/png;base64,...).
func (c *Client) RecognizeBase64(ctx context.Context, b64 string) (*Result, error) {
	// Strip data URI prefix if present
	data := b64
	if strings.HasPrefix(b64, "data:image/") {
		parts := strings.SplitN(b64, ",", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid base64 data URI format")
		}
		data = parts[1]
	}

	// Estimate decoded size before decoding
	estimatedSize := len(data) * 3 / 4
	if estimatedSize > MaxImageSize {
		return nil, ErrImageTooLarge
	}

	imageData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return c.RecognizeText(ctx, imageData)
}

// CheckCredentials verifies that valid credentials are available and the API is accessible.
func CheckCredentials(ctx context.Context) error {
	client, err := NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Try a simple API call to verify credentials work
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Use a minimal 1x1 PNG to test API access
	minimalPNG := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}

	// Make a minimal API call - don't care about results, just testing credentials
	_, err = client.RecognizeText(ctx, minimalPNG)
	// Any error except "no text found" indicates a problem
	if err != nil && err != ErrUnsupportedFormat {
		return err
	}
	return nil
}

// isValidImageFormat checks if the image data starts with a valid magic number.
func isValidImageFormat(data []byte) bool {
	if len(data) < 8 {
		return false
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}

	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}

	// TIFF: 49 49 2A 00 (little-endian) or 4D 4D 00 2A (big-endian)
	if (data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A && data[3] == 0x00) ||
		(data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00 && data[3] == 0x2A) {
		return true
	}

	// PDF: 25 50 44 46 (%PDF)
	if data[0] == 0x25 && data[1] == 0x50 && data[2] == 0x44 && data[3] == 0x46 {
		return true
	}

	return false
}

// calculateConfidence calculates the average confidence from word-level confidences.
func calculateConfidence(resp *visionpb.TextAnnotation) float64 {
	if resp == nil || len(resp.Pages) == 0 {
		return 0.0
	}

	var totalConfidence float64
	var wordCount int

	for _, page := range resp.Pages {
		for _, block := range page.Blocks {
			for _, para := range block.Paragraphs {
				for _, word := range para.Words {
					if word.Confidence > 0 {
						totalConfidence += float64(word.Confidence)
						wordCount++
					}
				}
			}
		}
	}

	if wordCount == 0 {
		return 0.0
	}

	return totalConfidence / float64(wordCount)
}

// isCredentialError checks if an error is related to missing credentials.
func isCredentialError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "could not find default credentials") ||
		strings.Contains(errStr, "credentials") ||
		strings.Contains(errStr, "authentication")
}

// classifyError converts API errors to user-friendly errors.
func classifyError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for timeout
	if strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "timeout") {
		return ErrTimeout
	}

	// Check for quota project missing (most common issue with ADC)
	if strings.Contains(errStr, "quota project") ||
		(strings.Contains(errStr, "PermissionDenied") && strings.Contains(errStr, "quota")) {
		return ErrNoQuotaProject
	}

	// Check for billing disabled
	if strings.Contains(errStr, "BILLING_DISABLED") ||
		strings.Contains(errStr, "billing to be enabled") ||
		strings.Contains(errStr, "enable billing") {
		return ErrBillingDisabled
	}

	// Check for API not enabled
	if strings.Contains(errStr, "SERVICE_DISABLED") ||
		strings.Contains(errStr, "API has not been used") ||
		strings.Contains(errStr, "it is disabled") {
		return ErrAPINotEnabled
	}

	// Check for quota exceeded
	if strings.Contains(errStr, "RESOURCE_EXHAUSTED") ||
		(strings.Contains(errStr, "quota") && strings.Contains(errStr, "exceeded")) {
		return ErrQuotaExceeded
	}

	// Check for invalid credentials
	if strings.Contains(errStr, "invalid") &&
		(strings.Contains(errStr, "credentials") || strings.Contains(errStr, "token")) {
		return ErrInvalidCredentials
	}

	// Check for missing credentials
	if isCredentialError(err) {
		return ErrNoCredentials
	}

	return fmt.Errorf("OCR failed: %w", err)
}
