package common

import (
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/services/open"
)

const HeaderContextKey = "ctx"

var termEditors = []string{"vim", "nvim", "emacs", "nano"}

func OpenInEditor(path string, stdIn, stdOut *os.File) error {
	preferred := os.Getenv("EDITOR")
	if preferred != "" && !slices.Contains(termEditors, preferred) {
		return open.OpenWith(preferred, path)
	}
	if preferred == "" {
		preferred = "vim"
	}
	cmd := exec.Command(preferred, path) // #nosec G204
	cmd.Stdin = stdIn
	cmd.Stdout = stdOut
	return cmd.Run()
}

const (
	YAMLFormat = "yaml"
	ymlFormat  = "yml"
	JSONFormat = "json"
)

func NormalizeFormat(format string) string {
	switch strings.ToLower(format) {
	case YAMLFormat, ymlFormat:
		return YAMLFormat
	case JSONFormat:
		return JSONFormat
	default:
		// tui is a special case, it's the default output mode and should not be logged as an unsupported format
		if format != "" && format != "tui" {
			logger.Log().Warnf("Unsupported output format '%s', defaulting to YAML", format)
		}
		return YAMLFormat
	}
}
