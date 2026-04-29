package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/flowexec/flow/internal/version"
	"github.com/flowexec/flow/pkg/store"
)

const (
	cacheKey      = "updater:latest-release"
	cacheTTL      = 24 * time.Hour
	githubRepo    = "flowexec/flow"
	noCheckEnvVar = "FLOW_NO_UPDATE_CHECK"
)

// These vars are package-level so tests can override them without hitting
// the real GitHub API.
var (
	githubReleaseURL        = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	githubReleaseTagBaseURL = "https://api.github.com/repos/" + githubRepo + "/releases/tags"

	// currentSemVer is a var so tests can inject a known version.
	currentSemVer = version.SemVer
)

// ReleaseInfo holds metadata about a GitHub release fetched from the API.
type ReleaseInfo struct {
	TagName   string    `json:"tag_name"`
	HTMLURL   string    `json:"html_url"`
	CheckedAt time.Time `json:"checked_at"`
}

// CheckInBackground fires a background goroutine to fetch the latest release
// from GitHub and cache it. It is a no-op when:
//   - enabled is false (controlled by the updateCheck config field)
//   - FLOW_NO_UPDATE_CHECK is set (CI / offline override)
//   - the current binary has no known version (dev build)
//   - the cached result is still fresh (< 24 h old)
//
// Safe to call on every command invocation.
func CheckInBackground(ds store.DataStore, enabled bool) {
	if !enabled || IsDisabled() || currentSemVer() == "" {
		return
	}
	if data, err := ds.GetCacheEntry(cacheKey); err == nil && len(data) > 0 {
		var info ReleaseInfo
		if json.Unmarshal(data, &info) == nil && time.Since(info.CheckedAt) < cacheTTL {
			return // still fresh, skip network round-trip
		}
	}
	go func() {
		_ = checkAndCache(ds)
	}()
}

// CachedUpdateNotice returns a human-readable notice when a newer version is
// recorded in the cache. Returns "" when up to date, unknown, or disabled.
func CachedUpdateNotice(ds store.DataStore) string {
	if IsDisabled() || currentSemVer() == "" {
		return ""
	}
	data, err := ds.GetCacheEntry(cacheKey)
	if err != nil || len(data) == 0 {
		return ""
	}
	var info ReleaseInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return ""
	}
	if !IsNewer(currentSemVer(), info.TagName) {
		return ""
	}
	return fmt.Sprintf(
		"A new version of flow is available: %s → %s\nRun `flow cli update` to upgrade.",
		currentSemVer(), info.TagName,
	)
}

// LatestRelease fetches the latest release info from GitHub (live, not cached).
func LatestRelease() (*ReleaseInfo, error) {
	return fetchLatestRelease()
}

// ReleaseByTag fetches a specific release by its tag (e.g. "v2.1.0").
// A leading "v" is added automatically if omitted.
func ReleaseByTag(tag string) (*ReleaseInfo, error) {
	if tag != "" && tag[0] != 'v' {
		tag = "v" + tag
	}
	return fetchRelease(githubReleaseTagBaseURL + "/" + tag)
}

// RefreshCache stores release info in the data store with the current timestamp.
func RefreshCache(ds store.DataStore, info *ReleaseInfo) error {
	info.CheckedAt = time.Now()
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal release info: %w", err)
	}
	return ds.SetCacheEntry(cacheKey, data)
}

// IsNewer returns true if latestTag is a higher semver than currentSemVer.
// latestTag may include a leading "v" (e.g. "v2.1.0").
func IsNewer(current, latestTag string) bool {
	c, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	l, err := semver.NewVersion(latestTag)
	if err != nil {
		return false
	}
	return l.GreaterThan(c)
}

// IsDisabled returns true when update checks are suppressed via
// FLOW_NO_UPDATE_CHECK (any non-empty value other than "0" or "false").
func IsDisabled() bool {
	val := os.Getenv(noCheckEnvVar)
	return val != "" && val != "0" && val != "false"
}

func checkAndCache(ds store.DataStore) error {
	info, err := fetchLatestRelease()
	if err != nil {
		return err
	}
	return RefreshCache(ds, info)
}

func fetchLatestRelease() (*ReleaseInfo, error) {
	return fetchRelease(githubReleaseURL)
}

func fetchRelease(url string) (*ReleaseInfo, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release response: %w", err)
	}
	return &release, nil
}
