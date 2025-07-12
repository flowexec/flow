//nolint:gocritic
package library

import (
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flowexec/tuikit/types"

	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

func (l *Library) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	cmds = append(
		cmds,
		tea.SetWindowTitle("flow library"),
		tea.Tick(time.Millisecond*250, func(t time.Time) tea.Msg {
			return types.Tick()
		}),
	)
	cmds = append(
		cmds,
		l.paneZeroViewport.Init(),
		l.paneOneViewport.Init(),
		l.paneTwoViewport.Init(),
	)

	if l.ctx.TUIContainer.Width() >= 150 {
		l.splitView = true
	}
	l.setSize()
	go func() {
		l.setVisibleWorkspaces()
		l.setVisibleNamespaces()
		l.setVisibleExecs()
	}()

	return tea.Batch(cmds...)
}

func (l *Library) setVisibleExecs() {
	if len(l.allExecutables) == 0 || len(l.visibleWorkspaces) == 0 {
		return
	}

	curWs := l.filter.Workspace

	l.mu.RLock()
	if label := l.visibleWorkspaces[l.currentWorkspace]; label != "" && label != allWorkspacesLabel {
		curWs = label
	} else if curWs == allWorkspacesLabel {
		curWs = ""
	}
	l.mu.RUnlock()

	curNs := l.filter.Namespace
	if l.showNamespaces && len(l.visibleNamespaces) > 0 {
		l.mu.RLock()
		if label := l.visibleNamespaces[l.currentNamespace]; label != "" {
			switch label {
			case withoutNamespaceLabel:
				curNs = ""
			case allNamespacesLabel:
				curNs = executable.WildcardNamespace
			default:
				curNs = label
			}
		}
		l.mu.RUnlock()
	}

	filter := l.filter
	filteredExec := l.allExecutables
	filteredExec = filteredExec.
		FilterByWorkspace(curWs).
		FilterByNamespace(curNs).
		FilterByVerb(filter.Verb).
		FilterByTags(filter.Tags).
		FilterBySubstring(filter.Substring)

	slices.SortFunc(filteredExec, func(i, j *executable.Executable) int {
		return strings.Compare(i.Ref().String(), j.Ref().String())
	})

	l.mu.Lock()
	l.visibleExecutables = filteredExec
	l.mu.Unlock()
}

func (l *Library) setVisibleWorkspaces() {
	if l.allWorkspaces == nil {
		return
	}

	filter := l.filter
	filteredWs := l.allWorkspaces
	switch filter.Workspace {
	case "":
		// do nothing
	case allWorkspacesLabel, executable.WildcardWorkspace:
		// do nothing
	default:
		for _, ws := range filteredWs {
			if ws.AssignedName() == filter.Workspace {
				filteredWs = workspace.WorkspaceList{ws}
				break
			}
		}
	}

	var labels, prepend []string
	if len(filteredWs) > 1 {
		prepend = []string{allWorkspacesLabel}
	}
	for _, ws := range filteredWs {
		labels = append(labels, ws.AssignedName())
	}
	slices.Sort(labels)

	l.mu.Lock()
	l.visibleWorkspaces = append(prepend, labels...) //nolint:gocritic
	l.mu.Unlock()
}

func (l *Library) setVisibleNamespaces() {
	if l.allExecutables == nil || len(l.visibleWorkspaces) == 0 {
		return
	}

	var labels, prepend []string
	var someWithoutNs bool
	filter := l.filter

	l.mu.RLock()
	filterWs := l.visibleWorkspaces[l.currentWorkspace]
	l.mu.RUnlock()

	nsSet := make(map[string]struct{})
	for _, ex := range l.allExecutables {
		ns := ex.Ref().Namespace()
		ws := ex.Ref().Workspace()
		if filter.Namespace != executable.WildcardNamespace && filter.Namespace != "" && ns != filter.Namespace {
			continue
		} else if filterWs != allWorkspacesLabel && filterWs != "" && ws != filterWs {
			continue
		} else if ns == "" || ns == executable.WildcardNamespace {
			someWithoutNs = true
			continue
		}

		if _, ok := nsSet[ns]; !ok {
			nsSet[ns] = struct{}{}
			labels = append(labels, ns)
		}
	}
	slices.Sort(labels)
	if len(labels) > 1 {
		prepend = append(prepend, allNamespacesLabel)
	}
	if someWithoutNs {
		prepend = append(prepend, withoutNamespaceLabel)
	}

	l.mu.Lock()
	l.visibleNamespaces = append(prepend, labels...) //nolint:gocritic
	l.mu.Unlock()
}
