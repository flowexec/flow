package exec_test

import (
	stdCtx "context"
	"errors"
	"os"
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

func TestExec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Exec Suite")
}

var _ = Describe("Exec Runner", func() {
	var (
		execRnr    runner.Runner
		ctx        *testUtils.ContextWithMocks
		mockEngine *mocks.MockEngine

		cmdCalls       []runCall
		fileCalls      []runCall
		containerSpecs []run.ContainerSpec
		cmdErr         error
		fileErr        error
		containerErr   error

		restoreCmd       func()
		restoreFile      func()
		restoreContainer func()
	)

	BeforeEach(func() {
		ctx = testUtils.NewContextWithMocks(stdCtx.Background(), GinkgoTB())
		execRnr = exec.NewRunner()
		ctrl := gomock.NewController(GinkgoT())
		mockEngine = mocks.NewMockEngine(ctrl)

		cmdCalls = nil
		fileCalls = nil
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
