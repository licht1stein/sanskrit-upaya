package download

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestGetDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	path, err := GetDatabasePath()
	if err != nil {
		t.Fatalf("GetDatabasePath() error = %v", err)
	}

	expected := filepath.Join(tmpDir, "sanskrit-dictionary", "sanskrit.db")
	if path != expected {
		t.Errorf("GetDatabasePath() = %v, want %v", path, expected)
	}
}

func TestCheckDatabase_Missing(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	status := CheckDatabase()
	if status != DatabaseMissing {
		t.Errorf("CheckDatabase() = %v, want DatabaseMissing", status)
	}
}

func TestCheckDatabase_NeedsUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	// Create directory and fake database file with wrong checksum
	dbDir := filepath.Join(tmpDir, "sanskrit-dictionary")
	os.MkdirAll(dbDir, 0755)
	dbPath := filepath.Join(dbDir, "sanskrit.db")
	os.WriteFile(dbPath, []byte("fake database content"), 0644)

	status := CheckDatabase()
	if status != DatabaseNeedsUpdate {
		t.Errorf("CheckDatabase() = %v, want DatabaseNeedsUpdate", status)
	}
}

func TestDatabaseExists(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	// Initially doesn't exist
	if DatabaseExists() {
		t.Error("DatabaseExists() = true, want false (no file)")
	}

	// Create fake file (wrong checksum)
	dbDir := filepath.Join(tmpDir, "sanskrit-dictionary")
	os.MkdirAll(dbDir, 0755)
	dbPath := filepath.Join(dbDir, "sanskrit.db")
	os.WriteFile(dbPath, []byte("fake"), 0644)

	// Still false because checksum doesn't match
	if DatabaseExists() {
		t.Error("DatabaseExists() = true, want false (wrong checksum)")
	}
}

func TestGetDatabaseChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	// Create directory and test file
	dbDir := filepath.Join(tmpDir, "sanskrit-dictionary")
	os.MkdirAll(dbDir, 0755)
	dbPath := filepath.Join(dbDir, "sanskrit.db")

	content := []byte("test database content")
	os.WriteFile(dbPath, content, 0644)

	// Calculate expected checksum
	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	checksum, err := GetDatabaseChecksum()
	if err != nil {
		t.Fatalf("GetDatabaseChecksum() error = %v", err)
	}

	if checksum != expected {
		t.Errorf("GetDatabaseChecksum() = %v, want %v", checksum, expected)
	}
}

func TestGetDatabaseChecksum_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	_, err := GetDatabaseChecksum()
	if err == nil {
		t.Error("GetDatabaseChecksum() expected error for non-existent file")
	}
}

func TestComputeFileChecksum(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	content := []byte("hello world")
	os.WriteFile(tmpFile, content, 0644)

	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	got, err := computeFileChecksum(tmpFile)
	if err != nil {
		t.Fatalf("computeFileChecksum() error = %v", err)
	}

	if got != expected {
		t.Errorf("computeFileChecksum() = %v, want %v", got, expected)
	}
}

func TestDatabaseStatusConstants(t *testing.T) {
	// Verify constants have expected values
	if DatabaseMissing != 0 {
		t.Errorf("DatabaseMissing = %d, want 0", DatabaseMissing)
	}
	if DatabaseValid != 1 {
		t.Errorf("DatabaseValid = %d, want 1", DatabaseValid)
	}
	if DatabaseNeedsUpdate != 2 {
		t.Errorf("DatabaseNeedsUpdate = %d, want 2", DatabaseNeedsUpdate)
	}
}
