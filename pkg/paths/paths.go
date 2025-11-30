// Package paths provides utilities for locating application data directories and files.
package paths

import (
	"os"
	"path/filepath"
)

// GetDataDir returns the XDG data directory for the app.
// It checks XDG_DATA_HOME environment variable first, falling back to ~/.local/share.
// The directory is created if it doesn't exist.
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

// GetDatabasePath returns the path where the main dictionary database should be stored.
func GetDatabasePath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "sanskrit.db"), nil
}

// GetStatePath returns the path where the user state database should be stored.
func GetStatePath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "state.db"), nil
}
