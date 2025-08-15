package fileparser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/utils"
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

	for _, file := range files {
		fn := filepath.Base(file)
		expandedFile := utils.ExpandPath(file, filepath.Dir(flowFilePath), nil)

		if info, err := os.Stat(expandedFile); err != nil {
			logger.Log().Error(err, fmt.Sprintf("unable to import executables from file %s", file))
			continue
		} else if info.IsDir() {
			logger.Log().Errorx("unable to import executables", "err", fmt.Sprintf("%s is not a file", file))
			continue
		}

		switch strings.ToLower(fn) {
		case "package.json":
			execs, err := ExecutablesFromPackageJSON(wsPath, expandedFile)
			if err != nil {
				logger.Log().Error(err, fmt.Sprintf("unable to import executables from file (%s)", file))
			}
			for _, exec := range execs {
				exec.SetContext(wsName, wsPath, flowFileNs, flowFilePath)
				exec.SetInheritedFields(flowFile)
				executables = append(executables, exec)
			}
		case "makefile":
			execs, err := ExecutablesFromMakefile(wsPath, expandedFile)
			if err != nil {
				logger.Log().Error(err, fmt.Sprintf("unable to import executables from file (%s)", file))
			}
			for _, exec := range execs {
				exec.SetContext(wsName, wsPath, flowFileNs, flowFilePath)
				exec.SetInheritedFields(flowFile)
				executables = append(executables, exec)
			}
		case "docker-compose.yml", "docker-compose.yaml":
			execs, err := ExecutablesFromDockerCompose(wsPath, expandedFile)
			if err != nil {
				logger.Log().Error(err, fmt.Sprintf("unable to import executables from file (%s)", file))
			}
			for _, exec := range execs {
				exec.SetContext(wsName, wsPath, flowFileNs, flowFilePath)
				exec.SetInheritedFields(flowFile)
				executables = append(executables, exec)
			}
		default:
			ext := filepath.Ext(fn)
			if ext != ".sh" {
				logger.Log().Warnx("unable to import executables - unsupported file type", "file", file)
				continue
			}
			exec, err := ExecutablesFromShFile(wsPath, expandedFile)
			if err != nil {
				logger.Log().Error(err, fmt.Sprintf("unable to import executables from file (%s)", file))
				continue
			}
			exec.SetContext(wsName, wsPath, flowFileNs, flowFilePath)
			exec.SetInheritedFields(flowFile)
			executables = append(executables, exec)
		}
	}

	return executables, nil
}

func shortenWsPath(wsPath string, path string) string {
	if strings.HasPrefix(path, wsPath) {
		return "//" + strings.TrimPrefix(path[len(wsPath):], "/")
	}

	return path
}
