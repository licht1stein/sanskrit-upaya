package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDataDir(t *testing.T) {
	// Test with XDG_DATA_HOME set
	t.Run("with XDG_DATA_HOME", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_DATA_HOME", tmpDir)
		defer os.Unsetenv("XDG_DATA_HOME")

		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir() error = %v", err)
		}

		expected := filepath.Join(tmpDir, "sanskrit-dictionary")
		if dir != expected {
			t.Errorf("GetDataDir() = %v, want %v", dir, expected)
		}

		// Check directory was created
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory was not created: %v", dir)
		}
	})

	// Test without XDG_DATA_HOME (uses ~/.local/share)
	t.Run("without XDG_DATA_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_DATA_HOME")

		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir() error = %v", err)
		}

		// Should contain .local/share/sanskrit-dictionary
		if !strings.Contains(dir, ".local/share/sanskrit-dictionary") &&
			!strings.Contains(dir, "sanskrit-dictionary") {
			t.Errorf("GetDataDir() = %v, expected to contain sanskrit-dictionary", dir)
		}
	})
}

func TestGetDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	path, err := GetDatabasePath()
	if err != nil {
		t.Fatalf("GetDatabasePath() error = %v", err)
	}

	if !strings.HasSuffix(path, "sanskrit.db") {
		t.Errorf("GetDatabasePath() = %v, want suffix sanskrit.db", path)
	}

	expected := filepath.Join(tmpDir, "sanskrit-dictionary", "sanskrit.db")
	if path != expected {
		t.Errorf("GetDatabasePath() = %v, want %v", path, expected)
	}
}

func TestGetStatePath(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	path, err := GetStatePath()
	if err != nil {
		t.Fatalf("GetStatePath() error = %v", err)
	}

	if !strings.HasSuffix(path, "state.db") {
		t.Errorf("GetStatePath() = %v, want suffix state.db", path)
	}

	expected := filepath.Join(tmpDir, "sanskrit-dictionary", "state.db")
	if path != expected {
		t.Errorf("GetStatePath() = %v, want %v", path, expected)
	}
}

func TestPathsConsistency(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	dataDir, _ := GetDataDir()
	dbPath, _ := GetDatabasePath()
	statePath, _ := GetStatePath()

	// All paths should be under dataDir
	if !strings.HasPrefix(dbPath, dataDir) {
		t.Errorf("DatabasePath %v not under DataDir %v", dbPath, dataDir)
	}
	if !strings.HasPrefix(statePath, dataDir) {
		t.Errorf("StatePath %v not under DataDir %v", statePath, dataDir)
	}
}
