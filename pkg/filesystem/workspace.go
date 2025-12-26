package filesystem

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/types/workspace"
)

const WorkspaceConfigFileName = "flow.yaml"

func DefaultWorkspaceDir() string {
	return CachedDataDirPath() + "/default"
}

func InitWorkspaceConfig(name, path string) error {
	wsCfg := workspace.DefaultWorkspaceConfig(name)

	if err := EnsureWorkspaceDir(path); err != nil {
		return errors.Wrap(err, "unable to ensure workspace directory")
	}

	if err := WriteWorkspaceConfig(path, wsCfg); err != nil {
		return errors.Wrap(err, "unable to write workspace config")
	}
	return nil
}

func EnsureDefaultWorkspace() error {
	if _, err := os.Stat(DefaultWorkspaceDir()); os.IsNotExist(err) {
		err = os.MkdirAll(DefaultWorkspaceDir(), 0750)
		if err != nil {
			return errors.Wrap(err, "unable to create default workspace directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "unable to check for default workspace directory")
	}
	return nil
}

func WorkspaceConfigExists(workspacePath string) bool {
	_, err := os.Stat(filepath.Join(workspacePath, WorkspaceConfigFileName))
	return !os.IsNotExist(err)
}

func EnsureWorkspaceDir(workspacePath string) error {
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		err = os.MkdirAll(workspacePath, 0750)
		if err != nil {
			return errors.Wrap(err, "unable to create workspace directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "unable to check for workspace directory")
	}
	return nil
}

func EnsureWorkspaceConfig(workspaceName, workspacePath string) error {
	if _, err := os.Stat(filepath.Join(workspacePath, WorkspaceConfigFileName)); os.IsNotExist(err) {
		return InitWorkspaceConfig(workspaceName, workspacePath)
	} else if err != nil {
		return errors.Wrapf(err, "unable to check for workspace %s config file", workspaceName)
	}
	return nil
}

func WriteWorkspaceConfig(workspacePath string, config *workspace.Workspace) error {
	wsFile := filepath.Join(workspacePath, WorkspaceConfigFileName)
	file, err := os.OpenFile(filepath.Clean(wsFile), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "unable to open workspace config file")
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return errors.Wrap(err, "unable to truncate workspace config file")
	}

	err = yaml.NewEncoder(file).Encode(config)
	if err != nil {
		return errors.Wrap(err, "unable to encode workspace config file")
	}

	return nil
}

func LoadWorkspaceConfig(workspaceName, workspacePath string) (*workspace.Workspace, error) {
	if err := EnsureWorkspaceDir(workspacePath); err != nil {
		return nil, errors.Wrap(err, "unable to ensure workspace directory")
	} else if err := EnsureWorkspaceConfig(workspaceName, workspacePath); err != nil {
		return nil, errors.Wrap(err, "unable to ensure workspace config file")
	}

	wsCfg := &workspace.Workspace{}
	wsFile := filepath.Join(workspacePath, WorkspaceConfigFileName)
	file, err := os.Open(filepath.Clean(wsFile))
	if err != nil {
		return nil, errors.Wrap(err, "unable to open workspace config file")
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(wsCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode workspace config file")
	}

	wsCfg.SetContext(workspaceName, workspacePath)
	return wsCfg, nil
}
