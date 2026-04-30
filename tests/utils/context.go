package utils

import (
	stdCtx "context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	tuikitIO "github.com/flowexec/tuikit/io"
	tuikitIOMocks "github.com/flowexec/tuikit/io/mocks"
	"github.com/onsi/ginkgo/v2"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/runner/mocks"
	"github.com/flowexec/flow/pkg/cache"
	cacheMocks "github.com/flowexec/flow/pkg/cache/mocks"
	"github.com/flowexec/flow/pkg/context"
	flowerrors "github.com/flowexec/flow/pkg/errors"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/pkg/store"
	"github.com/flowexec/flow/tests/utils/builder"
	"github.com/flowexec/flow/types/config"
	"github.com/flowexec/flow/types/workspace"
)

const (
	TestWorkspaceName        = "test"
	TestWorkspaceDisplayName = "Test Workspace"

	userConfigSubdir = "config"
	cacheSubdir      = "cache"
)

type Context struct {
	*context.Context
	cacheDir  string
	configDir string
	wsDir     string
	exit      *ExitRecorder
}

func (c *Context) WorkspaceDir() string {
	return c.wsDir
}

// ExpectFailure flips the test context into capture mode: subsequent fatal
// logger calls will be recorded on the context and cause the currently
// executing command to return an error instead of failing the test. Reset
// on every ResetTestContext call.
func (c *Context) ExpectFailure() {
	c.exit.setExpect(true)
}

// ExitCalls returns the fatal messages captured while the context was in
// failure-capture mode.
func (c *Context) ExitCalls() []string {
	return c.exit.snapshot()
}

// ExitRecorder routes logger fatal calls so they can be asserted on in tests
// rather than aborting the process or the test goroutine.
type ExitRecorder struct {
	mu     sync.Mutex
	expect bool
	calls  []string
	tb     testing.TB
}

// fatalExit is the panic value emitted by the recorder when capture mode is on.
// CommandRunner.Run recovers from it and surfaces the message as an error.
type fatalExit struct{ msg string }

func (f fatalExit) String() string { return f.msg }

func newExitRecorder(tb testing.TB) *ExitRecorder {
	return &ExitRecorder{tb: tb}
}

func (r *ExitRecorder) setTB(tb testing.TB) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tb = tb
	r.expect = false
	r.calls = nil
}

func (r *ExitRecorder) setExpect(v bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expect = v
	if v {
		r.calls = nil
	}
}

func (r *ExitRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.calls...)
}

func (r *ExitRecorder) exit(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	r.mu.Lock()
	expect := r.expect
	if expect {
		r.calls = append(r.calls, formatted)
	}
	tb := r.tb
	r.mu.Unlock()
	if expect {
		panic(fatalExit{msg: formatted})
	}
	tb.Fatalf("logger exit called - %s", formatted)
}

// exitOnCode wraps exit for the pkg/errors.ExitFunc hook, which signals via
// exit code rather than a formatted message. The recorder stores a generic
// marker so assertions that use ExitCalls still observe that a fatal occurred.
func (r *ExitRecorder) exitOnCode(code int) {
	r.exit("process exit (code %d)", code)
}

// NewContext creates a new context for testing runners. It initializes the context with
// a real logger that writes it's output to a temporary file.
// It also creates a temporary testing directory for the test workspace, user configs, and caches.
// Test environment variables are set the config and cache directories override paths.
func NewContext(ctx stdCtx.Context, tb testing.TB) *Context {
	stdOut, stdIn := createTempIOFiles(tb)
	recorder := newExitRecorder(tb)
	tempLogger := tuikitIO.NewLogger(
		tuikitIO.WithOutput(stdOut),
		tuikitIO.WithTheme(logger.Theme("")),
		tuikitIO.WithMode(tuikitIO.Text),
		tuikitIO.WithExitFunc(recorder.exit),
	)
	logger.Init(logger.InitOptions{Logger: tempLogger, TestingTB: tb})
	flowerrors.ExitFunc = recorder.exitOnCode
	ctxx, configDir, cacheDir, wsDir := newTestContext(ctx, tb, stdIn, stdOut)
	// Route stderr to the same file as stdout so tests can assert on the
	// structured envelope emitted by HandleFatal in JSON/YAML mode.
	ctxx.SetStdErr(stdOut)
	return &Context{
		Context:   ctxx,
		configDir: configDir,
		cacheDir:  cacheDir,
		wsDir:     wsDir,
		exit:      recorder,
	}
}

type ContextWithMocks struct {
	Ctx             *context.Context
	Logger          *tuikitIOMocks.MockLogger
	ExecutableCache *cacheMocks.MockExecutableCache
	WorkspaceCache  *cacheMocks.MockWorkspaceCache
	RunnerMock      *mocks.MockRunner
}

// NewContextWithMocks creates a new context for testing runners. It initializes the context with
// a mock logger and mock caches. The mock logger is set to expect debug calls.
func NewContextWithMocks(ctx stdCtx.Context, tb testing.TB) *ContextWithMocks {
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		tb.Fatalf("unable to open devnull: %v", err)
	}
	configDir, cacheDir, wsDir := initTestDirectories(tb)
	setTestEnv(tb, configDir, cacheDir)
	testWsCfg, err := testWsConfig(wsDir)
	if err != nil {
		tb.Fatalf("unable to create workspace config: %v", err)
	}
	testUserCfg, err := testConfig(wsDir)
	if err != nil {
		tb.Fatalf("unable to create config: %v", err)
	}
	cancel := func() {
		<-ctx.Done()
	}
	mockLogger := tuikitIOMocks.NewMockLogger(gomock.NewController(tb))
	expectInternalMockLoggerCalls(mockLogger)
	logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: tb})
	wsCache := cacheMocks.NewMockWorkspaceCache(gomock.NewController(tb))
	execCache := cacheMocks.NewMockExecutableCache(gomock.NewController(tb))
	ctxx := &context.Context{
		Config:           testUserCfg,
		CurrentWorkspace: testWsCfg,
		WorkspacesCache:  wsCache,
		ExecutableCache:  execCache,
	}
	ctxx.SetContext(ctx, cancel)
	ctxx.SetIO(null, null)
	return &ContextWithMocks{
		Ctx:             ctxx,
		Logger:          mockLogger,
		ExecutableCache: execCache,
		WorkspaceCache:  wsCache,
		RunnerMock:      mocks.NewMockRunner(gomock.NewController(tb)),
	}
}

func ResetTestContext(ctx *Context, tb testing.TB) {
	c, cancel := stdCtx.WithCancel(stdCtx.Background())
	ctx.SetContext(c, cancel)
	stdIn, stdOut := createTempIOFiles(tb)
	ctx.SetIO(stdIn, stdOut)
	ctx.SetStdErr(stdOut)
	setTestEnv(tb, ctx.configDir, ctx.cacheDir)
	ctx.exit.setTB(tb)
	newLogger := tuikitIO.NewLogger(
		tuikitIO.WithOutput(stdOut),
		tuikitIO.WithTheme(logger.Theme("")),
		tuikitIO.WithMode(tuikitIO.Text),
		tuikitIO.WithExitFunc(ctx.exit.exit),
	)
	logger.Init(logger.InitOptions{Logger: newLogger, TestingTB: tb})
	flowerrors.ExitFunc = ctx.exit.exitOnCode
}

func createTempIOFiles(tb testing.TB) (stdIn *os.File, stdOut *os.File) {
	var err error
	stdOut, err = os.CreateTemp(tb.TempDir(), "flow-test-out")
	if err != nil {
		tb.Fatalf("unable to create temp file: %v", err)
	}
	stdIn, err = os.CreateTemp(tb.TempDir(), "flow-test-in")
	if err != nil {
		tb.Fatalf("unable to create temp file: %v", err)
	}
	// Registered after tb.TempDir() so it runs first (LIFO), ensuring handles
	// are closed before the temp directories are removed on Windows.
	tb.Cleanup(func() {
		_ = stdOut.Close()
		_ = stdIn.Close()
	})
	return
}

func newTestContext(
	ctx stdCtx.Context,
	tb testing.TB,
	stdIn, stdOut *os.File,
) (*context.Context, string, string, string) {
	configDir, cacheDir, wsDir := initTestDirectories(tb)
	setTestEnv(tb, configDir, cacheDir)

	testWsCfg, err := testWsConfig(wsDir)
	if err != nil {
		tb.Fatalf("unable to create workspace config: %v", err)
	}
	testCfg, err := testConfig(wsDir)
	if err != nil {
		tb.Fatalf("unable to create user config: %v", err)
	}

	wsCache, execCache, ds := testCaches(tb)

	cancel := func() {
		<-ctx.Done()
	}

	ctxx := &context.Context{
		Config:           testCfg,
		CurrentWorkspace: testWsCfg,
		WorkspacesCache:  wsCache,
		ExecutableCache:  execCache,
		DataStore:        ds,
	}
	ctxx.SetContext(ctx, cancel)
	ctxx.SetIO(stdIn, stdOut)
	return ctxx, configDir, cacheDir, wsDir
}

func initTestDirectories(tb testing.TB) (string, string, string) {
	replacer := strings.NewReplacer("-", "", "'", "-", "/", "-", " ", "_")
	suiteName := getSuiteName()
	tmpDir, err := os.MkdirTemp("", replacer.Replace(strings.ToLower(suiteName))) //nolint:usetesting
	if err != nil {
		tb.Fatalf("unable to create temp dir: %v", err)
	}

	tmpWsDir := filepath.Join(tmpDir, TestWorkspaceName)
	if err := os.Mkdir(filepath.Join(tmpDir, TestWorkspaceName), 0750); err != nil {
		tb.Fatalf("unable to create workspace directory: %v", err)
	}

	examplesFile := builder.ExamplesExecFlowFile()
	execDef, err := yaml.Marshal(examplesFile)
	if err != nil {
		tb.Fatalf("unable to marshal test data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpWsDir, "examples.flow"), execDef, 0600); err != nil {
		tb.Fatalf("unable to write test data: %v", err)
	}
	requestsFile := builder.ExamplesRequestExecFlowFile()
	reqDef, err := yaml.Marshal(requestsFile)
	if err != nil {
		tb.Fatalf("unable to marshal test data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpWsDir, "requests.flow"), reqDef, 0600); err != nil {
		tb.Fatalf("unable to write test data: %v", err)
	}
	rootFile := builder.RootExecFlowFile()
	rootDef, err := yaml.Marshal(rootFile)
	if err != nil {
		tb.Fatalf("unable to marshal test data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpWsDir, "root.flow"), rootDef, 0600); err != nil {
		tb.Fatalf("unable to write test data: %v", err)
	}

	tmpConfigDir := filepath.Join(tmpDir, userConfigSubdir)
	tmpCacheDir := filepath.Join(tmpDir, cacheSubdir)

	return tmpConfigDir, tmpCacheDir, tmpWsDir
}

func testConfig(wsDir string) (*config.Config, error) {
	if err := filesystem.InitConfig(); err != nil {
		return nil, err
	}
	userCfg, err := filesystem.LoadConfig()
	if err != nil {
		return nil, err
	}
	userCfg.DefaultLogMode = tuikitIO.Text
	userCfg.CurrentWorkspace = TestWorkspaceName
	userCfg.Workspaces = map[string]string{
		TestWorkspaceName: wsDir,
	}
	userCfg.Interactive = &config.Interactive{Enabled: false}
	if err = filesystem.WriteConfig(userCfg); err != nil {
		return nil, err
	}

	return userCfg, nil
}

func testWsConfig(wsDir string) (*workspace.Workspace, error) {
	if err := filesystem.InitWorkspaceConfig(TestWorkspaceName, wsDir); err != nil {
		return nil, err
	}
	wsCfg, err := filesystem.LoadWorkspaceConfig(TestWorkspaceName, wsDir)
	if err != nil {
		return nil, err
	}
	wsCfg.DisplayName = TestWorkspaceDisplayName
	if err = filesystem.WriteWorkspaceConfig(wsDir, wsCfg); err != nil {
		return nil, err
	}
	return wsCfg, nil
}

// testCaches must be called after the user and workspace configs have been created.
func testCaches(tb testing.TB) (cache.WorkspaceCache, cache.ExecutableCache, store.DataStore) {
	dbPath := filepath.Join(tb.TempDir(), "test_store.db")
	ds, err := store.NewDataStore(dbPath)
	if err != nil {
		tb.Fatalf("unable to open test data store: %v", err)
	}
	tb.Cleanup(func() { _ = ds.Close() })

	wsCache := cache.NewWorkspaceCache(ds)
	execCache := cache.NewExecutableCache(wsCache, ds)

	if err := wsCache.Update(); err != nil {
		tb.Fatalf("unable to update cache: %v", err)
	}
	if err := execCache.Update(); err != nil {
		tb.Fatalf("unable to update cache: %v", err)
	}
	return wsCache, execCache, ds
}

func setTestEnv(tb testing.TB, configDir, cacheDir string) {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0750); err != nil {
			tb.Fatalf("unable to create config directory: %v", err)
		}
	}
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0750); err != nil {
			tb.Fatalf("unable to create cache directory: %v", err)
		}
	}

	tb.Setenv(filesystem.FlowConfigDirEnvVar, configDir)
	tb.Setenv(filesystem.FlowCacheDirEnvVar, cacheDir)
	tb.Setenv(store.BucketEnv, "")
	tb.Setenv("NO_COLOR", "1")
}

func expectInternalMockLoggerCalls(logger *tuikitIOMocks.MockLogger) {
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().LogMode().AnyTimes()
	logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
}

// getSuiteName returns the name of the current Ginkgo test suite
func getSuiteName() string {
	if len(ginkgo.CurrentSpecReport().ContainerHierarchyTexts) > 0 {
		return ginkgo.CurrentSpecReport().ContainerHierarchyTexts[0]
	}
	return "flow-e2e-test" // generic fallback name
}
