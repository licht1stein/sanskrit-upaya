// Package gcloud provides utilities for Google Cloud CLI operations.
package gcloud

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const ocrProjectIDPrefix = "sanskrit-upaya-ocr"

// IsInstalled checks if gcloud CLI is available in PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("gcloud")
	return err == nil
}

// IsAuthenticated checks if gcloud CLI has an active account.
func IsAuthenticated() bool {
	cmd := exec.Command("gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
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
	cmd := exec.Command("gcloud", "projects", "describe", projectID, "--format=value(projectId)")
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
	cmd := exec.Command("gcloud", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	return err == nil
}

// RunCommandWithOutput runs a gcloud command and streams output to the provided writers.
// Returns true if the command succeeded.
func RunCommandWithOutput(stdout, stderr io.Writer, args ...string) bool {
	cmd := exec.Command("gcloud", args...)
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
		cmd := exec.Command("gcloud", args...)

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
