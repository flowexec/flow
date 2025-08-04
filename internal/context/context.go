package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/pkg/errors"

	"github.com/flowexec/flow/internal/cache"
	"github.com/flowexec/flow/internal/filesystem"
	flowIO "github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/config"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

const (
	AppName      = "flow"
	HeaderCtxKey = "ctx"
)

type Context struct {
	Ctx              context.Context
	CancelFunc       context.CancelFunc
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

	stdOut, stdIn *os.File
	callbacks     []func(*Context) error
}

func NewContext(ctx context.Context, stdIn, stdOut *os.File) *Context {
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
	if workspaceCache == nil {
		panic("workspace cache initialization error")
	}
	executableCache := cache.NewExecutableCache(workspaceCache)
	if executableCache == nil {
		panic("executable cache initialization error")
	}

	ctxx, cancel := context.WithCancel(ctx)
	c := &Context{
		Ctx:              ctxx,
		CancelFunc:       cancel,
		Config:           cfg,
		CurrentWorkspace: wsConfig,
		WorkspacesCache:  workspaceCache,
		ExecutableCache:  executableCache,
		stdOut:           stdOut,
		stdIn:            stdIn,
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

func currentWorkspace(cfg *config.Config) (*workspace.Workspace, error) {
	var ws, wsPath string
	mode := cfg.WorkspaceMode

	switch mode {
	case config.ConfigWorkspaceModeDynamic:
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		if runtime.GOOS == "darwin" {
			// On macOS, paths that start with /tmp (and some other system directories)
			// are actually symbolic links to paths under /private. The OS may return
			// either form of the path - e.g., both "/tmp/file" and "/private/tmp/file"
			// refer to the same location. We strip the "/private" prefix for consistent
			// path comparison, while preserving the original paths for filesystem operations.
			wd = strings.TrimPrefix(wd, "/private")
		}

		for wsName, path := range cfg.Workspaces {
			rel, err := filepath.Rel(filepath.Clean(path), filepath.Clean(wd))
			if err != nil {
				return nil, err
			}
			if !strings.HasPrefix(rel, "..") {
				ws = wsName
				wsPath = path
				break
			}
		}
		fallthrough
	case config.ConfigWorkspaceModeFixed:
		if ws != "" && wsPath != "" {
			break
		}
		ws = cfg.CurrentWorkspace
		wsPath = cfg.Workspaces[ws]
	}
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
