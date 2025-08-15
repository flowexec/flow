package cache

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/fileparser"
	"github.com/flowexec/flow/internal/filesystem"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

const execCacheKey = "executables"

//go:generate mockgen -destination=mocks/mock_executable_cache.go -package=mocks github.com/flowexec/flow/internal/cache ExecutableCache
type ExecutableCache interface {
	Update() error
	GetExecutableByRef(ref executable.Ref) (*executable.Executable, error)
	GetExecutableList() (executable.ExecutableList, error)
}
type WorkspaceInfo struct {
	WorkspaceName string `json:"workspaceName" yaml:"workspaceName"`
	WorkspacePath string `json:"workspacePath" yaml:"workspacePath"`
}

type ExecutableCacheData struct {
	// Map of executable ref to config path
	ExecutableMap map[executable.Ref]string `json:"executableMap" yaml:"executableMap"`
	// Map of executable alias ref to primary executable ref
	AliasMap map[executable.Ref]executable.Ref `json:"aliasMap" yaml:"aliasMap"`
	// Map of config paths to their workspace / workspace path
	ConfigMap map[string]WorkspaceInfo `json:"configMap" yaml:"configMap"`

	loadedExecutables map[string]*executable.Executable
}

type ExecutableCacheImpl struct {
	Data           *ExecutableCacheData `json:",inline" yaml:",inline"`
	WorkspaceCache WorkspaceCache       `json:"-"       yaml:"-"`
}

func NewExecutableCache(wsCache WorkspaceCache) ExecutableCache {
	return &ExecutableCacheImpl{
		Data: &ExecutableCacheData{
			ExecutableMap: make(map[executable.Ref]string),
			AliasMap:      make(map[executable.Ref]executable.Ref),
			ConfigMap:     make(map[string]WorkspaceInfo),
		},
		WorkspaceCache: wsCache,
	}
}

func (c *ExecutableCacheImpl) Update() error { //nolint:gocognit
	logger.Log().Debugf("Updating executable cache data")
	wsCacheData, err := c.WorkspaceCache.GetLatestData()
	if err != nil {
		return fmt.Errorf("failed to get workspace cache data\n%w", err)
	}

	cacheData := c.Data
	for name, wsCfg := range wsCacheData.Workspaces {
		wsCfg.SetContext(name, wsCacheData.WorkspaceLocations[name])
		flowFiles, err := filesystem.LoadWorkspaceFlowFiles(wsCfg)
		if err != nil {
			logger.Log().Errorx("failed to load workspace executable configs", "workspace", wsCfg.AssignedName(), "err", err)
			continue
		}
		for _, flowFile := range flowFiles {
			if len(flowFile.FromFile) > 0 || len(flowFile.Imports) > 0 {
				generated, err := fileparser.ExecutablesFromImports(name, flowFile)
				if err != nil {
					logger.Log().Errorx(
						"failed to generate executables from files",
						"flowFilePath", flowFile.ConfigPath(),
						"err", err,
					)
				}
				flowFile.Executables = append(flowFile.Executables, generated...)
			}

			if flowFile.Visibility == nil ||
				common.Visibility(*flowFile.Visibility).IsHidden() ||
				len(flowFile.Executables) == 0 {
				continue
			}
			for _, e := range flowFile.Executables {
				if e == nil || (e.Visibility != nil && common.Visibility(*e.Visibility).IsHidden()) {
					continue
				}

				if existingPath, exists := cacheData.ExecutableMap[e.Ref()]; exists && existingPath != flowFile.ConfigPath() {
					logger.Log().Warnx(
						"duplicate executable found during cache update",
						"ref", e.Ref().String(),
						"conflictPath", existingPath,
						"newPath", flowFile.ConfigPath(),
						"workspace", wsCfg.AssignedName(),
					)
				}

				cacheData.ExecutableMap[e.Ref()] = flowFile.ConfigPath()

				for _, ref := range enumerateExecutableAliasRefs(e, wsCfg.VerbAliases) {
					if existingPrimaryRef, exists := cacheData.AliasMap[ref]; exists && existingPrimaryRef != e.Ref() {
						logger.Log().Warnx(
							"duplicate executable alias found during cache update",
							"aliasRef", ref.String(),
							"conflictRef", existingPrimaryRef.String(),
							"primaryRef", e.Ref().String(),
							"workspace", wsCfg.AssignedName(),
						)
					}
					cacheData.AliasMap[ref] = e.Ref()
				}

				cacheData.ConfigMap[flowFile.ConfigPath()] = WorkspaceInfo{
					WorkspaceName: wsCfg.AssignedName(),
					WorkspacePath: wsCfg.Location(),
				}
			}
		}
	}

	data, err := yaml.Marshal(cacheData)
	if err != nil {
		return errors.Wrap(err, "unable to encode cache data")
	}

	err = filesystem.WriteLatestCachedData(execCacheKey, data)
	if err != nil {
		return errors.Wrap(err, "unable to write cache data")
	}

	logger.Log().Debugx("Successfully updated executable cache data", "count", len(cacheData.ExecutableMap))
	return nil
}

func (c *ExecutableCacheImpl) GetExecutableByRef(ref executable.Ref) (*executable.Executable, error) {
	err := c.initExecutableCacheData()
	if err != nil {
		return nil, err
	} else if c.Data == nil {
		return nil, errors.New("no cached executables found")
	}

	if c.Data.loadedExecutables == nil {
		c.Data.loadedExecutables = make(map[string]*executable.Executable)
	} else if exec, found := c.Data.loadedExecutables[ref.String()]; found {
		return exec, nil
	}

	var primaryRef executable.Ref
	cfgPath, found := c.Data.ExecutableMap[ref]
	//nolint:nestif
	if !found {
		if aliasedPrimaryRef, aliasFound := c.Data.AliasMap[ref]; aliasFound {
			primaryRef = aliasedPrimaryRef
			cfgPath, found = c.Data.ExecutableMap[primaryRef]
			if !found {
				return nil, NewExecutableNotFoundError(ref.String())
			}
		} else {
			return nil, NewExecutableNotFoundError(ref.String())
		}
	} else {
		primaryRef = ref
	}

	cfg, err := filesystem.LoadFlowFile(cfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load executable config")
	}

	wsInfo, found := c.Data.ConfigMap[cfgPath]
	if !found {
		return nil, errors.Wrap(err, "unable to find workspace info for config")
	}

	cfg.SetDefaults()
	cfg.SetContext(wsInfo.WorkspaceName, wsInfo.WorkspacePath, cfgPath)

	generated, err := fileparser.ExecutablesFromImports(wsInfo.WorkspaceName, cfg)
	if err != nil {
		logger.Log().Warnx(
			"failed to generate executables from files",
			"cfgPath", cfgPath,
			"err", err,
		)
	}
	cfg.Executables = append(cfg.Executables, generated...)

	execs := cfg.Executables
	exec, err := execs.FindByVerbAndID(primaryRef.Verb(), primaryRef.ID())
	if err != nil {
		return nil, err
	} else if exec == nil {
		return nil, NewExecutableNotFoundError(ref.String())
	}

	c.Data.loadedExecutables[ref.String()] = exec

	return exec, nil
}

func (c *ExecutableCacheImpl) GetExecutableList() (executable.ExecutableList, error) {
	err := c.initExecutableCacheData()
	if err != nil {
		return nil, err
	} else if c.Data == nil {
		return nil, errors.New("no cached executables found")
	}

	list := make(executable.ExecutableList, 0)
	for cfgPath := range c.Data.ConfigMap {
		cfg, err := filesystem.LoadFlowFile(cfgPath)
		if err != nil {
			logger.Log().Errorx("unable to load executable config", "cfgPath", cfgPath, "err", err)
			continue
		}
		wsInfo, found := c.Data.ConfigMap[cfgPath]
		if !found {
			logger.Log().Errorx("unable to find workspace info for config", "cfgPath", cfgPath)
			continue
		}
		cfg.SetDefaults()
		cfg.SetContext(wsInfo.WorkspaceName, wsInfo.WorkspacePath, cfgPath)

		generated, err := fileparser.ExecutablesFromImports(wsInfo.WorkspaceName, cfg)
		if err != nil {
			logger.Log().Warnx(
				"failed to generate executables from files",
				"cfgPath", cfgPath,
				"err", err,
			)
		}
		cfg.Executables = append(cfg.Executables, generated...)

		list = append(list, cfg.Executables...)
	}
	return list, nil
}

func (c *ExecutableCacheImpl) initExecutableCacheData() error {
	cacheData, err := filesystem.LoadLatestCachedData(execCacheKey)
	if err != nil {
		return errors.Wrap(err, "unable to load executable cache data")
	} else if cacheData == nil {
		if err := c.Update(); err != nil {
			return errors.Wrap(err, "unable to update executable cache data")
		}
	}

	c.Data = &ExecutableCacheData{}
	if err := yaml.Unmarshal(cacheData, c.Data); err != nil {
		return errors.Wrap(err, "unable to decode executable cache data")
	}
	return nil
}

func enumerateExecutableAliasRefs(
	exec *executable.Executable,
	override *workspace.WorkspaceVerbAliases,
) executable.RefList {
	refs := make(executable.RefList, 0)

	for _, v := range exec.VerbAliases {
		if err := v.Validate(); err != nil {
			continue
		}
		refs = append(refs, executable.NewRef(exec.ID(), v))
		for _, id := range exec.AliasesIDs() {
			refs = append(refs, executable.NewRef(id, v))
		}
	}

	switch {
	case override == nil:
		// use default aliases
		for _, verb := range executable.RelatedVerbs(exec.Verb) {
			refs = append(refs, executable.NewRef(exec.ID(), verb))
			for _, id := range exec.AliasesIDs() {
				refs = append(refs, executable.NewRef(id, verb))
			}
		}
	case len(*override) == 0:
		// disable all aliases if override is set but empty
		return refs
	default:
		// use overrides if provided
		o := *override
		if verbs, found := o[exec.Verb.String()]; found {
			for _, v := range verbs {
				vv := executable.Verb(v)
				if err := vv.Validate(); err != nil {
					// If the verb is not valid, skip it
					continue
				}
				refs = append(refs, executable.NewRef(exec.ID(), vv))
				for _, id := range exec.AliasesIDs() {
					refs = append(refs, executable.NewRef(id, vv))
				}
			}
		}
	}

	return refs
}
