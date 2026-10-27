// Package download handles downloading the dictionary database on first run.
package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/licht1stein/sanskrit-upaya/pkg/paths"
)

const (
	// DatabaseURL is the URL to download the dictionary database from.
	DatabaseURL = "https://sanskrit.myke.blog/dict.db"

	// AppSecret is sent as a header with the download request.
	// NOTE: This is NOT a security mechanism. It is compiled into every
	// binary and trivially recoverable (e.g. `strings`), so it only deters
	// casual hotlinking. Do not rely on it for access control — enforce that
	// server-side (rate limiting, referrer checks, or per-user tokens).
	AppSecret = "mitra-2024-sanskrit-app"

	// HeaderName is the custom header name for authentication.
	HeaderName = "X-Sanskrit-Mitra"

	// ExpectedChecksum is the SHA256 checksum of the database file.
	// Update this (and ExpectedSize) when the database is updated on the server.
	ExpectedChecksum = "482eee2da1bfce8cc17c58b0fd2c92357e774ef78ca560d827fa8e6062b0f6e9"

	// ExpectedSize is the size in bytes of the published database file.
	ExpectedSize int64 = 786153472
)

// LegacyChecksums lists checksums of previously published databases whose
// content is equivalent to the current one. A local file matching one of
// these is accepted as-is instead of forcing a re-download; search.Open adds
// any missing indexes on first open.
//
// When a new database with different content is published, clear this list
// so existing installs pick up the new file.
var LegacyChecksums = []string{
	// Original database (same data, without idx_words_lower_* indexes).
	"2eeb4a92e8da19b24e4889ce15d57ea9b41ec4db29327116a9dd3e614c547b34",
}

// expectedChecksum and expectedSize are the values Download verifies
// against. They are variables only so tests can substitute a small fixture.
var (
	expectedChecksum = ExpectedChecksum
	expectedSize     = ExpectedSize
)

// isAcceptedChecksum reports whether checksum identifies a database this
// build can use without re-downloading.
func isAcceptedChecksum(checksum string) bool {
	if checksum == "" {
		return false
	}
	if checksum == expectedChecksum {
		return true
	}
	for _, c := range LegacyChecksums {
		if checksum == c {
			return true
		}
	}
	return false
}

// verifiedMarkerPath returns the path of the sidecar file recording which
// published checksum the local database was verified against.
//
// The database is verified by SHA256 once (after download, or on first
// check of an existing file). The app then modifies the file in place:
// search.Open switches it to WAL journal mode (rewriting header bytes) and
// may create missing indexes. Re-hashing on every startup would therefore
// report a mismatch and re-download the database on every launch. The
// marker records the successful verification so later startups do not
// depend on the file staying byte-identical.
func verifiedMarkerPath(dbPath string) string {
	return dbPath + ".verified"
}

// readVerifiedMarker returns the checksum recorded in the marker, or "".
func readVerifiedMarker(dbPath string) string {
	b, err := os.ReadFile(verifiedMarkerPath(dbPath))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// writeVerifiedMarker records that dbPath was verified against checksum.
func writeVerifiedMarker(dbPath, checksum string) error {
	return os.WriteFile(verifiedMarkerPath(dbPath), []byte(checksum+"\n"), 0644)
}

// sqliteMagic is the 16-byte header every SQLite database file starts with.
const sqliteMagic = "SQLite format 3\x00"

// looksLikeSQLite does a cheap sanity check that path is a SQLite database.
func looksLikeSQLite(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, len(sqliteMagic))
	if _, err := io.ReadFull(f, buf); err != nil {
		return false
	}
	return string(buf) == sqliteMagic
}

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
//
// A database is valid if it was previously verified against an accepted
// checksum (see verifiedMarkerPath) and still looks like a SQLite file, or
// if its SHA256 matches an accepted checksum now (in which case the
// verification is recorded so it is not repeated on every startup).
func CheckDatabase() DatabaseStatus {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return DatabaseMissing
	}
	return checkDatabaseAt(dbPath)
}

func checkDatabaseAt(dbPath string) DatabaseStatus {
	if _, err := os.Stat(dbPath); err != nil {
		return DatabaseMissing
	}

	// If no expected checksum is set, just check file exists
	if expectedChecksum == "" {
		return DatabaseValid
	}

	// Already verified (the file may since have been modified in place by
	// search.Open, which is expected).
	if isAcceptedChecksum(readVerifiedMarker(dbPath)) {
		if looksLikeSQLite(dbPath) {
			return DatabaseValid
		}
		return DatabaseNeedsUpdate
	}

	// Verify checksum
	checksum, err := computeFileChecksum(dbPath)
	if err != nil {
		return DatabaseNeedsUpdate
	}
	if !isAcceptedChecksum(checksum) {
		return DatabaseNeedsUpdate
	}
	if err := writeVerifiedMarker(dbPath, checksum); err != nil {
		log.Printf("Warning: could not record database verification: %v", err)
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
	return downloadTo(DatabaseURL, dbPath, progress)
}

func downloadTo(url, dbPath string, progress ProgressFunc) error {
	// Create temporary file for download
	tmpPath := dbPath + ".tmp"

	// Create HTTP request with custom header
	req, err := http.NewRequest("GET", url, nil)
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
	if expectedSize > 0 && totalSize > 0 && totalSize != expectedSize {
		return fmt.Errorf("unexpected database size: expected %d bytes, server reports %d", expectedSize, totalSize)
	}

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

	if expectedSize > 0 && downloaded != expectedSize {
		os.Remove(tmpPath)
		return fmt.Errorf("incomplete download: expected %d bytes, got %d", expectedSize, downloaded)
	}

	// Verify checksum if expected checksum is set
	checksum := hex.EncodeToString(hasher.Sum(nil))
	if expectedChecksum != "" && checksum != expectedChecksum {
		os.Remove(tmpPath)
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, checksum)
	}

	// Drop any marker for the file being replaced before swapping it in.
	os.Remove(verifiedMarkerPath(dbPath))

	// Move temp file to final location
	if err := os.Rename(tmpPath, dbPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename file: %w", err)
	}

	// A WAL file left over from the replaced database must not be applied
	// to the new one.
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	if expectedChecksum != "" {
		if err := writeVerifiedMarker(dbPath, checksum); err != nil {
			log.Printf("Warning: could not record database verification: %v", err)
		}
	}

	return nil
}
