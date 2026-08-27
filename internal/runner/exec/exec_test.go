package exec_test

import (
	stdCtx "context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	tuikitIO "github.com/flowexec/tuikit/io"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/internal/runner"
	"github.com/flowexec/flow/v2/internal/runner/engine/mocks"
	"github.com/flowexec/flow/v2/internal/runner/exec"
	"github.com/flowexec/flow/v2/internal/services/run"
	testUtils "github.com/flowexec/flow/v2/tests/utils"
	"github.com/flowexec/flow/v2/types/executable"
)

type runCall struct {
	target  string // cmd string for runCmd, filename for runFile
	dir     string
	envList []string
	mode    tuikitIO.LogMode
}

// absHostVolumePath is an absolute host path valid on the current platform.
// expandVolumeHost gates on filepath.IsAbs, which rejects a Unix-style path on
// Windows - there an absolute path needs a drive letter. The container side of a
// volume stays Unix, since container paths are always Linux paths.
func absHostVolumePath() string {
	if runtime.GOOS == "windows" {
		return `C:\opt\data`
	}
	return "/opt/data"
}

func TestExec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Exec Suite")
}

var _ = Describe("Exec Runner", func() {
	var (
		execRnr    runner.Runner
		ctx        *testUtils.ContextWithMocks
		mockEngine *mocks.MockEngine

		cmdCalls        []runCall
		fileCalls       []runCall
		pythonCalls     []runCall
		pythonFileCalls []runCall
		containerSpecs  []run.ContainerSpec
		cmdErr          error
		fileErr         error
		containerErr    error

		restoreCmd        func()
		restoreFile       func()
		restorePython     func()
		restorePythonFile func()
		restoreContainer  func()
	)

	BeforeEach(func() {
		ctx = testUtils.NewContextWithMocks(stdCtx.Background(), GinkgoTB())
		execRnr = exec.NewRunner()
		ctrl := gomock.NewController(GinkgoT())
		mockEngine = mocks.NewMockEngine(ctrl)

		cmdCalls = nil
		fileCalls = nil
		pythonCalls = nil
		pythonFileCalls = nil
		containerSpecs = nil
		cmdErr = nil
		fileErr = nil
		containerErr = nil

		restoreCmd = exec.SetRunCmdFnForTest(func(
			s, dir string, envList []string, logMode tuikitIO.LogMode,
			_ tuikitIO.Logger, _ *os.File, _ map[string]any, _ *tuikitIO.TaskContext,
		) error {
			cmdCalls = append(cmdCalls, runCall{target: s, dir: dir, envList: envList, mode: logMode})
			return cmdErr
		})
		restoreFile = exec.SetRunFileFnForTest(func(
			s, dir string, envList []string, logMode tuikitIO.LogMode,
			_ tuikitIO.Logger, _ *os.File, _ map[string]any, _ *tuikitIO.TaskContext,
		) error {
			fileCalls = append(fileCalls, runCall{target: s, dir: dir, envList: envList, mode: logMode})
			return fileErr
		})
		restorePython = exec.SetRunPythonFnForTest(func(
			s, dir string, envList []string, logMode tuikitIO.LogMode,
			_ tuikitIO.Logger, _ *os.File, _ map[string]any, _ *tuikitIO.TaskContext,
		) error {
			pythonCalls = append(pythonCalls, runCall{target: s, dir: dir, envList: envList, mode: logMode})
			return cmdErr
		})
		restorePythonFile = exec.SetRunPythonFileFnForTest(func(
			s, dir string, envList []string, logMode tuikitIO.LogMode,
			_ tuikitIO.Logger, _ *os.File, _ map[string]any, _ *tuikitIO.TaskContext,
		) error {
			pythonFileCalls = append(pythonFileCalls, runCall{target: s, dir: dir, envList: envList, mode: logMode})
			return fileErr
		})
		restoreContainer = exec.SetRunContainerFnForTest(func(
			_ stdCtx.Context, spec run.ContainerSpec, _ tuikitIO.LogMode,
			_ tuikitIO.Logger, _ *os.File, _ map[string]any, _ *tuikitIO.TaskContext,
		) error {
			containerSpecs = append(containerSpecs, spec)
			return containerErr
		})
	})

	AfterEach(func() {
		restoreCmd()
		restoreFile()
		restorePython()
		restorePythonFile()
		restoreContainer()
	})

	Context("Name", func() {
		It("returns 'exec'", func() {
			Expect(execRnr.Name()).To(Equal("exec"))
		})
	})

	Context("IsCompatible", func() {
		It("is false when executable is nil", func() {
			Expect(execRnr.IsCompatible(nil)).To(BeFalse())
		})
		It("is false when Exec spec is nil", func() {
			Expect(execRnr.IsCompatible(&executable.Executable{})).To(BeFalse())
		})
		It("is true when Exec spec is set", func() {
			Expect(execRnr.IsCompatible(&executable.Executable{Exec: &executable.ExecExecutableType{}})).To(BeTrue())
		})
	})

	Describe("Exec", func() {
		It("errors when neither cmd nor file is set", func() {
			e := &executable.Executable{Exec: &executable.ExecExecutableType{}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			err := execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)
			Expect(err).To(MatchError(ContainSubstring("either cmd or file must be specified")))
			Expect(cmdCalls).To(BeEmpty())
			Expect(fileCalls).To(BeEmpty())
		})

		It("errors when both cmd and file are set", func() {
			e := &executable.Executable{Exec: &executable.ExecExecutableType{Cmd: "echo hi", File: "script.sh"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			err := execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)
			Expect(err).To(MatchError(ContainSubstring("cannot set both cmd and file")))
			Expect(cmdCalls).To(BeEmpty())
			Expect(fileCalls).To(BeEmpty())
		})

		It("routes a cmd through runCmd", func() {
			e := &executable.Executable{Exec: &executable.ExecExecutableType{Cmd: "echo hello"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(cmdCalls).To(HaveLen(1))
			Expect(cmdCalls[0].target).To(Equal("echo hello"))
			Expect(fileCalls).To(BeEmpty())
		})

		It("routes a python cmd through runPython instead of the shell", func() {
			interpreter := executable.InterpreterPython
			e := &executable.Executable{Exec: &executable.ExecExecutableType{
				Cmd: "print('hello')", Interpreter: &interpreter,
			}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(pythonCalls).To(HaveLen(1))
			Expect(pythonCalls[0].target).To(Equal("print('hello')"))
			Expect(cmdCalls).To(BeEmpty())
		})

		It("routes an explicit sh cmd through runCmd", func() {
			interpreter := executable.InterpreterSh
			e := &executable.Executable{Exec: &executable.ExecExecutableType{
				Cmd: "echo hello", Interpreter: &interpreter,
			}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(cmdCalls).To(HaveLen(1))
			Expect(pythonCalls).To(BeEmpty())
		})

		It("routes a .py file through runPythonFile without an explicit interpreter", func() {
			e := &executable.Executable{Exec: &executable.ExecExecutableType{File: "script.py"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(pythonFileCalls).To(HaveLen(1))
			// The python file seam receives a full path, unlike runFile.
			Expect(pythonFileCalls[0].target).To(HaveSuffix("script.py"))
			Expect(fileCalls).To(BeEmpty())
		})

		It("lets an explicit interpreter override the file extension", func() {
			interpreter := executable.InterpreterPython
			e := &executable.Executable{Exec: &executable.ExecExecutableType{
				File: "script.sh", Interpreter: &interpreter,
			}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(pythonFileCalls).To(HaveLen(1))
			Expect(fileCalls).To(BeEmpty())
		})

		It("routes a file through runFile", func() {
			e := &executable.Executable{Exec: &executable.ExecExecutableType{File: "script.sh"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(fileCalls).To(HaveLen(1))
			Expect(fileCalls[0].target).To(Equal("script.sh"))
			Expect(cmdCalls).To(BeEmpty())
		})

		It("surfaces errors returned from runCmd", func() {
			cmdErr = errors.New("run failed")
			e := &executable.Executable{Exec: &executable.ExecExecutableType{Cmd: "bad"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			err := execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)
			Expect(err).To(MatchError("run failed"))
		})

		It("propagates provided env entries into the envList passed to runCmd", func() {
			e := &executable.Executable{Exec: &executable.ExecExecutableType{Cmd: "echo hi"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			inputEnv := map[string]string{"FLOW_TEST_VAR": "hello"}
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, inputEnv, nil)).To(Succeed())
			Expect(cmdCalls).To(HaveLen(1))
			Expect(cmdCalls[0].envList).To(ContainElement("FLOW_TEST_VAR=hello"))
		})
	})

	Describe("Exec with container", func() {
		var wsPath string

		newContainerExec := func(c *executable.ExecContainer, dir executable.Directory) *executable.Executable {
			e := &executable.Executable{Exec: &executable.ExecExecutableType{Cmd: "echo hi", Dir: dir, Container: c}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), wsPath, "", "")
			e.SetDefaults() // container tests build literals, so defaults must be applied explicitly
			return e
		}

		BeforeEach(func() {
			wsPath = ctx.Ctx.CurrentWorkspace.Location()
			restore := exec.SetLookPathForContainerTest()
			DeferCleanup(restore)
		})

		newPythonContainerExec := func(c *executable.ExecContainer, dir executable.Directory) *executable.Executable {
			interpreter := executable.InterpreterPython
			e := &executable.Executable{Exec: &executable.ExecExecutableType{
				Cmd: "print('hi')", Dir: dir, Container: c, Interpreter: &interpreter,
			}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), wsPath, "", "")
			e.SetDefaults()
			return e
		}

		It("mounts python code as a script instead of passing it as -c", func() {
			// spec.Cmd would become `python3 -c <code>`, putting the code in the
			// process table and losing traceback line numbers.
			e := newPythonContainerExec(&executable.ExecContainer{Image: "python:3.13"}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs).To(HaveLen(1))
			Expect(containerSpecs[0].Cmd).To(BeEmpty())
			Expect(containerSpecs[0].Script).To(HaveSuffix(".py"))

			var scriptMount *run.Mount
			for i := range containerSpecs[0].Mounts {
				if containerSpecs[0].Mounts[i].ContainerPath == containerSpecs[0].Script {
					scriptMount = &containerSpecs[0].Mounts[i]
				}
			}
			Expect(scriptMount).ToNot(BeNil(), "the generated script should be bind-mounted")
			Expect(scriptMount.ReadOnly).To(BeTrue())
		})

		It("defaults the entrypoint to python3 for a python interpreter", func() {
			e := newPythonContainerExec(&executable.ExecContainer{Image: "python:3.13"}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].Entrypoint).To(Equal("python3"))
			Expect(containerSpecs[0].OverrideEntry).To(BeTrue())
		})

		It("lets an explicit entrypoint override the python default", func() {
			e := newPythonContainerExec(&executable.ExecContainer{
				Image: "python:3.13", Entrypoint: strPtrFor("python3.13"),
			}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].Entrypoint).To(Equal("python3.13"))
		})

		It("uses the image ENTRYPOINT when entrypoint is explicitly empty", func() {
			e := newPythonContainerExec(&executable.ExecContainer{
				Image: "python:3.13", Entrypoint: strPtrFor(""),
			}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].OverrideEntry).To(BeFalse())
		})

		It("keeps a shell command on the sh entrypoint", func() {
			e := newContainerExec(&executable.ExecContainer{Image: "alpine:3"}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].Entrypoint).To(Equal("sh"))
		})

		It("drops host python paths from the container environment", func() {
			// A host venv path names nothing inside the container, and could even
			// resolve to an unrelated mounted directory.
			e := newPythonContainerExec(&executable.ExecContainer{Image: "python:3.13"}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{
				"VIRTUAL_ENV":     "/host/.venv",
				"PYTHONPATH":      "/host/site-packages",
				"PYTHONHOME":      "/host/python",
				"FLOW_PYTHON_BIN": "/host/.venv/bin/python",
				"KEEP_ME":         "yes",
			}, nil)).To(Succeed())

			env := containerSpecs[0].Env
			Expect(env).ToNot(HaveKey("VIRTUAL_ENV"))
			Expect(env).ToNot(HaveKey("PYTHONPATH"))
			Expect(env).ToNot(HaveKey("PYTHONHOME"))
			Expect(env).ToNot(HaveKey("FLOW_PYTHON_BIN"))
			Expect(env).To(HaveKeyWithValue("KEEP_ME", "yes"))
			Expect(env).To(HaveKeyWithValue("PYTHONUNBUFFERED", "1"))
		})

		It("routes to the container backend instead of runCmd", func() {
			e := newContainerExec(&executable.ExecContainer{Image: "alpine:3"}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(cmdCalls).To(BeEmpty())
			Expect(containerSpecs).To(HaveLen(1))
			Expect(containerSpecs[0].Image).To(Equal("alpine:3"))
			Expect(containerSpecs[0].Cmd).To(Equal("echo hi"))
		})

		It("translates a workspace-relative dir to the mount path", func() {
			e := newContainerExec(&executable.ExecContainer{Image: "alpine:3"}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].Workdir).To(Equal("/workspace"))
		})

		It("honors a custom mountWorkspace for the mount and workdir", func() {
			c := &executable.ExecContainer{Image: "alpine:3", MountWorkspace: "/src"}
			e := newContainerExec(c, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].Workdir).To(Equal("/src"))
			Expect(containerSpecs[0].Mounts[0].ContainerPath).To(Equal("/src"))
		})

		It("mounts an out-of-workspace tmp dir at the fallback workdir", func() {
			e := newContainerExec(&executable.ExecContainer{Image: "alpine:3"}, executable.Directory("f:tmp"))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].Workdir).To(Equal("/flow/workdir"))
			Expect(containerSpecs[0].Mounts).To(HaveLen(2))
			Expect(containerSpecs[0].Mounts[1].ContainerPath).To(Equal("/flow/workdir"))
		})

		It("drops host-only FLOW_* vars and rewrites the workspace path", func() {
			e := newContainerExec(&executable.ExecContainer{Image: "alpine:3"}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			env := containerSpecs[0].Env
			Expect(env).NotTo(HaveKey("FLOW_CONFIG_PATH"))
			Expect(env).NotTo(HaveKey("FLOW_CACHE_PATH"))
			Expect(env).To(HaveKeyWithValue("FLOW_WORKSPACE_PATH", "/workspace"))
			Expect(env).To(HaveKeyWithValue("FLOW_IN_CONTAINER", "true"))
		})

		It("passes no env when inheritEnv is false", func() {
			inherit := false
			c := &executable.ExecContainer{Image: "alpine:3", InheritEnv: &inherit}
			e := newContainerExec(c, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].Env).To(BeEmpty())
		})

		It("uses the image entrypoint when entrypoint is explicitly empty", func() {
			empty := ""
			e := newContainerExec(&executable.ExecContainer{Image: "node:18", Entrypoint: &empty}, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].OverrideEntry).To(BeFalse())
		})

		It("uses an explicit user verbatim", func() {
			user := "1000:1000"
			c := &executable.ExecContainer{Image: "alpine:3", User: &user}
			e := newContainerExec(c, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs[0].User).To(Equal("1000:1000"))
		})

		It("expands and mounts workspace-relative and absolute volumes", func() {
			c := &executable.ExecContainer{
				Image: "alpine:3",
				Volumes: []executable.ExecContainerVolume{
					"//cache:/cache",
					executable.ExecContainerVolume(absHostVolumePath() + ":/data:ro"),
				},
			}
			e := newContainerExec(c, executable.Directory(wsPath))
			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())

			mounts := containerSpecs[0].Mounts
			// mounts[0] is the workspace; the two user volumes follow in order.
			Expect(mounts[len(mounts)-2].HostPath).To(Equal(filepath.Join(wsPath, "cache")))
			Expect(mounts[len(mounts)-2].ContainerPath).To(Equal("/cache"))
			Expect(mounts[len(mounts)-1].HostPath).To(Equal(absHostVolumePath()))
			Expect(mounts[len(mounts)-1].ContainerPath).To(Equal("/data"))
			Expect(mounts[len(mounts)-1].Options).To(Equal("ro"))
		})

		It("returns an error for a malformed volume", func() {
			c := &executable.ExecContainer{
				Image:   "alpine:3",
				Volumes: []executable.ExecContainerVolume{"/host:relative-container"},
			}
			e := newContainerExec(c, executable.Directory(wsPath))
			err := execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)
			Expect(err).To(MatchError(ContainSubstring("container path must be absolute")))
			Expect(containerSpecs).To(BeEmpty())
		})

		It("maps a workspace-relative file to its container script path", func() {
			Expect(os.WriteFile(filepath.Join(wsPath, "build.sh"), []byte("echo hi\n"), 0o600)).To(Succeed())
			e := &executable.Executable{Exec: &executable.ExecExecutableType{
				File:      "build.sh",
				Dir:       executable.Directory(wsPath),
				Container: &executable.ExecContainer{Image: "alpine:3"},
			}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), wsPath, "", "")
			e.SetDefaults()

			Expect(execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(containerSpecs).To(HaveLen(1))
			Expect(containerSpecs[0].Cmd).To(BeEmpty())
			Expect(containerSpecs[0].Script).To(Equal("/workspace/build.sh"))
		})

		DescribeTable("expands volume host paths",
			func(host string, wantErr bool, wantSuffix string) {
				got, err := exec.ExpandVolumeHostForTest(host, "/ws/root")
				if wantErr {
					Expect(err).To(HaveOccurred())
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(HaveSuffix(wantSuffix))
			},
			Entry("workspace-relative", "//sub/dir", false, filepath.Join("/ws/root", "sub", "dir")),
			Entry("absolute", absHostVolumePath(), false, absHostVolumePath()),
			Entry("home-relative", "~/thing", false, "thing"),
			Entry("cwd-relative", "./local", false, "local"),
			Entry("bare relative is rejected", "relative/path", true, ""),
		)

		It("rejects a PowerShell file in a container", func() {
			e := &executable.Executable{Exec: &executable.ExecExecutableType{
				File:      "script.ps1",
				Dir:       executable.Directory(wsPath),
				Container: &executable.ExecContainer{Image: "alpine:3"},
			}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), wsPath, "", "")
			e.SetDefaults()

			err := execRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)
			Expect(err).To(MatchError(ContainSubstring("does not support")))
			Expect(containerSpecs).To(BeEmpty())
		})
	})
})

// strPtrFor is a local helper for building optional container fields in tests.
func strPtrFor(s string) *string { return &s }
