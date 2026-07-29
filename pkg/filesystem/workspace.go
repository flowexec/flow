package filesystem

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/types/workspace"
)

const WorkspaceConfigFileName = "flow.yaml"

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

// WorkspaceConfigExists reports whether workspacePath holds a flow.yaml file. Workspace
// discovery walks up through arbitrary directories testing this, so a directory that merely
// happens to be named flow.yaml must not count as a workspace root.
func WorkspaceConfigExists(workspacePath string) bool {
	info, err := os.Stat(filepath.Join(workspacePath, WorkspaceConfigFileName))
	if err != nil {
		return false
	}
	return !info.IsDir()
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

// LoadWorkspaceConfig reads a registered workspace's config, creating the directory and a
// default flow.yaml if either is missing. Only use this for paths the user has explicitly
// registered — see ReadWorkspaceConfig for the read-only variant discovery must use.
func LoadWorkspaceConfig(workspaceName, workspacePath string) (*workspace.Workspace, error) {
	if err := EnsureWorkspaceDir(workspacePath); err != nil {
		return nil, errors.Wrap(err, "unable to ensure workspace directory")
	} else if err := EnsureWorkspaceConfig(workspaceName, workspacePath); err != nil {
		return nil, errors.Wrap(err, "unable to ensure workspace config file")
	}
	return ReadWorkspaceConfig(workspaceName, workspacePath)
}

// ReadWorkspaceConfig reads a workspace's flow.yaml without creating anything. Workspace
// discovery resolves roots the user never registered, so it must never write into them the way
// LoadWorkspaceConfig does.
func ReadWorkspaceConfig(workspaceName, workspacePath string) (*workspace.Workspace, error) {
	wsCfg := &workspace.Workspace{}
	wsFile := filepath.Join(workspacePath, WorkspaceConfigFileName)
	file, err := os.Open(filepath.Clean(wsFile))
	if err != nil {
		return nil, errors.Wrap(err, "unable to open workspace config file")
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(wsCfg)
	if err != nil && !errors.Is(err, io.EOF) {
		// An empty flow.yaml is a valid workspace marker with all-default settings.
		return nil, errors.Wrap(err, "unable to decode workspace config file")
	}

	wsCfg.SetContext(workspaceName, workspacePath)
	return wsCfg, nil
}
