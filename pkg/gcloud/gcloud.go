// Package gcloud provides utilities for Google Cloud CLI operations.
package gcloud

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const ocrProjectIDPrefix = "sanskrit-upaya-ocr"

var (
	resolvedGcloudPath string
	resolveOnce        sync.Once
)

// knownGcloudLocations returns platform-specific paths where gcloud may be installed.
func knownGcloudLocations() []string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		var paths []string
		if home != "" {
			paths = append(paths, filepath.Join(home, "google-cloud-sdk", "bin", "gcloud"))
		}
		paths = append(paths,
			"/opt/homebrew/bin/gcloud",
			"/opt/homebrew/share/google-cloud-sdk/bin/gcloud",
			"/usr/local/bin/gcloud",
			"/usr/local/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/bin/gcloud",
			"/opt/google-cloud-sdk/bin/gcloud",
		)
		return paths
	case "linux":
		var paths []string
		if home != "" {
			paths = append(paths, filepath.Join(home, "google-cloud-sdk", "bin", "gcloud"))
		}
		paths = append(paths,
			"/usr/bin/gcloud",
			"/usr/local/bin/gcloud",
			"/snap/bin/gcloud",
			"/opt/google-cloud-sdk/bin/gcloud",
		)
		return paths
	case "windows":
		var paths []string
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			paths = append(paths,
				filepath.Join(localAppData, "Google", "Cloud SDK", "google-cloud-sdk", "bin", "gcloud.cmd"),
			)
		}
		paths = append(paths, `C:\Program Files (x86)\Google\Cloud SDK\google-cloud-sdk\bin\gcloud.cmd`)
		return paths
	default:
		return nil
	}
}

// shellLookupPATH attempts to get the user's full shell PATH on macOS/Linux.
// GUI apps on macOS inherit a minimal PATH that excludes user-installed tools.
// This runs a login shell to discover the real PATH.
func shellLookupPATH() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	// Try the user's default shell first, then fall back to common shells
	shells := []string{}
	if s := os.Getenv("SHELL"); s != "" {
		shells = append(shells, s)
	}
	shells = append(shells, "/bin/zsh", "/bin/bash")

	for _, shell := range shells {
		if _, err := os.Stat(shell); err != nil {
			continue
		}
		cmd := exec.Command(shell, "-l", "-c", "echo $PATH")
		cmd.Env = []string{
			"HOME=" + os.Getenv("HOME"),
			"USER=" + os.Getenv("USER"),
			"PATH=" + os.Getenv("PATH"),
		}
		out, err := cmd.Output()
		if err == nil {
			if p := strings.TrimSpace(string(out)); p != "" {
				return p
			}
		}
	}
	return ""
}

// resolveGcloudBinary finds the gcloud binary using multiple strategies:
//  1. Standard PATH lookup (works when launched from terminal)
//  2. Probe known installation locations (works when launched from Finder/Dock on macOS)
//  3. Shell PATH discovery — run a login shell to get the full user PATH, then retry
func resolveGcloudBinary() string {
	// Strategy 1: standard PATH lookup
	if p, err := exec.LookPath("gcloud"); err == nil {
		return p
	}

	// Strategy 2: probe known installation locations
	for _, candidate := range knownGcloudLocations() {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}

	// Strategy 3: discover full PATH from login shell (macOS GUI apps get minimal PATH)
	if shellPATH := shellLookupPATH(); shellPATH != "" {
		// Temporarily override PATH for lookup
		origPATH := os.Getenv("PATH")
		os.Setenv("PATH", shellPATH)
		p, err := exec.LookPath("gcloud")
		os.Setenv("PATH", origPATH)
		if err == nil {
			return p
		}
	}

	return ""
}

// GcloudPath returns the resolved absolute path to the gcloud binary.
// The result is cached after the first call. Returns "" if gcloud is not found.
func GcloudPath() string {
	resolveOnce.Do(func() {
		resolvedGcloudPath = resolveGcloudBinary()
	})
	return resolvedGcloudPath
}

// gcloudCommand creates an exec.Cmd for a gcloud invocation using the resolved binary path.
// The parent directory of gcloud is added to the command's PATH so that gcloud's
// own sub-processes (e.g. Python, bundled tools) can be found.
func gcloudCommand(args ...string) *exec.Cmd {
	return gcloudCommandContext(context.Background(), args...)
}

// queryTimeout bounds non-interactive gcloud status queries so a hung CLI
// (e.g. network partition) cannot block the caller indefinitely.
const queryTimeout = 30 * time.Second

// gcloudCommandContext builds a gcloud command bound to ctx. Interactive
// commands should pass context.Background(); status queries should pass a
// context with a timeout (see runQuery).
func gcloudCommandContext(ctx context.Context, args ...string) *exec.Cmd {
	gcloudBin := GcloudPath()
	if gcloudBin == "" {
		// Fall back to bare name — will fail, but gives a clear error
		gcloudBin = "gcloud"
	}

	cmd := exec.CommandContext(ctx, gcloudBin, args...)

	// Ensure the directory containing gcloud is on PATH for child processes
	gcloudDir := filepath.Dir(gcloudBin)
	currentPATH := os.Getenv("PATH")
	if !strings.Contains(currentPATH, gcloudDir) {
		newPATH := gcloudDir + string(os.PathListSeparator) + currentPATH
		env := os.Environ()
		updated := make([]string, 0, len(env))
		for _, e := range env {
			if strings.HasPrefix(e, "PATH=") {
				continue // skip old PATH, we'll add the new one
			}
			updated = append(updated, e)
		}
		updated = append(updated, "PATH="+newPATH)
		cmd.Env = updated
	}

	return cmd
}

// IsInstalled checks if gcloud CLI is available.
// On macOS, this probes known installation paths beyond the inherited PATH,
// so it works even when the app is launched from Finder/Dock.
func IsInstalled() bool {
	return GcloudPath() != ""
}

// IsAuthenticated checks if gcloud CLI has an active account.
func IsAuthenticated() bool {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	cmd := gcloudCommandContext(ctx, "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) != ""
}

// HasApplicationDefaultCredentials checks if ADC credentials exist.
func HasApplicationDefaultCredentials() bool {
	// Check GOOGLE_APPLICATION_CREDENTIALS env var
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		return true
	}
	// Check default location
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	credPath := filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
	_, err = os.Stat(credPath)
	return err == nil
}

// ProjectExists checks if a GCP project exists.
func ProjectExists(projectID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	cmd := gcloudCommandContext(ctx, "projects", "describe", projectID, "--format=value(projectId)")
	cmd.Stderr = nil
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == projectID
}

// GetOCRProjectConfigPath returns the path to the file storing the user's OCR project ID.
func GetOCRProjectConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "sanskrit-upaya")
	return filepath.Join(configDir, "ocr-project-id"), nil
}

// GetOrCreateOCRProjectID returns the user's OCR project ID, creating one if needed.
// The project ID is stored in ~/.config/sanskrit-upaya/ocr-project-id
func GetOrCreateOCRProjectID() (string, error) {
	configPath, err := GetOCRProjectConfigPath()
	if err != nil {
		return "", err
	}

	// Check if we already have a project ID stored
	if data, err := os.ReadFile(configPath); err == nil {
		projectID := strings.TrimSpace(string(data))
		if projectID != "" {
			return projectID, nil
		}
	}

	// Generate a new unique project ID with random suffix
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random suffix: %w", err)
	}
	suffix := hex.EncodeToString(randomBytes)
	projectID := fmt.Sprintf("%s-%s", ocrProjectIDPrefix, suffix)

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save the project ID
	if err := os.WriteFile(configPath, []byte(projectID+"\n"), 0644); err != nil {
		return "", fmt.Errorf("failed to save project ID: %w", err)
	}

	return projectID, nil
}

// RunCommand runs a gcloud command with output visible to user.
// Returns true if the command succeeded.
func RunCommand(args ...string) bool {
	cmd := gcloudCommand(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	return err == nil
}

// RunCommandWithOutput runs a gcloud command and streams output to the provided writers.
// Returns true if the command succeeded.
func RunCommandWithOutput(stdout, stderr io.Writer, args ...string) bool {
	cmd := gcloudCommand(args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	return err == nil
}

// RunCommandAsync runs a gcloud command asynchronously, streaming output line by line.
// The onLine callback is called for each line of output (both stdout and stderr).
// Returns a channel that receives true on success, false on failure.
func RunCommandAsync(onLine func(line string), args ...string) <-chan bool {
	result := make(chan bool, 1)

	go func() {
		cmd := gcloudCommand(args...)

		// Create pipes for stdout and stderr
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			onLine(fmt.Sprintf("Error creating stdout pipe: %v", err))
			result <- false
			return
		}
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			onLine(fmt.Sprintf("Error creating stderr pipe: %v", err))
			result <- false
			return
		}

		if err := cmd.Start(); err != nil {
			onLine(fmt.Sprintf("Error starting command: %v", err))
			result <- false
			return
		}

		// Read stdout and stderr concurrently
		done := make(chan struct{}, 2)

		readPipe := func(pipe io.ReadCloser) {
			scanner := bufio.NewScanner(pipe)
			for scanner.Scan() {
				onLine(scanner.Text())
			}
			done <- struct{}{}
		}

		go readPipe(stdoutPipe)
		go readPipe(stderrPipe)

		// Wait for both pipes to be read
		<-done
		<-done

		err = cmd.Wait()
		result <- err == nil
	}()

	return result
}

// PromptYesNo asks user a yes/no question via stdin.
func PromptYesNo(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [Y/n]: ", question)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

// OpenBrowser opens a URL in the default browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// GetInstallURL returns the platform-specific gcloud installation URL.
func GetInstallURL() string {
	return "https://cloud.google.com/sdk/docs/install"
}

// GetBillingURL returns the billing enablement URL for a project.
func GetBillingURL(projectID string) string {
	return "https://console.developers.google.com/billing/enable?project=" + projectID
}

// GetConsoleURL returns the Google Cloud Console URL.
func GetConsoleURL() string {
	return "https://console.cloud.google.com"
}
