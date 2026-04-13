package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/pkg/store"
	"github.com/flowexec/flow/types/workspace"
)

const wsCacheKey = "workspaces"

//go:generate mockgen -destination=mocks/mock_workspace_cache.go -package=mocks github.com/flowexec/flow/pkg/cache WorkspaceCache
type WorkspaceCache interface {
	Update() error
	GetData() *WorkspaceCacheData
	GetLatestData() (*WorkspaceCacheData, error)
	GetWorkspaceConfigList() (workspace.WorkspaceList, error)
}
type WorkspaceCacheData struct {
	// Map of workspace name to workspace config
	Workspaces map[string]*workspace.Workspace `json:"workspaces" yaml:"workspaces"`
	// Map of workspace name to workspace path
	WorkspaceLocations map[string]string `json:"workspaceLocations" yaml:"workspaceLocations"`
}

type WorkspaceCacheImpl struct {
	Data  *WorkspaceCacheData
	Store store.DataStore
}

func NewWorkspaceCache(s store.DataStore) WorkspaceCache {
	workspaceCache := &WorkspaceCacheImpl{
		Store: s,
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
			logger.Log().Error("config not found for workspace", "name", name, "path", path)
			continue
		}
		cacheData.Workspaces[name] = wsCfg
		cacheData.WorkspaceLocations[name] = path
	}
	data, err := json.Marshal(cacheData)
	if err != nil {
		return errors.Wrap(err, "unable to encode cache data")
	}

	if err := c.Store.SetCacheEntry(wsCacheKey, data); err != nil {
		return errors.Wrap(err, "unable to write cache data")
	}

	logger.Log().Debug("Successfully updated workspace cache data", "count", len(cacheData.Workspaces))
	return nil
}

func (c *WorkspaceCacheImpl) GetData() *WorkspaceCacheData {
	return c.Data
}

func (c *WorkspaceCacheImpl) GetLatestData() (*WorkspaceCacheData, error) {
	cacheData, err := c.Store.GetCacheEntry(wsCacheKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load workspace cache data")
	}

	if cacheData == nil {
		// Lazy-migrate from legacy YAML file if it exists.
		cacheData, err = migrateWorkspaceCacheFromFile()
		if err != nil {
			return nil, errors.Wrap(err, "unable to migrate workspace cache")
		}
		if cacheData != nil {
			if writeErr := c.Store.SetCacheEntry(wsCacheKey, cacheData); writeErr != nil {
				logger.Log().Warn("failed to persist migrated workspace cache", "err", writeErr)
			}
		}
	}

	if cacheData == nil {
		if err := c.Update(); err != nil {
			return nil, errors.Wrap(err, "unable to get updated workspace cache data")
		}
		return c.Data, nil
	}

	c.Data = &WorkspaceCacheData{}
	if err := json.Unmarshal(cacheData, c.Data); err != nil {
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

// migrateWorkspaceCacheFromFile attempts to read the legacy YAML cache file, re-encodes it
// as JSON, deletes the old file, and returns the JSON bytes. Returns nil, nil if no legacy file.
func migrateWorkspaceCacheFromFile() ([]byte, error) {
	// Legacy key was "workspace" (singular); legacy dir was <cacheDir>/latestcache/
	legacyPath := filepath.Join(filesystem.CachedDataDirPath(), "latestcache", "workspace")
	yamlData, err := os.ReadFile(legacyPath)
	if err != nil {
		return nil, nil //nolint:nilerr
	}

	var legacy WorkspaceCacheData
	if err := yaml.Unmarshal(yamlData, &legacy); err != nil {
		return nil, nil //nolint:nilerr
	}

	jsonData, err := json.Marshal(&legacy)
	if err != nil {
		return nil, nil //nolint:nilerr
	}

	_ = os.Remove(legacyPath)
	return jsonData, nil
}
