package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/logger"
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
			logger.Log().Errorx("unable to load executable config file", "configFile", cfgFile, "err", err)
			continue
		}
		cfg.SetDefaults()
		cfg.SetContext(workspaceCfg.AssignedName(), workspaceCfg.Location(), cfgFile)
		cfgs = append(cfgs, cfg)
	}
	logger.Log().Debugx(
		fmt.Sprintf("loaded %d config files", len(cfgs)),
		"workspace",
		workspaceCfg.AssignedName(),
	)

	return cfgs, nil
}

var defaultExcutablePaths = []string{
	"vendor/",
	"third_party/",
	"external/",
	"node_modules/",
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
				logger.Log().Debugx("cfg path does not exist", "path", path)
				return nil
			}
			return err
		}
		if isPathIncluded(path, workspaceCfg.Location(), includePaths) {
			if isPathExcluded(path, workspaceCfg.Location(), excludedPaths) {
				return filepath.SkipDir
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
	if patterns == nil {
		return false
	}

	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		// fallback to absolute if relative path cannot be determined
		relPath = path
	}

	for _, p := range patterns {
		pattern := p
		if strings.HasPrefix(pattern, "//") {
			pattern = strings.Replace(pattern, "//", basePath+"/", 1)
		}

		if path == pattern || strings.HasPrefix(path, pattern) {
			return true
		}
		if relPath == pattern || strings.HasPrefix(relPath, pattern) {
			return true
		}

		for _, checkPath := range []string{path, relPath} {
			if strings.Contains(pattern, "*") ||
				strings.Contains(pattern, "?") ||
				strings.Contains(pattern, "[") {
				fileName := filepath.Base(checkPath)
				isMatch, err := filepath.Match(pattern, fileName)
				if err == nil && isMatch {
					return true
				}
			}
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
