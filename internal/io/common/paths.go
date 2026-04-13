package common

import (
	"path/filepath"
	"strings"
)

// ShortenPath returns the last two directory components of a path prefixed
// with "…/". If the path has two or fewer components, it is returned as-is.
func ShortenPath(p string) string {
	parts := strings.Split(filepath.ToSlash(p), "/")
	if len(parts) <= 2 {
		return p
	}
	return "…/" + strings.Join(parts[len(parts)-2:], "/")
}
