package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/flowexec/flow/internal/logger"
)

// ExpandPath expands a general path to an absolute path with security validation.
// The following transformations are applied:
// - empty path -> fallbackPath (directory portion)
// - ~/path -> home directory + path
// - ./path or . -> current working directory + path
// - /path -> absolute path
// - relative path -> fallbackPath (directory portion) + path
// - ${envVar} -> expanded to the value from env map.
func ExpandPath(path, fallbackDir string, env map[string]string) string {
	var targetPath string
	switch {
	case path == "":
		targetPath = fallbackDir
	case path == "." || strings.HasPrefix(path, "./"):
		wd, err := os.Getwd()
		if err != nil {
			logger.Log().Warnx("unable to get working directory for relative path expansion", "err", err)
			targetPath = filepath.Join(fallbackDir, path)
		} else {
			targetPath = filepath.Join(wd, path[1:])
		}
	case strings.HasPrefix(path, "~/"):
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Log().Warnx("unable to get user home directory for relative path expansion", "err", err)
			targetPath = filepath.Join(fallbackDir, path)
		} else {
			targetPath = filepath.Join(homeDir, path[2:])
		}
	case strings.HasPrefix(path, "/"), strings.HasPrefix(path, "$"):
		targetPath = path
	default:
		targetPath = filepath.Join(fallbackDir, path)
	}

	targetPath = os.Expand(targetPath, func(key string) string {
		val, found := env[key]
		if !found {
			logger.Log().Warnx("unable to find env key in path expansion", "key", key)
		}
		return val
	})

	if err := validateSecurePath(targetPath); err != nil {
		logger.Log().Fatalx("path failed security validation", "path", targetPath, "err", err)
		return "" // Shouldn't get here with fatal logger, but just in case
	}

	return filepath.Clean(targetPath)
}

// ExpandDirectory expands the directory field of an executable to an absolute path.
// The following transformations are applied:
// - empty dir -> execPath (directory portion)
// - // -> wsPath + dir path (workspace-specific)
// - all other paths -> delegated to ExpandPath
// If the input contains a filename, returns just the directory portion.
func ExpandDirectory(dir, wsPath, execPath string, env map[string]string) string {
	var expandedPath string
	if wsPath != "" && strings.HasPrefix(dir, "//") {
		expandedPath = strings.Replace(dir, "//", wsPath+"/", 1)
	} else {
		expandedPath = ExpandPath(dir, filepath.Dir(execPath), env)
	}

	if ext := filepath.Ext(expandedPath); ext != "" && !isHiddenDir(expandedPath) {
		return filepath.Dir(expandedPath)
	}

	return expandedPath
}

func isHiddenDir(path string) bool {
	// Regex to match hidden directories like .config
	// This assumes that filenames starting with a dot (.) are hidden directories - this may not be universally true
	hiddenDirRegex := regexp.MustCompile(`(^|/)\.[^/]+$`)
	return hiddenDirRegex.MatchString(path)
}

// validateSecurePath checks if a path is safe to use
func validateSecurePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for directory traversal attempts
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	// Ensure the path is absolute after expansion
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Basic check that we're not accessing sensitive system directories
	systemDirs := []string{"/etc", "/sys", "/proc", "/dev"}
	for _, sysDir := range systemDirs {
		if strings.HasPrefix(absPath, sysDir) {
			return fmt.Errorf("path accesses sensitive system directory")
		}
	}

	return nil
}

func PathFromWd(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return path, errors.Wrap(err, "unable to get working directory")
	}
	relPath, err := filepath.Rel(wd, path)
	if err != nil {
		return path, errors.Wrap(err, "unable to get relative path")
	}
	return relPath, nil
}

func ValidateOneOf(fieldName string, vals ...interface{}) error {
	var count int
	for _, val := range vals {
		if val == nil {
			continue
		}

		isPtr := reflect.ValueOf(val).Kind() == reflect.Ptr && !reflect.ValueOf(val).IsNil()
		isVal := reflect.ValueOf(val).Kind() != reflect.Ptr && !reflect.ValueOf(val).IsZero()
		if isPtr || isVal {
			count++
		}
	}
	if count == 0 {
		return fmt.Errorf("must define at least one %s", fieldName)
	} else if count > 1 {
		return fmt.Errorf("must define only one %s", fieldName)
	}
	return nil
}

func IsZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}
	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}
