package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/teatest"
)

var updateEnvKey = "UPDATE_GOLDEN_FILES"

func UpdateGolden(tb testing.TB, out []byte) {
	golden := filepath.Join("testdata", tb.Name()+".golden")
	if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
		tb.Fatal(err)
	}
	normalized := NormalizeTmpDirs(string(out))
	if err := os.WriteFile(golden, []byte(normalized), 0o600); err != nil {
		tb.Fatal(err)
	}
}

func MaybeUpdateGolden(tb testing.TB, out []byte) {
	if os.Getenv(updateEnvKey) == "true" {
		UpdateGolden(tb, out)
	}
}

// NormalizeTmpDirs replaces all file paths under the system temp dir with /TMPDIR.
func NormalizeTmpDirs(input string) string {
	tmpDir := os.TempDir()
	if !strings.HasSuffix(tmpDir, "/") {
		tmpDir += "/"
	}
	tmpDirPattern := regexp.QuoteMeta(tmpDir)
	// Match temp dir followed by any path plus any trailing whitespace
	re := regexp.MustCompile(tmpDirPattern + `[^\s\x00-\x1F]+\s*`)
	result := re.ReplaceAllStringFunc(input, func(m string) string {
		pathAfterTmp := m[len(tmpDir):]
		// Trim whitespace from the captured path
		pathAfterTmp = strings.TrimSpace(pathAfterTmp)
		parts := strings.Split(pathAfterTmp, "/")
		if len(parts) >= 2 {
			// Keep only the last meaningful part
			return "/TMPDIR/" + parts[len(parts)-1]
		}
		return "/TMPDIR/" + pathAfterTmp + "  "
	})

	return result
}

func RequireEqualSnapshot(tb testing.TB, out []byte) {
	normalized := NormalizeTmpDirs(string(out))
	teatest.RequireEqualOutput(tb, []byte(normalized))
}
