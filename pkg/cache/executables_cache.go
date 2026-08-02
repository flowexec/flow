package cache

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/internal/fileparser"
	flowErrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/common"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

const execCacheKey = "executables"

//go:generate mockgen -destination=mocks/mock_executable_cache.go -package=mocks github.com/flowexec/flow/v2/pkg/cache ExecutableCache
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
	Store          store.DataStore
}

func NewExecutableCache(wsCache WorkspaceCache, s store.DataStore) ExecutableCache {
	return &ExecutableCacheImpl{
		Store:          s,
		Data:           newExecutableCacheData(),
		WorkspaceCache: wsCache,
	}
}

// newExecutableCacheData returns an empty, ready-to-populate index.
func newExecutableCacheData() *ExecutableCacheData {
	return &ExecutableCacheData{
		ExecutableMap: make(map[executable.Ref]string),
		AliasMap:      make(map[executable.Ref]executable.Ref),
		ConfigMap:     make(map[string]WorkspaceInfo),
	}
}

// indexWorkspaceExecutables walks one workspace's flow files and records every visible, valid
// executable (and its aliases) in data. wsCfg must already carry its name and location.
//
// This is shared by the persisted cache and the in-memory overlay built for a workspace
// discovered from the working directory, so the two agree on visibility, validation, generated
// imports, and alias expansion.
func indexWorkspaceExecutables(data *ExecutableCacheData, wsCfg *workspace.Workspace) { //nolint:gocognit
	name := wsCfg.AssignedName()
	flowFiles, err := filesystem.LoadWorkspaceFlowFiles(wsCfg)
	if err != nil {
		logger.Log().Error("failed to load workspace executable configs", "workspace", name, "err", err)
		return
	}
	for _, flowFile := range flowFiles {
		if len(flowFile.Imports) > 0 {
			generated, err := fileparser.ExecutablesFromImports(name, flowFile)
			if err != nil {
				logger.Log().Error(
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
			if vErr := e.Validate(); vErr != nil {
				logger.Log().Warn(
					"invalid executable found during cache update",
					"ref", e.Ref().String(),
					"workspace", name,
					"err", vErr,
				)
				continue
			}

			if e == nil || (e.Visibility != nil && common.Visibility(*e.Visibility).IsHidden()) {
				continue
			}

			if existingPath, exists := data.ExecutableMap[e.Ref()]; exists && existingPath != flowFile.ConfigPath() {
				logger.Log().Warn(
					"duplicate executable found during cache update",
					"ref", e.Ref().String(),
					"conflictPath", existingPath,
					"newPath", flowFile.ConfigPath(),
					"workspace", name,
				)
			}

			data.ExecutableMap[e.Ref()] = flowFile.ConfigPath()

			for _, ref := range enumerateExecutableAliasRefs(e, wsCfg.VerbAliases) {
				if existingPrimaryRef, exists := data.AliasMap[ref]; exists && existingPrimaryRef != e.Ref() {
					logger.Log().Warn(
						"duplicate executable alias found during cache update",
						"aliasRef", ref.String(),
						"conflictRef", existingPrimaryRef.String(),
						"primaryRef", e.Ref().String(),
						"workspace", name,
					)
				}
				data.AliasMap[ref] = e.Ref()
			}

			data.ConfigMap[flowFile.ConfigPath()] = WorkspaceInfo{
				WorkspaceName: name,
				WorkspacePath: wsCfg.Location(),
			}
		}
	}
}

func (c *ExecutableCacheImpl) Update() error {
	logger.Log().Debugf("Updating executable cache data")
	wsCacheData, err := c.WorkspaceCache.GetLatestData()
	if err != nil {
		return fmt.Errorf("failed to get workspace cache data\n%w", err)
	}

	cacheData := c.Data
	for name, wsCfg := range wsCacheData.Workspaces {
		wsCfg.SetContext(name, wsCacheData.WorkspaceLocations[name])
		indexWorkspaceExecutables(cacheData, wsCfg)
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return errors.Wrap(err, "unable to encode cache data")
	}

	if err := c.Store.SetCacheEntry(execCacheKey, data); err != nil {
		return errors.Wrap(err, "unable to write cache data")
	}

	logger.Log().Debug("Successfully updated executable cache data", "count", len(cacheData.ExecutableMap))
	return nil
}

// lookupExecutable resolves ref against an index, following the alias map when the ref is not a
// primary one, and loading the owning flow file to return the executable itself.
func lookupExecutable(data *ExecutableCacheData, ref executable.Ref) (*executable.Executable, error) {
	if data.loadedExecutables == nil {
		data.loadedExecutables = make(map[string]*executable.Executable)
	} else if exec, found := data.loadedExecutables[ref.String()]; found {
		return exec, nil
	}

	primaryRef := ref
	cfgPath, found := data.ExecutableMap[ref]
	if !found {
		aliasedPrimaryRef, aliasFound := data.AliasMap[ref]
		if !aliasFound {
			return nil, flowErrors.NewExecutableNotFoundError(ref.String())
		}
		primaryRef = aliasedPrimaryRef
		if cfgPath, found = data.ExecutableMap[primaryRef]; !found {
			return nil, flowErrors.NewExecutableNotFoundError(ref.String())
		}
	}

	wsInfo, found := data.ConfigMap[cfgPath]
	if !found {
		return nil, errors.Errorf("unable to find workspace info for config %s", cfgPath)
	}

	cfg, err := loadFlowFileWithImports(cfgPath, wsInfo)
	if err != nil {
		return nil, err
	}

	exec, err := cfg.Executables.FindByVerbAndID(primaryRef.Verb(), primaryRef.ID())
	if err != nil {
		return nil, err
	} else if exec == nil {
		return nil, flowErrors.NewExecutableNotFoundError(ref.String())
	}

	data.loadedExecutables[ref.String()] = exec

	return exec, nil
}

// listExecutables returns every executable in an index, ordered by flow file path. Callers
// paginate this list across separate calls, so map order would silently drop entries.
func listExecutables(data *ExecutableCacheData) executable.ExecutableList {
	list := make(executable.ExecutableList, 0)
	for _, cfgPath := range slices.Sorted(maps.Keys(data.ConfigMap)) {
		cfg, err := loadFlowFileWithImports(cfgPath, data.ConfigMap[cfgPath])
		if err != nil {
			logger.Log().Error("unable to load executable config", "cfgPath", cfgPath, "err", err)
			continue
		}
		list = append(list, cfg.Executables...)
	}
	return list
}

// loadFlowFileWithImports reads a flow file, attaches its workspace context, and appends the
// executables generated from its imports.
func loadFlowFileWithImports(cfgPath string, wsInfo WorkspaceInfo) (*executable.FlowFile, error) {
	cfg, err := filesystem.LoadFlowFile(cfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load executable config")
	}
	cfg.SetDefaults()
	cfg.SetContext(wsInfo.WorkspaceName, wsInfo.WorkspacePath, cfgPath)

	generated, err := fileparser.ExecutablesFromImports(wsInfo.WorkspaceName, cfg)
	if err != nil {
		logger.Log().Warn("failed to generate executables from files", "cfgPath", cfgPath, "err", err)
	}
	cfg.Executables = append(cfg.Executables, generated...)
	return cfg, nil
}

func (c *ExecutableCacheImpl) GetExecutableByRef(ref executable.Ref) (*executable.Executable, error) {
	err := c.initExecutableCacheData()
	if err != nil {
		return nil, err
	} else if c.Data == nil {
		return nil, errors.New("no cached executables found")
	}
	return lookupExecutable(c.Data, ref)
}

func (c *ExecutableCacheImpl) GetExecutableList() (executable.ExecutableList, error) {
	err := c.initExecutableCacheData()
	if err != nil {
		return nil, err
	} else if c.Data == nil {
		return nil, errors.New("no cached executables found")
	}
	return listExecutables(c.Data), nil
}

func (c *ExecutableCacheImpl) initExecutableCacheData() error {
	cacheData, err := c.Store.GetCacheEntry(execCacheKey)
	if err != nil {
		return errors.Wrap(err, "unable to load executable cache data")
	}

	if cacheData == nil {
		// Lazy-migrate from legacy YAML file if it exists.
		cacheData = migrateExecutableCacheFromFile()
		if cacheData != nil {
			if writeErr := c.Store.SetCacheEntry(execCacheKey, cacheData); writeErr != nil {
				logger.Log().Warn("failed to persist migrated executable cache", "err", writeErr)
			}
		}
	}

	if cacheData == nil {
		if err := c.Update(); err != nil {
			return errors.Wrap(err, "unable to update executable cache data")
		}
		return nil
	}

	c.Data = &ExecutableCacheData{}
	if err := json.Unmarshal(cacheData, c.Data); err != nil {
		return errors.Wrap(err, "unable to decode executable cache data")
	}
	return nil
}

func enumerateExecutableAliasRefs(
	exec *executable.Executable,
	override *workspace.WorkspaceVerbAliases,
) executable.RefList {
	refs := make(executable.RefList, 0)

	// Always include explicit verb aliases defined on the executable itself
	for _, v := range exec.VerbAliases {
		if err := v.Validate(); err != nil {
			continue
		}
		refs = append(refs, executable.NewRef(exec.ID(), v))
		for _, id := range exec.AliasesIDs() {
			refs = append(refs, executable.NewRef(id, v))
		}
	}

	// Always include name (ID) aliases for the executable's primary verb
	for _, id := range exec.AliasesIDs() {
		refs = append(refs, executable.NewRef(id, exec.Verb))
	}

	switch {
	case override == nil:
		// use default aliases (related verbs for the primary verb)
		for _, verb := range executable.RelatedVerbs(exec.Verb) {
			refs = append(refs, executable.NewRef(exec.ID(), verb))
			for _, id := range exec.AliasesIDs() {
				refs = append(refs, executable.NewRef(id, verb))
			}
		}
	case len(*override) == 0:
		// verb overrides explicitly disable verb aliases; keep only name aliases and explicit verb aliases
		// nothing more to add
	default:
		// use overrides if provided
		o := *override
		if verbs, found := o[exec.Verb.String()]; found {
			for _, v := range verbs {
				vv := executable.Verb(v)
				if err := vv.Validate(); err != nil {
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

// migrateExecutableCacheFromFile attempts to read the legacy YAML cache file, re-encodes it
// as JSON, deletes the old file, and returns the JSON bytes. Returns nil if no legacy file
// or if any step fails (best-effort migration).
func migrateExecutableCacheFromFile() []byte {
	legacyPath := filepath.Join(filesystem.CachedDataDirPath(), "latestcache", execCacheKey)
	yamlData, err := os.ReadFile(legacyPath)
	if err != nil {
		return nil
	}

	var legacy ExecutableCacheData
	if err := yaml.Unmarshal(yamlData, &legacy); err != nil {
		return nil
	}

	jsonData, err := json.Marshal(&legacy)
	if err != nil {
		return nil
	}

	_ = os.Remove(legacyPath)
	return jsonData
}
