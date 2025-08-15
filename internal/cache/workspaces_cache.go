package cache

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/filesystem"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/workspace"
)

const wsCacheKey = "workspace"

//go:generate mockgen -destination=mocks/mock_workspace_cache.go -package=mocks github.com/flowexec/flow/internal/cache WorkspaceCache
type WorkspaceCache interface {
	Update() error
	GetData() *WorkspaceCacheData
	GetLatestData() (*WorkspaceCacheData, error)
	GetWorkspaceConfigList() (workspace.WorkspaceList, error)
}
type WorkspaceCacheData struct {
	// Map of workspace name to workspace config
	Workspaces map[string]*workspace.Workspace `yaml:"workspaces"`
	// Map of workspace name to workspace path
	WorkspaceLocations map[string]string `yaml:"workspaceLocations"`
}

type WorkspaceCacheImpl struct {
	Data *WorkspaceCacheData
}

func NewWorkspaceCache() WorkspaceCache {
	workspaceCache := &WorkspaceCacheImpl{
		Data: &WorkspaceCacheData{
			Workspaces:         make(map[string]*workspace.Workspace),
			WorkspaceLocations: make(map[string]string),
		},
	}
	return workspaceCache
}

func (c *WorkspaceCacheImpl) Update() error {
	logger.Log().Debugf("Updating workspace cache data")

	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return err
	}

	cacheData := c.Data
	for name, path := range cfg.Workspaces {
		wsCfg, err := filesystem.LoadWorkspaceConfig(name, path)
		if err != nil {
			return errors.Wrap(err, "failed loading workspace config")
		} else if wsCfg == nil {
			logger.Log().Errorx("config not found for workspace", "name", name, "path", path)
			continue
		}
		cacheData.Workspaces[name] = wsCfg
		cacheData.WorkspaceLocations[name] = path
	}
	data, err := yaml.Marshal(cacheData)
	if err != nil {
		return errors.Wrap(err, "unable to encode cache data")
	}

	err = filesystem.WriteLatestCachedData(wsCacheKey, data)
	if err != nil {
		return errors.Wrap(err, "unable to write cache data")
	}

	logger.Log().Debugx("Successfully updated workspace cache data", "count", len(cacheData.Workspaces))
	return nil
}

func (c *WorkspaceCacheImpl) GetData() *WorkspaceCacheData {
	return c.Data
}

func (c *WorkspaceCacheImpl) GetLatestData() (*WorkspaceCacheData, error) {
	cacheData, err := filesystem.LoadLatestCachedData(wsCacheKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load workspace cache data")
	} else if cacheData == nil {
		if err := c.Update(); err != nil {
			return nil, errors.Wrap(err, "unable to get updated workspace cache data")
		}
	}

	c.Data = &WorkspaceCacheData{}
	if err := yaml.Unmarshal(cacheData, c.Data); err != nil {
		return nil, errors.Wrap(err, "unable to decode workspace cache data")
	}
	return c.Data, nil
}

func (c *WorkspaceCacheImpl) GetWorkspaceConfigList() (workspace.WorkspaceList, error) {
	var cache *WorkspaceCacheData
	if len(c.Data.Workspaces) == 0 {
		var err error
		cache, err = c.GetLatestData()
		if err != nil {
			return nil, err
		}
	} else {
		cache = c.GetData()
	}

	wsCfgs := make(workspace.WorkspaceList, 0, len(c.Data.Workspaces))
	for wsName, wsCfg := range cache.Workspaces {
		wsCfg.SetContext(wsName, cache.WorkspaceLocations[wsName])
		wsCfgs = append(wsCfgs, wsCfg)
	}
	return wsCfgs, nil
}
