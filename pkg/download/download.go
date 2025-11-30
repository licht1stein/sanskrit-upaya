// Package download handles downloading the dictionary database on first run.
package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// DatabaseURL is the URL to download the dictionary database from.
	DatabaseURL = "https://sanskrit.myke.blog/dict.db"

	// AppSecret is sent as a header to authenticate the download.
	// This prevents casual hotlinking but isn't meant to be secure.
	AppSecret = "mitra-2024-sanskrit-app"

	// HeaderName is the custom header name for authentication.
	HeaderName = "X-Sanskrit-Mitra"
)

// ProgressFunc is called during download with bytes downloaded and total size.
type ProgressFunc func(downloaded, total int64)

// GetDataDir returns the XDG data directory for the app.
func GetDataDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	appDir := filepath.Join(dataHome, "sanskrit-dictionary")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return appDir, nil
}

// GetDatabasePath returns the path where the database should be stored.
func GetDatabasePath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "sanskrit.db"), nil
}

// DatabaseExists checks if the database file exists.
func DatabaseExists() bool {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(dbPath)
	return err == nil
}

// Download downloads the database with progress reporting.
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

	// Download with progress
	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				os.Remove(tmpPath)
				return fmt.Errorf("write file: %w", writeErr)
			}
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, totalSize)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("read response: %w", err)
		}
	}

	// Close file before renaming
	out.Close()

	// Move temp file to final location
	if err := os.Rename(tmpPath, dbPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename file: %w", err)
	}

	return nil
}
