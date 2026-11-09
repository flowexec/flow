package git

import (
	"os"
	"path/filepath"
	"strings"
)

// CheckoutEnvKey names the environment variable that carries a run's resolved checkout.
const CheckoutEnvKey = "FLOW_CHECKOUT_DIR"

// ResolveCheckout returns the local checkout an executable in a workspace operates on. A workspace
// without a git source is its own checkout. For a git workspace it is the bound checkout, unless
// cwd is inside that same repository (the bound clone or one of its worktrees), in which case the
// caller's own working tree wins. It is "" when a git workspace has no checkout bound.
func ResolveCheckout(wsPath string, gitSourced bool, bound, cwd string) string {
	if !gitSourced {
		return wsPath
	}
	bound = expandHome(bound)
	if bound == "" {
		return ""
	}
	if cwd != "" {
		if top, ok := sameRepoTop(bound, cwd); ok {
			return top
		}
	}
	return bound
}

// sameRepoTop reports the top level of cwd's working tree when cwd and bound share a git common
// directory, which is true for a clone and all of its worktrees.
func sameRepoTop(bound, cwd string) (string, bool) {
	cwdCommon, err := commonDir(cwd)
	if err != nil {
		return "", false
	}
	boundCommon, err := commonDir(bound)
	if err != nil || cwdCommon != boundCommon {
		return "", false
	}
	top, err := gitOutput(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(top), true
}

func commonDir(dir string) (string, error) {
	out, err := gitOutput(dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return resolvePath(strings.TrimSpace(out)), nil
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path[1:], "/"))
		}
	}
	return path
}
