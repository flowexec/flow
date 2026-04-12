package version

import (
	// using blank import for embed as it is only used inside comments.
	_ "embed"
	"fmt"
	"runtime"
	"strings"
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
)

// GoVersion returns the version of the go runtime used to compile the binary.
var goVersion = runtime.Version()

// OsArch returns the os and arch used to build the binary.
var osArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

// generateOutput return the output of the version command.
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
	return fmt.Sprintf(`

Version: %s
Git Commit: %s
Build date: %s
Go version: %s
OS / Arch : %s
`, strings.TrimSpace(version), strings.TrimSpace(gitCommit), strings.TrimSpace(buildDate), goVersion, osArch)
}

func String() string {
	return generateOutput()
}

func SemVer() string {
	return strings.TrimSpace(version)
}
