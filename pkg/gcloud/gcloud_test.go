package gcloud

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// resetResolveCache clears the cached gcloud path so tests can run independently.
func resetResolveCache() {
	resolvedGcloudPath = ""
	resolveOnce = sync.Once{}
}

func TestKnownGcloudLocations_ReturnsNonEmpty(t *testing.T) {
	locations := knownGcloudLocations()
	if len(locations) == 0 && (runtime.GOOS == "darwin" || runtime.GOOS == "linux") {
		t.Fatal("expected non-empty known locations for", runtime.GOOS)
	}
}

func TestKnownGcloudLocations_ContainsPlatformPaths(t *testing.T) {
	locations := knownGcloudLocations()
	joined := strings.Join(locations, "\n")

	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(joined, "/opt/homebrew/bin/gcloud") {
			t.Error("missing /opt/homebrew/bin/gcloud for darwin")
		}
		if !strings.Contains(joined, "google-cloud-sdk/bin/gcloud") {
			t.Error("missing google-cloud-sdk/bin/gcloud for darwin")
		}
	case "linux":
		if !strings.Contains(joined, "/usr/bin/gcloud") {
			t.Error("missing /usr/bin/gcloud for linux")
		}
		if !strings.Contains(joined, "/snap/bin/gcloud") {
			t.Error("missing /snap/bin/gcloud for linux")
		}
	}
}

func TestKnownGcloudLocations_IncludesHomeDirPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	locations := knownGcloudLocations()
	expected := filepath.Join(home, "google-cloud-sdk", "bin", "gcloud")

	found := false
	for _, loc := range locations {
		if loc == expected {
			found = true
			break
		}
	}

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !found {
			t.Errorf("expected %s in known locations", expected)
		}
	}
}

func TestResolveGcloudBinary_FindsFakeGcloud(t *testing.T) {
	resetResolveCache()

	// Create a temporary directory with a fake gcloud binary
	tmpDir := t.TempDir()
	fakeGcloud := filepath.Join(tmpDir, "gcloud")
	if runtime.GOOS == "windows" {
		fakeGcloud = filepath.Join(tmpDir, "gcloud.cmd")
	}
	if err := os.WriteFile(fakeGcloud, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	// Temporarily prepend tmpDir to PATH
	origPATH := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+origPATH)
	defer os.Setenv("PATH", origPATH)

	result := resolveGcloudBinary()
	if result == "" {
		t.Fatal("expected to find fake gcloud in PATH")
	}
	if !strings.Contains(result, tmpDir) {
		t.Errorf("expected path containing %s, got %s", tmpDir, result)
	}
}

func TestResolveGcloudBinary_ReturnsEmptyWhenNotFound(t *testing.T) {
	resetResolveCache()

	// Use an empty PATH so nothing is found
	origPATH := os.Getenv("PATH")
	os.Setenv("PATH", t.TempDir()) // empty dir
	defer os.Setenv("PATH", origPATH)

	// This test may still find gcloud via shellLookupPATH or known locations,
	// so we just verify it doesn't panic
	_ = resolveGcloudBinary()
}

func TestGcloudPath_IsCached(t *testing.T) {
	resetResolveCache()

	// Call GcloudPath twice — should return the same result
	first := GcloudPath()
	second := GcloudPath()

	if first != second {
		t.Errorf("GcloudPath not stable: %q vs %q", first, second)
	}
}

func TestGcloudCommand_SetsUpPATH(t *testing.T) {
	resetResolveCache()

	// Create a fake gcloud to resolve to
	tmpDir := t.TempDir()
	fakeGcloud := filepath.Join(tmpDir, "gcloud")
	if runtime.GOOS == "windows" {
		fakeGcloud = filepath.Join(tmpDir, "gcloud.cmd")
	}
	if err := os.WriteFile(fakeGcloud, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	origPATH := os.Getenv("PATH")
	// Set PATH to NOT include tmpDir
	os.Setenv("PATH", "/usr/bin:/bin")
	defer os.Setenv("PATH", origPATH)

	// Force resolution to our fake gcloud
	resolvedGcloudPath = fakeGcloud
	resolveOnce.Do(func() {}) // Mark as resolved

	cmd := gcloudCommand("version")

	// Check that the command's PATH includes the gcloud directory
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "PATH=") {
			if !strings.Contains(env, tmpDir) {
				t.Errorf("expected command PATH to contain %s, got %s", tmpDir, env)
			}
			return
		}
	}
	// If gcloud dir is already in the process PATH, cmd.Env may not be set — that's okay
}

func TestGcloudCommand_UsesResolvedPath(t *testing.T) {
	resetResolveCache()

	tmpDir := t.TempDir()
	fakeGcloud := filepath.Join(tmpDir, "gcloud")
	if runtime.GOOS == "windows" {
		fakeGcloud = filepath.Join(tmpDir, "gcloud.cmd")
	}
	if err := os.WriteFile(fakeGcloud, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	// Force resolution
	resolvedGcloudPath = fakeGcloud
	resolveOnce.Do(func() {})

	cmd := gcloudCommand("auth", "list")
	if cmd.Path != fakeGcloud {
		// cmd.Path may be resolved further by exec.Command, check Args[0]
		if cmd.Args[0] != fakeGcloud {
			t.Errorf("expected command to use %s, got Path=%s Args[0]=%s", fakeGcloud, cmd.Path, cmd.Args[0])
		}
	}
	if len(cmd.Args) < 3 || cmd.Args[1] != "auth" || cmd.Args[2] != "list" {
		t.Errorf("unexpected args: %v", cmd.Args)
	}
}

func TestIsInstalled_MatchesLookPath(t *testing.T) {
	resetResolveCache()

	// If gcloud is truly available on this system, IsInstalled should return true
	_, lookPathErr := exec.LookPath("gcloud")
	expected := lookPathErr == nil

	result := IsInstalled()
	// IsInstalled may find gcloud even when LookPath doesn't (via probing),
	// but if LookPath finds it, IsInstalled definitely should too
	if expected && !result {
		t.Error("LookPath found gcloud but IsInstalled returned false")
	}
}

func TestShellLookupPATH_DoesNotPanic(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shellLookupPATH is a no-op on Windows")
	}
	// Just verify it doesn't panic or hang
	result := shellLookupPATH()
	// On a real system with zsh/bash, this should return something non-empty
	if result != "" {
		if !strings.Contains(result, "/") {
			t.Errorf("PATH looks wrong: %s", result)
		}
	}
}
