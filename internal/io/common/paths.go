package common

import (
	"os"
	"path/filepath"
	"strings"
)

const maxPathSegments = 5

// ShortenPath returns a display-friendly version of p:
// the home directory prefix is replaced with ~/, and if the result still has
// more than maxPathSegments slash-separated components the last four are kept
// preceded by …/.
func ShortenPath(p string) string {
	if p == "" {
		return p
	}

	slashP := filepath.ToSlash(p)
	if home, err := os.UserHomeDir(); err == nil {
		slashHome := filepath.ToSlash(home)
		if strings.HasPrefix(slashP, slashHome+"/") {
			slashP = "~/" + slashP[len(slashHome)+1:]
		}
	}

	parts := strings.Split(slashP, "/")
	if len(parts) <= maxPathSegments {
		return slashP
	}
	return "…/" + strings.Join(parts[len(parts)-4:], "/")
}
