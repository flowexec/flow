package version

import (
	// using blank import for embed as it is only used inside comments.
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var (
	// gitCommit returns the git commit that was compiled.
	gitCommit string

	// version returns the main version number that is being exec at the moment.
	version string

	// buildDate returns the date the binary was built
	buildDate string
)

const (
	unknown = "unknown"
	ghLatestReleaseURL = "https://api.github.com/repos/flowexec/flow/releases/latest"
)

// GoVersion returns the version of the go runtime used to compile the binary.
var goVersion = runtime.Version()

// OsArch returns the os and arch used to build the binary.
var osArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

// Plain returns just the semantic version string (e.g., v0.7.3) or "unknown" if unset.
func Plain() string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = unknown
	}
	return v
}

// generateOutput returns the detailed output of the version command, including an update notice if available.
func generateOutput() string {
	if gitCommit == "" {
		gitCommit = unknown
	}
	if version == "" {
		version = unknown
	}
	if buildDate == "" {
		buildDate = unknown
	}

	out := fmt.Sprintf(`

Version: %s
Git Commit: %s
Build date: %s
Go version: %s
OS / Arch : %s
`, strings.TrimSpace(version), strings.TrimSpace(gitCommit), strings.TrimSpace(buildDate), goVersion, osArch)

	// Attempt to check for updates, but fail silently and NEVER block for long.
	if strings.TrimSpace(version) != unknown {
		if newer, latest := newerVersionAvailable(Plain()); newer && latest != "" {
			out += fmt.Sprintf("\nWarning: a newer version of flow is available: %s\nSee: https://github.com/flowexec/flow/releases\n", latest)
		}
	}
	return out
}

// String returns the detailed version output.
func String() string {
	return generateOutput()
}

// newerVersionAvailable checks GitHub for the latest release and compares it to current.
// Returns (true, latest) if a newer version exists. All errors are swallowed and return (false, "").
func newerVersionAvailable(current string) (bool, string) {
	type ghResp struct {
		TagName string `json:"tag_name"`
	}
	client := &http.Client{Timeout: 1 * time.Second}
	req, err := http.NewRequest(http.MethodGet, ghLatestReleaseURL, nil)
	if err != nil {
		return false, ""
	}
	req.Header.Set("User-Agent", "flow/"+current)
	resp, err := client.Do(req)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, ""
	}
	var data ghResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, ""
	}
	latest := strings.TrimSpace(data.TagName)
	if latest == "" {
		return false, ""
	}
	// Normalize versions by stripping leading 'v' for comparison, but keep original latest for display.
	curNorm := strings.TrimPrefix(strings.TrimSpace(current), "v")
	latNorm := strings.TrimPrefix(latest, "v")
	// Perform a simple semver-ish comparison without extra deps.
	if compareSemver(curNorm, latNorm) < 0 {
		return true, latest
	}
	return false, ""
}

// compareSemver compares two semantic versions (major.minor.patch[-prerelease]) numerically.
// Returns -1 if a<b, 0 if equal, 1 if a>b. Very lenient: treats missing parts as 0 and
// considers any prerelease as lower than the corresponding release.
func compareSemver(a, b string) int {
	// separate prerelease
	as := strings.SplitN(a, "-", 2)
	bs := strings.SplitN(b, "-", 2)
	apr, bpr := "", ""
	if len(as) == 2 {
		apr = as[1]
	}
	if len(bs) == 2 {
		bpr = bs[1]
	}
	ap := strings.Split(as[0], ".")
	bp := strings.Split(bs[0], ".")
	for len(ap) < 3 {
		ap = append(ap, "0")
	}
	for len(bp) < 3 {
		bp = append(bp, "0")
	}
	for i := 0; i < 3; i++ {
		ai, bi := parseIntSafe(ap[i]), parseIntSafe(bp[i])
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
	}
	// equal core; handle prerelease
	if apr == "" && bpr == "" {
		return 0
	} else if apr == "" { // a is release, b is prerelease
		return 1
	} else if bpr == "" { // a is prerelease, b is release
		return -1
	}
	// both prerelease: fallback to lexical compare
	if apr < bpr {
		return -1
	} else if apr > bpr {
		return 1
	}
	return 0
}

func parseIntSafe(s string) int {
	var n int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
