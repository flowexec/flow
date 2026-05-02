package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/themes"
	"github.com/pkg/errors"

	"github.com/flowexec/flow/v2/internal/version"
	"github.com/flowexec/flow/v2/pkg/cache"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/config"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

const HeaderCtxKey = "ctx"

type Context struct {
	appName               string
	ctx                   context.Context
	cancelFunc            context.CancelFunc
	stdOut, stdIn, stdErr *os.File
	callbacks             []func(*Context) error
	tuiOnce               sync.Once
	tuiContainer          *tuikit.Container

	Config           *config.Config
	CurrentWorkspace *workspace.Workspace
	WorkspacesCache  cache.WorkspaceCache
	ExecutableCache  cache.ExecutableCache
	DataStore        store.DataStore

	// RootExecutable is the executable that is being run in the current context.
	// This will be nil if the context is not associated with an executable run.
	RootExecutable *executable.Executable

	// ProcessTmpDir is the temporary directory for the current process. If set, it will be
	// used to store temporary files all executable runs when the tmpDir value is specified.
	ProcessTmpDir string

	// CurrentTask holds the task context for the currently executing step in a
	// parallel or serial runner. It is set per-goroutine (via shallow copy) so
	// that downstream writers can prefix output with the task name.
	CurrentTask *io.TaskContext

	// LogArchiveID is the unique identifier used in the log archive filename for
	// this process. It is set at startup and used to link execution records to
	// their log output.
	LogArchiveID string
}

// Option configures optional fields on a Context during construction.
type Option func(*Context)

// WithStdIn sets the standard input file. Defaults to os.Stdin.
func WithStdIn(f *os.File) Option { return func(c *Context) { c.stdIn = f } }

// WithStdOut sets the standard output file. Defaults to os.Stdout.
func WithStdOut(f *os.File) Option { return func(c *Context) { c.stdOut = f } }

// WithStdErr sets the standard error file. Defaults to os.Stderr.
func WithStdErr(f *os.File) Option { return func(c *Context) { c.stdErr = f } }

func WithAppName(name string) Option { return func(c *Context) { c.appName = name } }

func NewContext(ctx context.Context, cancelFunc context.CancelFunc, opts ...Option) *Context {
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
	var wsConfig *workspace.Workspace
	if len(cfg.Workspaces) > 0 {
		wsConfig, err = currentWorkspace(cfg)
		if err != nil {
			panic(errors.Wrap(err, "workspace config load error"))
		} else if wsConfig == nil {
			panic(fmt.Errorf("workspace config not found in current workspace (%s)", cfg.CurrentWorkspace))
		}
	}

	ds, err := store.NewDataStore(store.Path())
	if err != nil {
		panic(errors.Wrap(err, "data store initialization error"))
	}

	workspaceCache := cache.NewWorkspaceCache(ds)
	executableCache := cache.NewExecutableCache(workspaceCache, ds)

	c := &Context{
		appName:          "flow",
		ctx:              ctx,
		cancelFunc:       cancelFunc,
		stdOut:           os.Stdout,
		stdIn:            os.Stdin,
		stdErr:           os.Stderr,
		Config:           cfg,
		CurrentWorkspace: wsConfig,
		WorkspacesCache:  workspaceCache,
		ExecutableCache:  executableCache,
		DataStore:        ds,
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ShallowCopy returns a pointer to a new Context that shares all backing
// state (config, caches, data store, TUI container) but has its own mutable
// fields (CurrentTask, ProcessTmpDir, etc.)
func (ctx *Context) ShallowCopy() *Context {
	cp := &Context{
		appName:          ctx.appName,
		ctx:              ctx.ctx,
		cancelFunc:       ctx.cancelFunc,
		stdOut:           ctx.stdOut,
		stdIn:            ctx.stdIn,
		stdErr:           ctx.stdErr,
		tuiContainer:     ctx.tuiContainer, // share already-initialized container (if any)
		Config:           ctx.Config,
		CurrentWorkspace: ctx.CurrentWorkspace,
		WorkspacesCache:  ctx.WorkspacesCache,
		ExecutableCache:  ctx.ExecutableCache,
		DataStore:        ctx.DataStore,
		RootExecutable:   ctx.RootExecutable,
		ProcessTmpDir:    ctx.ProcessTmpDir,
		LogArchiveID:     ctx.LogArchiveID,
	}
	// If the parent has already initialized the TUI container, mark the copy
	// as initialized too so it won't re-create one.
	if ctx.tuiContainer != nil {
		cp.tuiOnce.Do(func() {})
	}
	return cp
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
	var ws string
	if ctx.CurrentWorkspace != nil {
		ws = ctx.CurrentWorkspace.AssignedName()
	}
	ns := ctx.Config.CurrentNamespace
	if ws == "" {
		ws = "unk"
	}
	if ns == "" {
		ns = executable.WildcardNamespace
	}
	return fmt.Sprintf("%s/%s", ws, ns)
}

func (ctx *Context) AppName() string {
	return ctx.appName
}

func (ctx *Context) StdOut() *os.File {
	return ctx.stdOut
}

func (ctx *Context) StdIn() *os.File {
	return ctx.stdIn
}

// StdErr returns the standard error file for structured error envelopes.
// Falls back to os.Stderr when unset (e.g. in tests that bypass NewContext).
func (ctx *Context) StdErr() *os.File {
	if ctx.stdErr == nil {
		return os.Stderr
	}
	return ctx.stdErr
}

// SetIO sets the standard input and output for the context
// This function should NOT be used outside of tests! The standard input and output
// should be set when creating the context.
func (ctx *Context) SetIO(stdIn, stdOut *os.File) {
	ctx.stdIn = stdIn
	ctx.stdOut = stdOut
}

// SetStdErr sets the standard error file. Intended for tests.
func (ctx *Context) SetStdErr(stdErr *os.File) {
	ctx.stdErr = stdErr
}

// SetContext sets the context and cancel function for the Context.
// This function should NOT be used outside of tests! The context and cancel function
// should be set when creating the context.
func (ctx *Context) SetContext(c context.Context, cancelFunc context.CancelFunc) {
	ctx.ctx = c
	ctx.cancelFunc = cancelFunc
}

// TUIContainer returns the TUI container, initializing it on first access.
// This avoids eagerly creating a bubbletea program (which reads terminal state
// from stdin) for commands that never use the TUI.
func (ctx *Context) TUIContainer() *tuikit.Container {
	ctx.tuiOnce.Do(func() {
		app := tuikit.NewApplication(
			ctx.appName,
			tuikit.WithState(HeaderCtxKey, ctx.String()),
			tuikit.WithVersion(version.Short()),
			tuikit.WithLoadingMsg("thinking..."),
		)

		theme := logger.Theme(ctx.Config.Theme.String())
		if ctx.Config.ColorOverride != nil {
			theme = overrideThemeColor(theme, ctx.Config.ColorOverride)
		}

		var err error
		ctx.tuiContainer, err = tuikit.NewContainer(
			ctx.ctx, app,
			tuikit.WithInput(ctx.stdIn),
			tuikit.WithOutput(ctx.stdOut),
			tuikit.WithTheme(theme),
		)
		if err != nil {
			panic(errors.Wrap(err, "TUI container initialization error"))
		}
	})
	return ctx.tuiContainer
}

// SetTUIContainer sets the TUI container directly, bypassing lazy init.
// This is intended for tests that provide a pre-built container.
func (ctx *Context) SetTUIContainer(c *tuikit.Container) {
	ctx.tuiOnce.Do(func() {}) // mark as initialized
	ctx.tuiContainer = c
}

func (ctx *Context) SetView(view tuikit.View) error {
	return ctx.TUIContainer().SetView(view)
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
			logger.Log().WrapError(err, "callback execution error")
		}
	}

	if ctx.ProcessTmpDir != "" {
		files, err := filepath.Glob(filepath.Join(ctx.ProcessTmpDir, "*"))
		if err != nil {
			logger.Log().WrapError(err, fmt.Sprintf("unable to list files in temp dir %s", ctx.ProcessTmpDir))
			return
		}
		for _, f := range files {
			err = os.RemoveAll(f)
			if err != nil {
				logger.Log().WrapError(err, fmt.Sprintf("unable to remove file %s", f))
			}
		}
		if err := os.Remove(ctx.ProcessTmpDir); err != nil {
			logger.Log().WrapError(err, fmt.Sprintf("unable to remove temp dir %s", ctx.ProcessTmpDir))
		}
	}
}

func ExpandRef(ctx *Context, ref executable.Ref) executable.Ref {
	id := ref.ID()
	ws, ns, name := executable.MustParseExecutableID(id)
	if (ws == "" || ws == executable.WildcardWorkspace) && ctx.CurrentWorkspace != nil {
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
