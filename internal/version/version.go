package version

import (
	// using blank import for embed as it is only used inside comments.
	_ "embed"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"golang.org/x/mod/module"
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

var readBuildInfo = debug.ReadBuildInfo

func init() {
	resolve()
}

// generateOutput return the output of the version command.
func generateOutput() string {
	v := version
	if v == "" {
		v = unknown
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n\nVersion: %s\n", v)

	// Commit and build date are only stamped by release builds, so omit them otherwise.
	if c := strings.TrimSpace(gitCommit); c != "" {
		fmt.Fprintf(&b, "Git Commit: %s\n", c)
	}
	if d := strings.TrimSpace(buildDate); d != "" {
		fmt.Fprintf(&b, "Build date: %s\n", d)
	}

	fmt.Fprintf(&b, "Go version: %s\nOS / Arch : %s\n", goVersion, osArch)
	return b.String()
}

func String() string {
	return generateOutput()
}

func SemVer() string {
	return version
}

// Short returns a simplified version string.
// Examples: "v2", "v2.1", "v2.1.3" (pre-release tags dropped, trailing .0 segments removed)
func Short() string {
	if version == "" {
		return ""
	}

	mainVersion := strings.TrimSpace(strings.SplitN(version, "-", 2)[0])
	if !strings.HasPrefix(mainVersion, "v") {
		mainVersion = "v" + mainVersion
	}
	segments := strings.Split(mainVersion, ".")
	if len(segments) < 3 {
		return mainVersion
	}

	switch {
	case segments[1] == "0" && segments[2] == "0":
		return segments[0]
	case segments[2] == "0":
		return strings.Join(segments[:2], ".")
	default:
		return mainVersion
	}
}

// resolve falls back to the module version recorded by the Go toolchain when the version
// wasn't set via ldflags.
func resolve() {
	version = strings.TrimSpace(version)
	if version == "" {
		version = buildInfoVersion()
	}
}

// buildInfoVersion returns the main module version only when it is a released tag. Local
// checkout builds report "(devel)", a pseudo-version, or a "+dirty" suffix and are treated as
// dev builds.
func buildInfoVersion() string {
	bi, ok := readBuildInfo()
	if !ok || bi == nil {
		return ""
	}
	v := bi.Main.Version
	if v == "" || v == "(devel)" || strings.Contains(v, "+dirty") || module.IsPseudoVersion(v) {
		return ""
	}
	return v
}
