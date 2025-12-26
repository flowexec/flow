package serial_test

import (
	stdCtx "context"
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/runner/engine/mocks"
	"github.com/flowexec/flow/internal/runner/serial"
	"github.com/flowexec/flow/pkg/context"
	testUtils "github.com/flowexec/flow/tests/utils"
	"github.com/flowexec/flow/tests/utils/builder"
	"github.com/flowexec/flow/types/executable"
)

func TestSerialRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Serial Runner Suite")
}

var _ = Describe("SerialRunner", func() {
	var (
		ctx        *testUtils.ContextWithMocks
		serialRnr  runner.Runner
		mockEngine *mocks.MockEngine
	)

	BeforeEach(func() {
		ctx = testUtils.NewContextWithMocks(stdCtx.Background(), GinkgoTB())
		runner.RegisterRunner(ctx.RunnerMock)
		serialRnr = serial.NewRunner()
		engCtl := gomock.NewController(GinkgoT())
		mockEngine = mocks.NewMockEngine(engCtl)
	})

	AfterEach(func() {
		runner.Reset()
	})

	Context("Name", func() {
		It("should return the correct runner name", func() {
			Expect(serialRnr.Name()).To(Equal("serial"))
		})
	})

	Context("IsCompatible", func() {
		It("should return false when executable is nil", func() {
			Expect(serialRnr.IsCompatible(nil)).To(BeFalse())
		})

		It("should return false when executable type is nil", func() {
			executable := &executable.Executable{}
			Expect(serialRnr.IsCompatible(executable)).To(BeFalse())
		})

		It("should return true when executable type is serial", func() {
			executable := &executable.Executable{
				Serial: &executable.SerialExecutableType{},
			}
			Expect(serialRnr.IsCompatible(executable)).To(BeTrue())
		})
	})

	When("Exec", func() {
		var (
			rootExec *executable.Executable
			subExecs executable.ExecutableList
		)

		BeforeEach(func() {
			ns := "examples"
			rootExec = builder.SerialExecByRefConfig(
				builder.WithNamespaceName(ns),
				builder.WithWorkspaceName(ctx.Ctx.CurrentWorkspace.AssignedName()),
				builder.WithWorkspacePath(ctx.Ctx.CurrentWorkspace.Location()),
			)
			execFlowfile := builder.ExamplesExecFlowFile(
				builder.WithNamespaceName(ns),
				builder.WithWorkspaceName(ctx.Ctx.CurrentWorkspace.AssignedName()),
				builder.WithWorkspacePath(ctx.Ctx.CurrentWorkspace.Location()),
			)
			subExecs = testUtils.FindSubExecs(rootExec, executable.FlowFileList{execFlowfile})

			runner.RegisterRunner(serialRnr)
			runner.RegisterRunner(ctx.RunnerMock)
			ctx.RunnerMock.EXPECT().IsCompatible(rootExec).Return(false).AnyTimes()
		})

		It("complete successfully when there are no engine errors", func() {
			promptedEnv := make(map[string]string)
			mockCache := ctx.ExecutableCache

			for i, e := range subExecs {
				switch i {
				case 0:
					mockCache.EXPECT().GetExecutableByRef(e.Ref()).Return(e, nil).Times(1)
				case 1:
					mockCache.EXPECT().GetExecutableByRef(e.Ref()).Return(e, nil).Times(1)
				}
			}
			results := engine.ResultSummary{Results: []engine.Result{{}}}
			mockEngine.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(results).Times(1)
			Expect(serialRnr.Exec(ctx.Ctx, rootExec, mockEngine, promptedEnv, nil)).To(Succeed())
		})

		It("should fail when there is an engine failure", func() {
			mockCache := ctx.ExecutableCache
			for i, e := range subExecs {
				switch i {
				case 0:
					mockCache.EXPECT().GetExecutableByRef(e.Ref()).Return(e, nil).Times(1)
				case 1:
					mockCache.EXPECT().GetExecutableByRef(e.Ref()).Return(e, nil).Times(1)
				}
			}
			results := engine.ResultSummary{Results: []engine.Result{{Error: errors.New("error")}}}
			mockEngine.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(results).Times(1)
			Expect(serialRnr.Exec(ctx.Ctx, rootExec, mockEngine, make(map[string]string), nil)).ToNot(Succeed())
		})

		It("should skip execution when condition is false", func() {
			serialSpec := rootExec.Serial
			serialSpec.Execs[0].If = "false"
			serialSpec.Execs[1].If = "true"
			mockCache := ctx.ExecutableCache
			// Only the first two execs use Ref (third uses Cmd, so no cache call)
			for i, e := range subExecs {
				if i < 2 {
					mockCache.EXPECT().GetExecutableByRef(e.Ref()).Return(e, nil).Times(1)
				}
			}

			results := engine.ResultSummary{Results: []engine.Result{{}}}
			mockEngine.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ stdCtx.Context, execs []engine.Exec, _ ...engine.OptionFunc) engine.ResultSummary {
					// Verify all execs from rootExec are in the list (conditions included)
					Expect(execs).To(HaveLen(len(serialSpec.Execs)))
					// Verify first two have conditions, third doesn't
					Expect(execs[0].Condition).ToNot(BeNil())
					Expect(execs[1].Condition).ToNot(BeNil())
					Expect(execs[2].Condition).To(BeNil())
					return results
				}).Times(1)
			Expect(serialRnr.Exec(ctx.Ctx, rootExec, mockEngine, make(map[string]string), nil)).To(Succeed())
		})

		It("should pass environment args from parent to child executables", func() {
			pos1 := 1
			parentExec := &executable.Executable{
				Serial: &executable.SerialExecutableType{
					Args: executable.ArgumentList{{EnvKey: "TEST_VAR", Pos: &pos1}},
					Execs: []executable.SerialRefConfig{{
						Ref:  "test:child",
						Args: []string{"var=$TEST_VAR"},
					},
					},
				},
			}
			parentExec.SetContext("test", "/test", "test", "/test/parent.flow")

			childExec := &executable.Executable{
				Exec: &executable.ExecExecutableType{
					Cmd:  "echo $TEST_VAR",
					Args: executable.ArgumentList{{EnvKey: "TEST_VAR", Flag: "var"}},
				},
			}
			childExec.SetContext("test", "/test", "test", "/test/child.flow")
			mockCache := ctx.ExecutableCache
			mockCache.EXPECT().GetExecutableByRef(gomock.Any()).Return(childExec, nil).Times(1)

			ctx.RunnerMock.EXPECT().IsCompatible(gomock.Any()).Return(true).Times(1)
			ctx.RunnerMock.EXPECT().
				Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), []string{"var=test_value"}).
				DoAndReturn(func(
					_ *context.Context,
					exec *executable.Executable,
					_ engine.Engine,
					inputEnv map[string]string,
					inputArgs []string,
				) error {
					Expect(inputEnv).To(HaveKeyWithValue("TEST_VAR", "test_value"))
					Expect(inputArgs).To(ContainElement("var=test_value"))
					return nil
				}).Times(1)

			results := engine.ResultSummary{Results: []engine.Result{{}}}
			mockEngine.EXPECT().
				Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(
					_ stdCtx.Context, execs []engine.Exec, _ ...engine.OptionFunc) engine.ResultSummary {
					for _, exec := range execs {
						Expect(exec.Function()).To(Succeed())
					}
					return results
				})

			Expect(serialRnr.Exec(ctx.Ctx, parentExec, mockEngine, make(map[string]string), []string{"test_value"})).
				To(Succeed())
		})

		It("refreshes cache between steps so later conditions see updates", func() {
			ns := "examples"
			parentExec := &executable.Executable{
				Serial: &executable.SerialExecutableType{
					Execs: []executable.SerialRefConfig{
						{Ref: executable.Ref(fmt.Sprintf("exec %s:first", ns))},
						{Ref: executable.Ref(fmt.Sprintf("exec %s:second", ns)), If: `store["flow-test_serial_updated"] == 'true'`},
					},
				},
			}
			parentExec.SetContext(
				ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(),
				ns, "/test/parent.flow",
			)

			firstChild := &executable.Executable{
				Verb: "exec", Name: "first",
				Exec: &executable.ExecExecutableType{Cmd: "echo first"},
			}
			firstChild.SetContext(
				ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(),
				ns, "/test/first.flow")

			secondChild := &executable.Executable{
				Verb: "exec", Name: "second",
				Exec: &executable.ExecExecutableType{Cmd: "echo second"},
			}
			secondChild.SetContext(
				ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(),
				ns, "/test/second.flow")

			mockCache := ctx.ExecutableCache
			mockCache.EXPECT().GetExecutableByRef(firstChild.Ref()).Return(firstChild, nil).Times(1)
			mockCache.EXPECT().GetExecutableByRef(secondChild.Ref()).Return(secondChild, nil).Times(1)

			ctx.RunnerMock.EXPECT().IsCompatible(gomock.Any()).Return(true).AnyTimes()
			ctx.RunnerMock.EXPECT().Exec(gomock.Any(), firstChild, gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(
					func(_ *context.Context, _ *executable.Executable, _ engine.Engine, _ map[string]string, _ []string) error {
						return nil
					}).Times(1)
			ctx.RunnerMock.
				EXPECT().
				Exec(gomock.Any(), secondChild, gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			results := engine.ResultSummary{Results: []engine.Result{{}, {}}}
			mockEngine.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ stdCtx.Context, execs []engine.Exec, _ ...engine.OptionFunc) engine.ResultSummary {
					for _, ex := range execs {
						Expect(ex.Function()).To(Succeed())
					}
					return results
				}).Times(1)

			Expect(serialRnr.Exec(ctx.Ctx, parentExec, mockEngine, make(map[string]string), nil)).To(Succeed())
		})
	})
})
