package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flowexec/flow/v2/types/config"
	"github.com/flowexec/flow/v2/types/workspace"
)

// WorkspaceOverrideEnv pins the workspace for a single invocation, by registered name or by
// path. Exporting it beats threading a flag through every call, which matters most for agents
// and CI where the caller can set the environment once but can't remember a flag every time.
const WorkspaceOverrideEnv = "FLOW_WORKSPACE"

// WorkspaceSource records how a workspace was resolved, so callers can explain the choice
// (`flow workspace get`, MCP get_info) instead of leaving the user to guess.
type WorkspaceSource string

const (
	// SourceOverride: named explicitly via --workspace or FLOW_WORKSPACE.
	SourceOverride WorkspaceSource = "override"
	// SourceRegistered: walking up from the working directory found a registered workspace root.
	SourceRegistered WorkspaceSource = "registered"
	// SourceDiscovered: walking up found a flow.yaml that is not registered anywhere.
	SourceDiscovered WorkspaceSource = "discovered"
	// SourcePrefix: no flow.yaml above the working directory, but it sits inside a registered
	// workspace's tree.
	SourcePrefix WorkspaceSource = "prefix"
	// SourceCurrent: fell back to the workspace persisted in the user config.
	SourceCurrent WorkspaceSource = "current"
)

// ResolvedWorkspace is the workspace a command should operate in, plus how it was chosen.
// Registered is false for a workspace found only by walking up from the working directory —
// it is valid to run in, but it exists nowhere in the user config and is never persisted.
type ResolvedWorkspace struct {
	Workspace  *workspace.Workspace
	Name       string
	Path       string
	Registered bool
	Source     WorkspaceSource
}

// ResolveOptions tunes workspace resolution. The zero value resolves from the process working
// directory with no override, which is what almost every caller wants.
type ResolveOptions struct {
	// Dir is the directory to resolve from. Defaults to the process working directory.
	Dir string
	// Override is an explicit workspace name or path (a --workspace flag value). When empty,
	// WorkspaceOverrideEnv is consulted instead.
	Override string
}

// discoveryBoundaryDirs are directory names whose contents are copies of, or dependencies of,
// some other project. A flow.yaml inside one of them belongs to that copy, not to the tree the
// user is working in, so discovery walks past it rather than adopting it as a root. These mirror
// the paths executable discovery already refuses to scan.
var discoveryBoundaryDirs = []string{
	".git",
	".claude",
	"vendor",
	"third_party",
	"external",
	"node_modules",
}

// ResolveWorkspace picks the workspace for a directory.
//
// Precedence: an explicit override always wins; otherwise, in dynamic mode, the nearest
// flow.yaml at or above dir wins — the same way make and bazel locate their root — falling back
// to a registered workspace containing dir and finally to the workspace persisted in the config.
// Fixed mode skips discovery entirely, because pinning a workspace and then auto-switching on
// every cd would defeat the point of pinning it.
//
// Returns (nil, nil) when nothing resolves. That is an ordinary state — a fresh install has no
// workspaces at all — so callers must handle a nil result rather than treating it as an error.
// A sentinel error would invert that: every caller would have to remember to special-case it,
// and forgetting once turns "no workspace yet" into a fatal.
//
//nolint:nilnil
func ResolveWorkspace(cfg *config.Config, opts ResolveOptions) (*ResolvedWorkspace, error) {
	if cfg == nil {
		return nil, nil
	}

	dir := opts.Dir
	if dir == "" {
		// A working directory that no longer exists is not an error here; resolution just falls
		// through to the config-based paths below.
		if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}
	if dir != "" {
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
		dir = NormalizePath(dir)
	}

	override := opts.Override
	if override == "" {
		override = os.Getenv(WorkspaceOverrideEnv)
	}
	if override != "" {
		return resolveOverride(cfg, override)
	}

	if cfg.WorkspaceMode == config.ConfigWorkspaceModeDynamic && dir != "" {
		if root, found := FindWorkspaceRootExcluding(dir, discoveryBoundary(cfg)); found {
			if name, registered := cfg.NameForWorkspacePath(root); registered {
				return loadResolved(name, cfg.Workspaces[name], SourceRegistered)
			}
			return readResolved(root, SourceDiscovered)
		}

		// No flow.yaml anywhere above dir. A registered workspace may still claim this path —
		// its own flow.yaml could have been deleted, or the workspace root may be unreachable.
		if name, found := cfg.WorkspaceForPath(dir); found {
			return loadResolved(name, cfg.Workspaces[name], SourcePrefix)
		}
	}

	if path, found := cfg.Workspaces[cfg.CurrentWorkspace]; found && path != "" {
		return loadResolved(cfg.CurrentWorkspace, path, SourceCurrent)
	}

	return nil, nil
}

// resolveOverride handles an explicit --workspace / FLOW_WORKSPACE value, which may be either a
// registered name or a path to any directory holding a flow.yaml. Overrides are honored in fixed
// mode too — they are how a fixed-mode user opts into a different workspace for one command.
func resolveOverride(cfg *config.Config, override string) (*ResolvedWorkspace, error) {
	if path, found := cfg.Workspaces[override]; found && path != "" {
		return loadResolved(override, path, SourceOverride)
	}

	if !looksLikePath(override) {
		return nil, fmt.Errorf(
			"unknown workspace %q - not a registered workspace name, and not a path to a directory with a %s",
			override, WorkspaceConfigFileName,
		)
	}

	path, err := expandOverridePath(override)
	if err != nil {
		return nil, err
	}
	if !WorkspaceConfigExists(path) {
		return nil, fmt.Errorf("no %s found in %s", WorkspaceConfigFileName, path)
	}
	if name, found := cfg.NameForWorkspacePath(path); found {
		return loadResolved(name, cfg.Workspaces[name], SourceOverride)
	}
	return readResolved(path, SourceOverride)
}

func looksLikePath(v string) bool {
	return v == "." || v == ".." ||
		filepath.IsAbs(v) ||
		strings.HasPrefix(v, "~") ||
		strings.ContainsRune(v, filepath.Separator) ||
		strings.ContainsRune(v, '/')
}

func expandOverridePath(v string) (string, error) {
	if strings.HasPrefix(v, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to expand %q - %w", v, err)
		}
		v = filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(v, "~"), "/"))
	}
	abs, err := filepath.Abs(v)
	if err != nil {
		return "", fmt.Errorf("unable to resolve workspace path %q - %w", v, err)
	}
	return NormalizePath(abs), nil
}

// discoveryBoundary rejects candidate roots that a registered workspace already refuses to scan
// — a vendored dependency or a repo copy checked out inside it. Without this, `cd vendor/somelib`
// would silently make that dependency the workspace.
//
// The check applies only *below* a registered workspace root. A standalone clone that happens to
// live in ~/external/ is a perfectly good workspace; only a copy sitting inside a workspace that
// deliberately excludes it is not.
func discoveryBoundary(cfg *config.Config) func(string) bool {
	return func(root string) bool {
		// An explicitly registered path is always a legitimate root, wherever it lives.
		if _, registered := cfg.NameForWorkspacePath(root); registered {
			return false
		}
		parent, found := cfg.WorkspaceForPath(root)
		if !found {
			return false
		}
		rel, err := filepath.Rel(NormalizePath(cfg.Workspaces[parent]), root)
		if err != nil {
			return false
		}
		for _, segment := range strings.Split(filepath.ToSlash(rel), "/") {
			if slices.Contains(discoveryBoundaryDirs, segment) {
				return true
			}
		}
		return false
	}
}

// loadResolved builds a result for a registered workspace, creating its flow.yaml if the user
// registered a path that does not have one yet.
func loadResolved(name, path string, src WorkspaceSource) (*ResolvedWorkspace, error) {
	ws, err := LoadWorkspaceConfig(name, path)
	if err != nil {
		return nil, err
	}
	return &ResolvedWorkspace{
		Workspace: ws, Name: name, Path: NormalizePath(path), Registered: true, Source: src,
	}, nil
}

// readResolved builds a result for a workspace found on disk but absent from the user config,
// reading its flow.yaml without writing anything into a directory the user never registered.
func readResolved(path string, src WorkspaceSource) (*ResolvedWorkspace, error) {
	name := filepath.Base(path)
	ws, err := ReadWorkspaceConfig(name, path)
	if err != nil {
		return nil, err
	}
	return &ResolvedWorkspace{
		Workspace: ws, Name: name, Path: path, Registered: false, Source: src,
	}, nil
}
