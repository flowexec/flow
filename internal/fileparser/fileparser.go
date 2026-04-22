package fileparser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/flow/internal/utils"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
)

const generatedTag = "generated"

func ExecutablesFromImports(
	wsName string, flowFile *executable.FlowFile,
) (executable.ExecutableList, error) {
	executables := make(executable.ExecutableList, 0)
	wsPath := flowFile.WorkspacePath()
	flowFilePath := flowFile.ConfigPath()
	flowFileNs := flowFile.Namespace
	files := append(flowFile.FromFile, flowFile.Imports...) //nolint:gocritic

	setCtx := func(execs ...*executable.Executable) {
		for _, e := range execs {
			e.SetContext(wsName, wsPath, flowFileNs, flowFilePath)
			e.SetInheritedFields(flowFile)
		}
	}

	for _, file := range files {
		fn := filepath.Base(file)
		expandedFile := utils.ExpandPath(file, filepath.Dir(flowFilePath), nil)

		if info, err := os.Stat(expandedFile); err != nil {
			logger.Log().WrapError(err, fmt.Sprintf("unable to import executables from file %s", file))
			continue
		} else if info.IsDir() {
			logger.Log().Error("unable to import executables", "err", fmt.Sprintf("%s is not a file", file))
			continue
		}

		parsed, err := parseImportFile(wsPath, fn, expandedFile)
		if err != nil {
			logger.Log().WrapError(err, fmt.Sprintf("unable to import executables from file (%s)", file))
			continue
		}
		setCtx(parsed...)
		executables = append(executables, parsed...)
	}

	return executables, nil
}

func parseImportFile(wsPath, fn, expandedFile string) (executable.ExecutableList, error) {
	switch strings.ToLower(fn) {
	case "package.json":
		return ExecutablesFromPackageJSON(wsPath, expandedFile)
	case "makefile":
		return ExecutablesFromMakefile(wsPath, expandedFile)
	case "docker-compose.yml", "docker-compose.yaml":
		return ExecutablesFromDockerCompose(wsPath, expandedFile)
	default:
		return parseScriptFile(wsPath, fn, expandedFile)
	}
}

func parseScriptFile(wsPath, fn, expandedFile string) (executable.ExecutableList, error) {
	ext := strings.ToLower(filepath.Ext(fn))
	var exec *executable.Executable
	var err error
	switch ext {
	case ".sh":
		exec, err = ExecutablesFromShFile(wsPath, expandedFile)
	case ".bat", ".cmd":
		exec, err = ExecutablesFromBatFile(wsPath, expandedFile)
	case ".ps1":
		exec, err = ExecutablesFromPs1File(wsPath, expandedFile)
	default:
		logger.Log().Warn("unable to import executables - unsupported file type", "file", fn)
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return executable.ExecutableList{exec}, nil
}

func shortenWsPath(wsPath string, path string) string {
	if strings.HasPrefix(path, wsPath) {
		return "//" + strings.TrimPrefix(path[len(wsPath):], string(filepath.Separator))
	}

	return path
}
