package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/pkg/errors"

	flowIO "github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/pkg/cache"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/config"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

const (
	AppName      = "flow"
	HeaderCtxKey = "ctx"
)

type Context struct {
	ctx           context.Context
	cancelFunc    context.CancelFunc
	stdOut, stdIn *os.File
	callbacks     []func(*Context) error

	Config           *config.Config
	CurrentWorkspace *workspace.Workspace
	TUIContainer     *tuikit.Container
	WorkspacesCache  cache.WorkspaceCache
	ExecutableCache  cache.ExecutableCache

	// RootExecutable is the executable that is being run in the current context.
	// This will be nil if the context is not associated with an executable run.
	RootExecutable *executable.Executable

	// ProcessTmpDir is the temporary directory for the current process. If set, it will be
	// used to store temporary files all executable runs when the tmpDir value is specified.
	ProcessTmpDir string
}

func NewContext(ctx context.Context, cancelFunc context.CancelFunc, stdIn, stdOut *os.File) *Context {
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		panic(errors.Wrap(err, "user config load error"))
	}

	cfg.SetDefaults()
	if cfg.DefaultTimeout != 0 && os.Getenv(executable.TimeoutOverrideEnv) == "" {
		// HACK: Set the default timeout as an environment variable to be used by the exec runner
		// This is a temporary solution until the config handling is refactored a bit
		_ = os.Setenv(executable.TimeoutOverrideEnv, cfg.DefaultTimeout.String())
	}
	wsConfig, err := currentWorkspace(cfg)
	if err != nil {
		panic(errors.Wrap(err, "workspace config load error"))
	} else if wsConfig == nil {
		panic(fmt.Errorf("workspace config not found in current workspace (%s)", cfg.CurrentWorkspace))
	}

	workspaceCache := cache.NewWorkspaceCache()
	executableCache := cache.NewExecutableCache(workspaceCache)

	c := &Context{
		ctx:              ctx,
		cancelFunc:       cancelFunc,
		stdOut:           stdOut,
		stdIn:            stdIn,
		Config:           cfg,
		CurrentWorkspace: wsConfig,
		WorkspacesCache:  workspaceCache,
		ExecutableCache:  executableCache,
	}

	app := tuikit.NewApplication(
		AppName,
		tuikit.WithState(HeaderCtxKey, c.String()),
		tuikit.WithLoadingMsg("thinking..."),
	)

	theme := flowIO.Theme(cfg.Theme.String())
	if cfg.ColorOverride != nil {
		theme = overrideThemeColor(theme, cfg.ColorOverride)
	}
	c.TUIContainer, err = tuikit.NewContainer(
		ctx, app,
		tuikit.WithInput(stdIn),
		tuikit.WithOutput(stdOut),
		tuikit.WithTheme(theme),
	)
	if err != nil {
		panic(errors.Wrap(err, "TUI container initialization error"))
	}
	return c
}

func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.ctx.Deadline()
}

func (ctx *Context) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx *Context) Err() error {
	return ctx.ctx.Err()
}

func (ctx *Context) Cancel() {
	if ctx.cancelFunc != nil {
		ctx.cancelFunc()
	}
}

// TODO: Move access to various context fields to this function
func (ctx *Context) Value(key any) any {
	if key == HeaderCtxKey {
		return ctx.String()
	}
	return ctx.ctx.Value(key)
}

func (ctx *Context) String() string {
	ws := ctx.CurrentWorkspace.AssignedName()
	ns := ctx.Config.CurrentNamespace
	if ws == "" {
		ws = "unk"
	}
	if ns == "" {
		ns = executable.WildcardNamespace
	}
	return fmt.Sprintf("%s/%s", ws, ns)
}

func (ctx *Context) StdOut() *os.File {
	return ctx.stdOut
}

func (ctx *Context) StdIn() *os.File {
	return ctx.stdIn
}

// SetIO sets the standard input and output for the context
// This function should NOT be used outside of tests! The standard input and output
// should be set when creating the context.
func (ctx *Context) SetIO(stdIn, stdOut *os.File) {
	ctx.stdIn = stdIn
	ctx.stdOut = stdOut
}

// SetContext sets the context and cancel function for the Context.
// This function should NOT be used outside of tests! The context and cancel function
// should be set when creating the context.
func (ctx *Context) SetContext(c context.Context, cancelFunc context.CancelFunc) {
	ctx.ctx = c
	ctx.cancelFunc = cancelFunc
}

func (ctx *Context) SetView(view tuikit.View) error {
	return ctx.TUIContainer.SetView(view)
}

func (ctx *Context) AddCallback(callback func(*Context) error) {
	if callback == nil {
		return
	}
	ctx.callbacks = append(ctx.callbacks, callback)
}

func (ctx *Context) Finalize() {
	_ = ctx.stdIn.Close()
	_ = ctx.stdOut.Close()

	for _, cb := range ctx.callbacks {
		if err := cb(ctx); err != nil {
			logger.Log().Error(err, "callback execution error")
		}
	}

	if ctx.ProcessTmpDir != "" {
		files, err := filepath.Glob(filepath.Join(ctx.ProcessTmpDir, "*"))
		if err != nil {
			logger.Log().Error(err, fmt.Sprintf("unable to list files in temp dir %s", ctx.ProcessTmpDir))
			return
		}
		for _, f := range files {
			err = os.RemoveAll(f)
			if err != nil {
				logger.Log().Error(err, fmt.Sprintf("unable to remove file %s", f))
			}
		}
		if err := os.Remove(ctx.ProcessTmpDir); err != nil {
			logger.Log().Error(err, fmt.Sprintf("unable to remove temp dir %s", ctx.ProcessTmpDir))
		}
	}
}

func ExpandRef(ctx *Context, ref executable.Ref) executable.Ref {
	id := ref.ID()
	ws, ns, name := executable.MustParseExecutableID(id)
	if ws == "" || ws == executable.WildcardWorkspace {
		ws = ctx.CurrentWorkspace.AssignedName()
	}
	if ns == "" {
		ns = ctx.Config.CurrentNamespace
	}
	return executable.NewRef(executable.NewExecutableID(ws, ns, name), ref.Verb())
}

func ExpandRefFromParent(parent *executable.Executable, ref executable.Ref) executable.Ref {
	id := ref.ID()
	ws, ns, name := executable.MustParseExecutableID(id)
	if ws == "" || ws == executable.WildcardWorkspace {
		ws = parent.Workspace()
	}
	if ns == "" {
		ns = parent.Namespace()
	}
	return executable.NewRef(executable.NewExecutableID(ws, ns, name), ref.Verb())
}

func currentWorkspace(cfg *config.Config) (*workspace.Workspace, error) {
	ws, err := cfg.CurrentWorkspaceName()
	if err != nil {
		return nil, err
	}
	wsPath := cfg.Workspaces[ws]
	if ws == "" || wsPath == "" {
		return nil, fmt.Errorf("current workspace not found")
	}

	return filesystem.LoadWorkspaceConfig(ws, wsPath)
}

func overrideThemeColor(theme themes.Theme, palette *config.ColorPalette) themes.Theme {
	if palette == nil {
		return theme
	}
	if palette.Primary != nil {
		theme.ColorPalette().Primary = *palette.Primary
	}
	if palette.Secondary != nil {
		theme.ColorPalette().Secondary = *palette.Secondary
	}
	if palette.Tertiary != nil {
		theme.ColorPalette().Tertiary = *palette.Tertiary
	}
	if palette.Success != nil {
		theme.ColorPalette().Success = *palette.Success
	}
	if palette.Warning != nil {
		theme.ColorPalette().Warning = *palette.Warning
	}
	if palette.Error != nil {
		theme.ColorPalette().Error = *palette.Error
	}
	if palette.Info != nil {
		theme.ColorPalette().Info = *palette.Info
	}
	if palette.Body != nil {
		theme.ColorPalette().Body = *palette.Body
	}
	if palette.Emphasis != nil {
		theme.ColorPalette().Emphasis = *palette.Emphasis
	}
	if palette.White != nil {
		theme.ColorPalette().White = *palette.White
	}
	if palette.Black != nil {
		theme.ColorPalette().Black = *palette.Black
	}
	if palette.Gray != nil {
		theme.ColorPalette().Gray = *palette.Gray
	}
	if palette.CodeStyle != nil {
		theme.ColorPalette().ChromaCodeStyle = *palette.CodeStyle
	}
	return theme
}
