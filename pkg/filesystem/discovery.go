package filesystem

import (
	"path/filepath"
	"runtime"
	"strings"
)

// MaxWalkUpDepth bounds the ancestor walk. The volume root normally terminates it; this is a
// backstop against pathological paths (deep symlink cycles resolved into a path, mount loops).
const MaxWalkUpDepth = 128

// NormalizePath cleans a path and, on macOS, strips the "/private" prefix. Paths under /tmp,
// /var, and /etc are symlinks into /private there, and the OS hands back either form depending
// on how it was derived — os.Getwd() returns the /private form while a path the user typed
// usually doesn't. Comparisons must normalize both sides or the same directory compares unequal
// to itself.
func NormalizePath(p string) string {
	if p == "" {
		return ""
	}
	p = filepath.Clean(p)
	if runtime.GOOS == "darwin" {
		if p == "/private" {
			return "/"
		}
		if strings.HasPrefix(p, "/private/") {
			p = strings.TrimPrefix(p, "/private")
		}
	}
	return p
}

// SamePath reports whether two paths refer to the same directory. It compares the normalized
// forms first, then the symlink-resolved forms, so a workspace registered through a symlink
// still matches the real path a walk-up produced.
func SamePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if NormalizePath(a) == NormalizePath(b) {
		return true
	}
	ra, err := filepath.EvalSymlinks(a)
	if err != nil {
		return false
	}
	rb, err := filepath.EvalSymlinks(b)
	if err != nil {
		return false
	}
	return NormalizePath(ra) == NormalizePath(rb)
}

// IsPathWithin reports whether path is base or lives underneath it. Unlike a raw string prefix
// check, "/a/wsX" is not within "/a/ws".
func IsPathWithin(path, base string) bool {
	if path == "" || base == "" {
		return false
	}
	rel, err := filepath.Rel(NormalizePath(base), NormalizePath(path))
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}

// FindWorkspaceRoot walks up from startDir and returns the first directory containing a
// flow.yaml. This is how a workspace is located without consulting the user config at all —
// the same way make and bazel find their root.
func FindWorkspaceRoot(startDir string) (string, bool) {
	return FindWorkspaceRootExcluding(startDir, nil)
}

// FindWorkspaceRootExcluding is FindWorkspaceRoot but continues walking upward past any
// candidate for which skip returns true. Used to keep a flow.yaml that sits inside a tree an
// ancestor workspace already excludes (vendor/, node_modules/, ...) from becoming a root of
// its own.
func FindWorkspaceRootExcluding(startDir string, skip func(root string) bool) (string, bool) {
	if startDir == "" {
		return "", false
	}
	dir := NormalizePath(startDir)
	if !filepath.IsAbs(dir) {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return "", false
		}
		dir = NormalizePath(abs)
	}

	for range MaxWalkUpDepth {
		if WorkspaceConfigExists(dir) && (skip == nil || !skip(dir)) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}
