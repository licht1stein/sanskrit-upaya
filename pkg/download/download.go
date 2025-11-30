// Package download handles downloading the dictionary database on first run.
package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/licht1stein/sanskrit-upaya/pkg/paths"
)

const (
	// DatabaseURL is the URL to download the dictionary database from.
	DatabaseURL = "https://sanskrit.myke.blog/dict.db"

	// AppSecret is sent as a header to authenticate the download.
	// This prevents casual hotlinking but isn't meant to be secure.
	AppSecret = "mitra-2024-sanskrit-app"

	// HeaderName is the custom header name for authentication.
	HeaderName = "X-Sanskrit-Mitra"

	// ExpectedChecksum is the SHA256 checksum of the database file.
	// Update this when the database is updated on the server.
	ExpectedChecksum = "2eeb4a92e8da19b24e4889ce15d57ea9b41ec4db29327116a9dd3e614c547b34"
)

// ProgressFunc is called during download with bytes downloaded and total size.
type ProgressFunc func(downloaded, total int64)

// GetDatabasePath returns the path where the database should be stored.
// Deprecated: Use paths.GetDatabasePath() directly. Kept for backward compatibility.
func GetDatabasePath() (string, error) {
	return paths.GetDatabasePath()
}

// DatabaseStatus represents the state of the local database.
type DatabaseStatus int

const (
	DatabaseMissing DatabaseStatus = iota
	DatabaseValid
	DatabaseNeedsUpdate // Checksum mismatch - corrupted or new version available
)

// CheckDatabase checks the database file status.
func CheckDatabase() DatabaseStatus {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return DatabaseMissing
	}
	_, err = os.Stat(dbPath)
	if err != nil {
		return DatabaseMissing
	}

	// If no expected checksum is set, just check file exists
	if ExpectedChecksum == "" {
		return DatabaseValid
	}

	// Verify checksum
	checksum, err := computeFileChecksum(dbPath)
	if err != nil {
		return DatabaseNeedsUpdate
	}
	if checksum != ExpectedChecksum {
		return DatabaseNeedsUpdate
	}
	return DatabaseValid
}

// computeFileChecksum calculates SHA256 checksum of a file.
func computeFileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Download downloads the database with progress reporting.
// Returns the SHA256 checksum of the downloaded file.
func Download(progress ProgressFunc) error {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return fmt.Errorf("get database path: %w", err)
	}

	// Create temporary file for download
	tmpPath := dbPath + ".tmp"

	// Create HTTP request with custom header
	req, err := http.NewRequest("GET", DatabaseURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set(HeaderName, AppSecret)

	// Perform request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength

	// Create output file
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	// Download with progress, computing checksum as we go
	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer
	hasher := sha256.New()

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				out.Close()
				os.Remove(tmpPath)
				return fmt.Errorf("write file: %w", writeErr)
			}
			hasher.Write(buf[:n])
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, totalSize)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			out.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("read response: %w", err)
		}
	}

	// Close file before verifying and renaming
	out.Close()

	// Verify checksum if expected checksum is set
	checksum := hex.EncodeToString(hasher.Sum(nil))
	if ExpectedChecksum != "" && checksum != ExpectedChecksum {
		os.Remove(tmpPath)
		return fmt.Errorf("checksum mismatch: expected %s, got %s", ExpectedChecksum, checksum)
	}

	// Move temp file to final location
	if err := os.Rename(tmpPath, dbPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename file: %w", err)
	}

	return nil
}
