package library

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

const (
	appName = "flow library"
)

type Library struct {
	ctx                                 *context.Context
	termWidth, termHeight               int
	noticeText                          string
	showHelp, showNamespaces, splitView bool

	visibleWorkspaces  []string
	visibleNamespaces  []string
	visibleExecutables executable.ExecutableList
	allWorkspaces      workspace.WorkspaceList
	allExecutables     executable.ExecutableList
	filter             Filter
	theme              themes.Theme

	currentPane, currentWorkspace, currentNamespace, currentExecutable uint
	currentFormat, currentHelpPage                                     uint
	paneZeroViewport, paneOneViewport, paneTwoViewport                 viewport.Model

	cmdRunFunc func(string) error

	// Mutex to protect concurrent access to shared fields
	mu sync.RWMutex
}

type Filter struct {
	Workspace, Namespace string
	Verb                 executable.Verb
	Tags                 common.Tags
	Substring            string
}

func NewLibrary(
	ctx *context.Context,
	workspaces workspace.WorkspaceList,
	execs executable.ExecutableList,
	filter Filter,
	theme themes.Theme,
	runFunc func(string) error,
) *Library {
	p1 := viewport.New(0, 0)
	p2 := viewport.New(0, 0)
	p3 := viewport.New(0, 0)
	return &Library{
		ctx:                ctx,
		allWorkspaces:      workspaces,
		allExecutables:     execs,
		filter:             filter,
		paneZeroViewport:   p1,
		paneOneViewport:    p2,
		paneTwoViewport:    p3,
		theme:              theme,
		cmdRunFunc:         runFunc,
		visibleWorkspaces:  make([]string, 0),
		visibleNamespaces:  make([]string, 0),
		visibleExecutables: make(executable.ExecutableList, 0),
	}
}

func NewLibraryView(
	ctx *context.Context,
	workspaces workspace.WorkspaceList,
	execs executable.ExecutableList,
	filter Filter,
	theme themes.Theme,
	runFunc func(string) error,
) tuikit.View {
	l := NewLibrary(ctx, workspaces, execs, filter, theme, runFunc)
	return views.NewFrameView(l)
}

func ctxVal(ws, ns string) string {
	if ws == "" {
		ws = "unk"
	}
	if ns == "" {
		ns = executable.WildcardNamespace
	}
	return fmt.Sprintf("%s/%s", ws, ns)
}
