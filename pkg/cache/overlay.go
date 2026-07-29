package cache

import (
	"maps"
	"sync"

	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

// The overlay caches let flow run inside a workspace the user never registered — a git worktree,
// a fresh clone — by indexing that one workspace in memory and layering it over the persisted
// cache.
//
// Nothing here is ever written to the data store. The persisted cache is shared by every flow
// invocation on the machine and is keyed by workspace name; writing an ad-hoc workspace into it
// would let a throwaway clone shadow a real workspace for every future command. The overlay
// lives and dies with the process instead, which is affordable because indexing one workspace
// costs about what the per-command cache refresh already costs.
//
// Where a discovered workspace's name collides with a registered one, the overlay wins. That is
// the same "closest wins" rule that selected it in the first place.

// NewLocalWorkspaceCache layers a discovered workspace over the registered ones so that path
// lookups and workspace listings can see it.
func NewLocalWorkspaceCache(base WorkspaceCache, ws *workspace.Workspace) WorkspaceCache {
	return &localWorkspaceCache{base: base, ws: ws}
}

type localWorkspaceCache struct {
	base WorkspaceCache
	ws   *workspace.Workspace
}

func (c *localWorkspaceCache) Update() error { return c.base.Update() }

func (c *localWorkspaceCache) GetData() *WorkspaceCacheData {
	return c.inject(c.base.GetData())
}

func (c *localWorkspaceCache) GetLatestData() (*WorkspaceCacheData, error) {
	data, err := c.base.GetLatestData()
	if err != nil {
		return nil, err
	}
	return c.inject(data), nil
}

// inject copies the base data before adding the discovered workspace. The base implementation
// marshals its own struct straight into the data store on Update, so mutating it here would
// persist the discovered workspace — exactly what this type exists to avoid.
func (c *localWorkspaceCache) inject(data *WorkspaceCacheData) *WorkspaceCacheData {
	if data == nil {
		data = &WorkspaceCacheData{}
	}
	merged := &WorkspaceCacheData{
		Workspaces:         maps.Clone(data.Workspaces),
		WorkspaceLocations: maps.Clone(data.WorkspaceLocations),
	}
	if merged.Workspaces == nil {
		merged.Workspaces = make(map[string]*workspace.Workspace)
	}
	if merged.WorkspaceLocations == nil {
		merged.WorkspaceLocations = make(map[string]string)
	}
	merged.Workspaces[c.ws.AssignedName()] = c.ws
	merged.WorkspaceLocations[c.ws.AssignedName()] = c.ws.Location()
	return merged
}

func (c *localWorkspaceCache) GetWorkspaceConfigList() (workspace.WorkspaceList, error) {
	base, err := c.base.GetWorkspaceConfigList()
	if err != nil {
		return nil, err
	}
	list := make(workspace.WorkspaceList, 0, len(base)+1)
	for _, ws := range base {
		if ws.AssignedName() != c.ws.AssignedName() {
			list = append(list, ws)
		}
	}
	return append(list, c.ws), nil
}

// NewLocalExecutableCache resolves executables from a discovered workspace first, falling back to
// the persisted cache for every other workspace.
func NewLocalExecutableCache(base ExecutableCache, ws *workspace.Workspace) ExecutableCache {
	return &localExecutableCache{base: base, ws: ws}
}

type localExecutableCache struct {
	base ExecutableCache
	ws   *workspace.Workspace

	mu   sync.Mutex
	data *ExecutableCacheData
}

// index builds the in-memory index on first use. Deferring it keeps commands that never touch
// executables (`flow config`, `flow workspace list`) from walking the workspace tree.
func (c *localExecutableCache) index() *ExecutableCacheData {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data == nil {
		c.data = newExecutableCacheData()
		indexWorkspaceExecutables(c.data, c.ws)
	}
	return c.data
}

func (c *localExecutableCache) Update() error {
	c.mu.Lock()
	c.data = nil
	c.mu.Unlock()
	return c.base.Update()
}

func (c *localExecutableCache) GetExecutableByRef(ref executable.Ref) (*executable.Executable, error) {
	if exec, err := lookupExecutable(c.index(), ref); err == nil {
		return exec, nil
	}
	return c.base.GetExecutableByRef(ref)
}

func (c *localExecutableCache) GetExecutableList() (executable.ExecutableList, error) {
	base, err := c.base.GetExecutableList()
	if err != nil {
		return nil, err
	}
	name := c.ws.AssignedName()
	list := make(executable.ExecutableList, 0, len(base))
	for _, e := range base {
		if e.Workspace() != name {
			list = append(list, e)
		}
	}
	return append(list, listExecutables(c.index())...), nil
}

// NewLocalTemplateCache resolves templates from a discovered workspace first, falling back to the
// persisted cache.
func NewLocalTemplateCache(base TemplateCache, ws *workspace.Workspace) TemplateCache {
	return &localTemplateCache{base: base, ws: ws}
}

type localTemplateCache struct {
	base TemplateCache
	ws   *workspace.Workspace

	mu   sync.Mutex
	data *TemplateCacheData
}

func (c *localTemplateCache) index() *TemplateCacheData {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data == nil {
		c.data = newTemplateCacheData()
		indexWorkspaceTemplates(c.data, c.ws)
	}
	return c.data
}

func (c *localTemplateCache) Update() error {
	c.mu.Lock()
	c.data = nil
	c.mu.Unlock()
	return c.base.Update()
}

func (c *localTemplateCache) GetTemplate(name string) (*executable.Template, error) {
	if tmpl, err := lookupTemplate(c.index(), name); err == nil {
		return tmpl, nil
	}
	return c.base.GetTemplate(name)
}

func (c *localTemplateCache) GetTemplateList() (executable.TemplateList, error) {
	base, err := c.base.GetTemplateList()
	if err != nil {
		return nil, err
	}
	local := listTemplates(c.index())
	localNames := make(map[string]struct{}, len(local))
	for _, tmpl := range local {
		localNames[tmpl.Name()] = struct{}{}
	}
	list := make(executable.TemplateList, 0, len(base)+len(local))
	for _, tmpl := range base {
		if _, shadowed := localNames[tmpl.Name()]; !shadowed {
			list = append(list, tmpl)
		}
	}
	return append(list, local...), nil
}
