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

	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine/mocks"
	"github.com/flowexec/flow/internal/runner/exec"
	testUtils "github.com/flowexec/flow/tests/utils"
	"github.com/flowexec/flow/types/executable"
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

		cmdCalls  []runCall
		fileCalls []runCall
		cmdErr    error
		fileErr   error

		restoreCmd  func()
		restoreFile func()
	)

	BeforeEach(func() {
		ctx = testUtils.NewContextWithMocks(stdCtx.Background(), GinkgoTB())
		execRnr = exec.NewRunner()
		ctrl := gomock.NewController(GinkgoT())
		mockEngine = mocks.NewMockEngine(ctrl)

		cmdCalls = nil
		fileCalls = nil
		cmdErr = nil
		fileErr = nil

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
	})

	AfterEach(func() {
		restoreCmd()
		restoreFile()
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
})
