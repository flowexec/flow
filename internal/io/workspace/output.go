package workspace

import (
	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/workspace"
)

func PrintWorkspaceList(format string, workspaces workspace.WorkspaceList) {
	logger.Log().Debugf("listing %d workspaces", len(workspaces))
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := workspaces.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal workspace list - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := workspaces.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal workspace list - %v", err)
		}
		logger.Log().Println(str)
	}
}

func PrintWorkspaceConfig(format string, ws *workspace.Workspace) {
	if ws == nil {
		logger.Log().Fatalf("Workspace type is nil")
	}
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := ws.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal workspace config - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := ws.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal workspace config - %v", err)
		}
		logger.Log().Println(str)
	}
}
