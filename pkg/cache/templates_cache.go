package cache

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/executable"
)

const tmplCacheKey = "templates"

//go:generate mockgen -destination=mocks/mock_template_cache.go -package=mocks github.com/flowexec/flow/v2/pkg/cache TemplateCache
type TemplateCache interface {
	Update() error
	// GetTemplate resolves a discovered template by its flat name, or by a
	// "workspace/name" qualified form to disambiguate cross-workspace collisions.
	GetTemplate(name string) (*executable.Template, error)
	GetTemplateList() (executable.TemplateList, error)
}

type TemplateCacheData struct {
	// Map of discovered template name to its file path
	TemplateMap map[string]string `json:"templateMap" yaml:"templateMap"`
	// Map of template file path to its owning workspace / workspace path
	LocationMap map[string]WorkspaceInfo `json:"locationMap" yaml:"locationMap"`

	loadedTemplates map[string]*executable.Template
}

type TemplateCacheImpl struct {
	Data           *TemplateCacheData `json:",inline" yaml:",inline"`
	WorkspaceCache WorkspaceCache     `json:"-"       yaml:"-"`
	Store          store.DataStore
}

func NewTemplateCache(wsCache WorkspaceCache, s store.DataStore) TemplateCache {
	return &TemplateCacheImpl{
		Store: s,
		Data: &TemplateCacheData{
			TemplateMap: make(map[string]string),
			LocationMap: make(map[string]WorkspaceInfo),
		},
		WorkspaceCache: wsCache,
	}
}

func (c *TemplateCacheImpl) Update() error {
	logger.Log().Debugf("Updating template cache data")
	wsCacheData, err := c.WorkspaceCache.GetLatestData()
	if err != nil {
		return fmt.Errorf("failed to get workspace cache data\n%w", err)
	}

	cacheData := c.Data
	for name, wsCfg := range wsCacheData.Workspaces {
		wsCfg.SetContext(name, wsCacheData.WorkspaceLocations[name])
		templates, err := filesystem.LoadWorkspaceFlowFileTemplates(wsCfg)
		if err != nil {
			logger.Log().Error("failed to load workspace templates", "workspace", wsCfg.AssignedName(), "err", err)
			continue
		}
		for _, tmpl := range templates {
			if vErr := tmpl.Validate(); vErr != nil {
				logger.Log().Warn(
					"invalid template found during cache update",
					"template", tmpl.Name(),
					"path", tmpl.Location(),
					"workspace", wsCfg.AssignedName(),
					"err", vErr,
				)
				continue
			}

			if existingPath, exists := cacheData.TemplateMap[tmpl.Name()]; exists && existingPath != tmpl.Location() {
				logger.Log().Warn(
					"duplicate template name found during cache update; "+
						"use a workspace/name qualified reference to disambiguate",
					"template", tmpl.Name(),
					"conflictPath", existingPath,
					"newPath", tmpl.Location(),
					"workspace", wsCfg.AssignedName(),
				)
			}

			cacheData.TemplateMap[tmpl.Name()] = tmpl.Location()
			cacheData.LocationMap[tmpl.Location()] = WorkspaceInfo{
				WorkspaceName: wsCfg.AssignedName(),
				WorkspacePath: wsCfg.Location(),
			}
		}
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return errors.Wrap(err, "unable to encode template cache data")
	}

	if err := c.Store.SetCacheEntry(tmplCacheKey, data); err != nil {
		return errors.Wrap(err, "unable to write template cache data")
	}

	logger.Log().Debug("Successfully updated template cache data", "count", len(cacheData.TemplateMap))
	return nil
}

func (c *TemplateCacheImpl) GetTemplate(name string) (*executable.Template, error) {
	if err := c.initTemplateCacheData(); err != nil {
		return nil, err
	} else if c.Data == nil {
		return nil, errors.New("no cached templates found")
	}

	if c.Data.loadedTemplates == nil {
		c.Data.loadedTemplates = make(map[string]*executable.Template)
	} else if tmpl, found := c.Data.loadedTemplates[name]; found {
		return tmpl, nil
	}

	path, err := c.resolvePath(name)
	if err != nil {
		return nil, err
	}

	// The unqualified name derived from a workspace/name reference is the trailing segment.
	loadName := name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		loadName = name[idx+1:]
	}

	tmpl, err := filesystem.LoadFlowFileTemplate(loadName, path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load template")
	}
	c.Data.loadedTemplates[name] = tmpl
	return tmpl, nil
}

// resolvePath maps a template reference to a discovered file path. A bare name is looked up
// directly; a "workspace/name" reference is matched against the owning workspace so callers
// can disambiguate the same template name discovered in multiple workspaces.
func (c *TemplateCacheImpl) resolvePath(name string) (string, error) {
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		wsName, tmplName := name[:idx], name[idx+1:]
		for path, wsInfo := range c.Data.LocationMap {
			if wsInfo.WorkspaceName != wsName {
				continue
			}
			// The flat map may point elsewhere (last-writer-wins on collisions), so derive
			// each candidate's name from its file to find the one in this workspace.
			loaded, err := filesystem.LoadFlowFileTemplate("", path)
			if err == nil && loaded.Name() == tmplName {
				return path, nil
			}
		}
		return "", fmt.Errorf("template %s not found", name)
	}

	path, found := c.Data.TemplateMap[name]
	if !found {
		return "", fmt.Errorf("template %s not found", name)
	}
	return path, nil
}

func (c *TemplateCacheImpl) GetTemplateList() (executable.TemplateList, error) {
	if err := c.initTemplateCacheData(); err != nil {
		return nil, err
	} else if c.Data == nil {
		return nil, errors.New("no cached templates found")
	}

	list := make(executable.TemplateList, 0, len(c.Data.TemplateMap))
	for _, name := range slices.Sorted(maps.Keys(c.Data.TemplateMap)) {
		path := c.Data.TemplateMap[name]
		tmpl, err := filesystem.LoadFlowFileTemplate(name, path)
		if err != nil {
			logger.Log().Error("unable to load template", "path", path, "err", err)
			continue
		}
		list = append(list, tmpl)
	}
	return list, nil
}

func (c *TemplateCacheImpl) initTemplateCacheData() error {
	cacheData, err := c.Store.GetCacheEntry(tmplCacheKey)
	if err != nil {
		return errors.Wrap(err, "unable to load template cache data")
	}

	if cacheData == nil {
		if err := c.Update(); err != nil {
			return errors.Wrap(err, "unable to update template cache data")
		}
		return nil
	}

	c.Data = &TemplateCacheData{}
	if err := json.Unmarshal(cacheData, c.Data); err != nil {
		return errors.Wrap(err, "unable to decode template cache data")
	}
	return nil
}
