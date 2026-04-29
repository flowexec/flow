package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

func EnsureExecutableDir(workspacePath, subPath string) error {
	if _, err := os.Stat(filepath.Join(workspacePath, subPath)); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Join(workspacePath, subPath), 0750)
		if err != nil {
			return errors.Wrap(err, "unable to create executable directory")
		}
	}
	return nil
}

func WriteFlowFile(cfgFile string, cfg *executable.FlowFile) error {
	file, err := os.OpenFile(filepath.Clean(cfgFile), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "unable to open cfg file")
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return errors.Wrap(err, "unable to truncate config file")
	}

	err = yaml.NewEncoder(file).Encode(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to encode config file")
	}

	return nil
}

func LoadFlowFile(cfgFile string) (*executable.FlowFile, error) {
	file, err := os.Open(filepath.Clean(cfgFile))
	if err != nil {
		return nil, errors.Wrap(err, "unable to open config file")
	}
	defer file.Close()

	cfg := &executable.FlowFile{}
	err = yaml.NewDecoder(file).Decode(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode config file")
	}
	return cfg, nil
}

func LoadWorkspaceFlowFiles(
	workspaceCfg *workspace.Workspace,
) (executable.FlowFileList, error) {
	cfgFiles, err := findFlowFiles(workspaceCfg)
	if err != nil {
		return nil, err
	}

	var cfgs executable.FlowFileList
	for _, cfgFile := range cfgFiles {
		cfg, err := LoadFlowFile(cfgFile)
		if err != nil {
			logger.Log().Error(
				fmt.Sprintf("unable to load flow file: %s", errors.Cause(err)),
				"file", cfgFile,
			)
			continue
		}
		cfg.SetDefaults()
		cfg.SetContext(workspaceCfg.AssignedName(), workspaceCfg.Location(), cfgFile)
		cfgs = append(cfgs, cfg)
	}
	logger.Log().Debug(
		fmt.Sprintf("loaded %d config files", len(cfgs)),
		"workspace",
		workspaceCfg.AssignedName(),
	)

	return cfgs, nil
}

// defaultExcutablePaths lists directory names and patterns that are never scanned for flow files.
// Simple directory names (e.g. "node_modules/") use any-depth matching and are skipped wherever
// they appear in the workspace tree.
var defaultExcutablePaths = []string{
	// Version control internals — .git/worktrees contains full repo checkouts
	".git/",

	// AI coding-assistant worktree directories — these are full repo copies
	".claude/",

	// Dependency copies — vendored or installed packages may ship their own flow files
	"vendor/",
	"third_party/",
	"external/",
	"node_modules/",

	// File extension exclusions
	"*.js.flow",
}

func findFlowFiles(workspaceCfg *workspace.Workspace) ([]string, error) {
	var includePaths, excludedPaths []string
	if workspaceCfg.Executables != nil {
		includePaths = workspaceCfg.Executables.Included
		if len(includePaths) == 0 {
			includePaths = []string{workspaceCfg.Location()}
		}

		excludedPaths = workspaceCfg.Executables.Excluded
	} else {
		includePaths = []string{workspaceCfg.Location()}
	}
	excludedPaths = append(excludedPaths, defaultExcutablePaths...)

	var cfgPaths []string
	walkDirFunc := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				logger.Log().Debug("cfg path does not exist", "path", path)
				return nil
			}
			return err
		}
		if isPathIncluded(path, workspaceCfg.Location(), includePaths) {
			if isPathExcluded(path, workspaceCfg.Location(), excludedPaths) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if executable.HasFlowFileExt(entry.Name()) {
				cfgPaths = append(cfgPaths, path)
			}
		}
		return nil
	}

	if err := filepath.WalkDir(workspaceCfg.Location(), walkDirFunc); err != nil {
		return nil, err
	}
	return cfgPaths, nil
}

func pathMatches(path, basePath string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		relPath = path
	}
	relSlash := filepath.ToSlash(relPath)

	for _, p := range patterns {
		pattern := p
		if strings.HasPrefix(pattern, "//") {
			pattern = strings.Replace(pattern, "//", basePath+string(filepath.Separator), 1)
		}
		if singlePatternMatches(pattern, path, relPath, relSlash) {
			return true
		}
	}
	return false
}

// singlePatternMatches reports whether pattern matches path (absolute), relPath, or relSlash.
// Pattern semantics:
//   - Trailing "/" is stripped for matching so "node_modules/" matches the dir entry "node_modules".
//   - Patterns without a path separator match at any depth ("node_modules/" matches "a/b/node_modules").
//   - Patterns containing *, ?, or [ are matched as globs against the filename and the relative path.
func singlePatternMatches(pattern, path, relPath, relSlash string) bool {
	sep := string(filepath.Separator)

	// Strip trailing separator so "node_modules/" also matches the directory entry itself.
	normal := strings.TrimSuffix(strings.TrimSuffix(pattern, sep), "/")

	// Absolute path: exact or prefix match.
	if path == normal || path == pattern || strings.HasPrefix(path, normal+sep) {
		return true
	}

	// Relative path: exact or prefix match (both OS separator and slash forms).
	if relPath == normal || relSlash == normal || relPath == pattern ||
		strings.HasPrefix(relPath, normal+sep) ||
		strings.HasPrefix(relSlash, normal+"/") {
		return true
	}

	// Any-depth match: a pattern with no separator matches the name at any level.
	// e.g. "node_modules/" matches "frontend/node_modules" and "a/b/node_modules/c".
	if !strings.Contains(normal, "/") && !strings.Contains(normal, sep) {
		if slices.Contains(strings.Split(relSlash, "/"), normal) {
			return true
		}
	}

	// Glob match against the filename and the full relative path.
	if strings.ContainsAny(pattern, "*?[") {
		if ok, _ := filepath.Match(pattern, filepath.Base(path)); ok {
			return true
		}
		if ok, _ := filepath.Match(pattern, relSlash); ok {
			return true
		}
	}

	return false
}

func isPathIncluded(path, basePath string, includePaths []string) bool {
	if includePaths == nil {
		return true
	}
	return pathMatches(path, basePath, includePaths)
}

func isPathExcluded(path, basePath string, excludedPaths []string) bool {
	return pathMatches(path, basePath, excludedPaths)
}
