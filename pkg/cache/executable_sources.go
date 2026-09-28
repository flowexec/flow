package cache

import (
	"cmp"
	"maps"
	"slices"
	"sync"

	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/common"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

// ExecutableSource contributes executables for a workspace from files flow does not read itself.
//
// Each executable should carry its namespace (via SetContext) and the file it came from (via SetSourceFilePath).
type ExecutableSource func(ws *workspace.Workspace) (executable.ExecutableList, error)

type SourcedExecutable struct {
	Source        string                 `json:"source"`
	WorkspaceName string                 `json:"workspaceName"`
	WorkspacePath string                 `json:"workspacePath"`
	Namespace     string                 `json:"namespace,omitempty"`
	SourceFile    string                 `json:"sourceFile"`
	Executable    *executable.Executable `json:"executable"`
}

var (
	sourcesMu sync.RWMutex
	sources   = map[string]ExecutableSource{}
)

// RegisterExecutableSource adds src to every executable cache in the process under name,
// replacing any source already registered with that name. Register before the first cache update.
func RegisterExecutableSource(name string, src ExecutableSource) {
	sourcesMu.Lock()
	defer sourcesMu.Unlock()
	sources[name] = src
}

func UnregisterExecutableSource(name string) {
	sourcesMu.Lock()
	defer sourcesMu.Unlock()
	delete(sources, name)
}

func sourceRegistered(name string) bool {
	sourcesMu.RLock()
	defer sourcesMu.RUnlock()
	_, ok := sources[name]
	return ok
}

func registeredSources() map[string]ExecutableSource {
	sourcesMu.RLock()
	defer sourcesMu.RUnlock()
	return maps.Clone(sources)
}

func (s *SourcedExecutable) hydrate() *executable.Executable {
	s.Executable.SetContext(s.WorkspaceName, s.WorkspacePath, s.Namespace, s.SourceFile)
	s.Executable.SetSourceFilePath(s.SourceFile)
	return s.Executable
}

// indexSourcedExecutables adds each registered source's executables for wsCfg to data. It runs
// after the workspace's flow files are indexed, and a flow file's executable wins a ref or alias
// that a source also claims.
func indexSourcedExecutables(data *ExecutableCacheData, wsCfg *workspace.Workspace) {
	name := wsCfg.AssignedName()
	srcs := registeredSources()
	for _, srcName := range slices.Sorted(maps.Keys(srcs)) {
		execs, err := srcs[srcName](wsCfg)
		if err != nil {
			logger.Log().Error("executable source failed", "source", srcName, "workspace", name, "err", err)
			continue
		}
		for _, e := range execs {
			if e == nil {
				continue
			}
			addSourcedExecutable(data, wsCfg, &SourcedExecutable{
				Source:        srcName,
				WorkspaceName: name,
				WorkspacePath: wsCfg.Location(),
				Namespace:     e.Namespace(),
				SourceFile:    e.SourceFilePath(),
				Executable:    e,
			})
		}
	}
}

// carrySourcedExecutables copies entries from prev that belong to sources not registered in this
// process, so an update from such a process doesn't drop them. Entries for workspaces that are no
// longer registered are dropped.
func carrySourcedExecutables(data, prev *ExecutableCacheData, workspaces map[string]*workspace.Workspace) {
	if prev == nil {
		return
	}
	for _, ref := range slices.Sorted(maps.Keys(prev.SourcedExecutables)) {
		s := prev.SourcedExecutables[ref]
		if s == nil || s.Executable == nil || sourceRegistered(s.Source) {
			continue
		}
		wsCfg, ok := workspaces[s.WorkspaceName]
		if !ok {
			continue
		}
		addSourcedExecutable(data, wsCfg, s)
	}
}

func addSourcedExecutable(data *ExecutableCacheData, wsCfg *workspace.Workspace, s *SourcedExecutable) {
	e := s.hydrate()
	if e.Visibility != nil && common.Visibility(*e.Visibility).IsHidden() {
		return
	}
	if err := e.Validate(); err != nil {
		logger.Log().Warn(
			"invalid executable from source",
			"ref", e.Ref().String(),
			"source", s.Source,
			"workspace", s.WorkspaceName,
			"err", err,
		)
		return
	}
	ref := e.Ref()
	if _, exists := data.ExecutableMap[ref]; exists {
		logger.Log().Debug("executable from source shadowed by flow file", "ref", ref.String(), "source", s.Source)
		return
	}
	if existing, exists := data.SourcedExecutables[ref]; exists {
		logger.Log().Warn(
			"duplicate executable from source",
			"ref", ref.String(),
			"source", s.Source,
			"conflictSource", existing.Source,
		)
		return
	}
	data.SourcedExecutables[ref] = s

	for _, alias := range enumerateExecutableAliasRefs(e, wsCfg.VerbAliases) {
		if _, exists := data.AliasMap[alias]; exists {
			continue
		}
		data.AliasMap[alias] = ref
	}
}

func visibleSourced(data *ExecutableCacheData, ref executable.Ref) (*executable.Executable, bool) {
	s, ok := data.SourcedExecutables[ref]
	if !ok || s == nil || s.Executable == nil || !sourceRegistered(s.Source) {
		return nil, false
	}
	return s.hydrate(), true
}

func listSourced(data *ExecutableCacheData) executable.ExecutableList {
	list := make(executable.ExecutableList, 0, len(data.SourcedExecutables))
	for _, s := range data.SourcedExecutables {
		if s != nil && s.Executable != nil && sourceRegistered(s.Source) {
			list = append(list, s.hydrate())
		}
	}
	slices.SortFunc(list, func(a, b *executable.Executable) int {
		return cmp.Or(
			cmp.Compare(a.SourceFilePath(), b.SourceFilePath()),
			cmp.Compare(a.Ref().String(), b.Ref().String()),
		)
	})
	return list
}
