// Package version provides version checking against GitHub releases.
package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// GitHubRepo is the repository to check for releases.
	GitHubRepo = "licht1stein/sanskrit-upaya"
	// CheckTimeout is the HTTP timeout for version checks.
	CheckTimeout = 10 * time.Second
)

// Release represents a GitHub release.
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckResult contains the result of a version check.
type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	UpdateURL      string
	HasUpdate      bool
}

// Check compares the current version against the latest GitHub release.
// Returns nil if the check fails (network error, etc.) - callers should
// treat nil as "unknown" rather than "no update".
func Check(currentVersion string) *CheckResult {
	// Skip check for dev builds
	if currentVersion == "dev" || currentVersion == "" {
		return nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

	client := &http.Client{Timeout: CheckTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil
	}

	result := &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		UpdateURL:      release.HTMLURL,
		HasUpdate:      isNewer(release.TagName, currentVersion),
	}

	return result
}

// isNewer returns true if latest is newer than current.
// Versions are expected in format "v1.2.3" (BreakVer).
func isNewer(latest, current string) bool {
	// Strip 'v' prefix if present
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// Parse version parts
	latestParts := parseVersion(latest)
	currentParts := parseVersion(current)

	// Compare each part
	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}

	// If latest has more parts, it's newer (e.g., 1.0.1 > 1.0)
	return len(latestParts) > len(currentParts)
}

// parseVersion splits a version string into numeric parts.
func parseVersion(v string) []int {
	parts := strings.Split(v, ".")
	result := make([]int, 0, len(parts))

	for _, p := range parts {
		var num int
		fmt.Sscanf(p, "%d", &num)
		result = append(result, num)
	}

	return result
}
