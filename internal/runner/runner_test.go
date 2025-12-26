package runner_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	engMocks "github.com/flowexec/flow/internal/runner/engine/mocks"
	"github.com/flowexec/flow/internal/runner/mocks"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/types/executable"
)

func TestRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runner Suite")
}

var _ = Describe("Runner", func() {
	var (
		ctrl       *gomock.Controller
		mockRunner *mocks.MockRunner
		mockEngine *engMocks.MockEngine
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockRunner = mocks.NewMockRunner(ctrl)
		runner.RegisterRunner(mockRunner)
		engCtrl := gomock.NewController(GinkgoT())
		mockEngine = engMocks.NewMockEngine(engCtrl)
	})

	AfterEach(func() {
		ctrl.Finish()
		runner.Reset()
	})

	Describe("Exec", func() {
		It("should execute the runner correctly", func() {
			ctx := &context.Context{}
			executable := &executable.Executable{
				Name: "test-executable",
			}
			inputEnv := make(map[string]string)
			inputArgs := make([]string, 0)

			mockRunner.EXPECT().IsCompatible(executable).Return(true)
			mockRunner.EXPECT().Exec(ctx, executable, mockEngine, inputEnv, inputArgs).Return(nil)
			Expect(runner.Exec(ctx, executable, mockEngine, inputEnv, inputArgs)).To(Succeed())
		})

		It("should return error when no compatible runner is found", func() {
			ctx := &context.Context{}
			exec := &executable.Executable{
				Name: "test-exec",
			}
			inputEnv := make(map[string]string)
			inputArgs := make([]string, 0)

			mockRunner.EXPECT().IsCompatible(exec).Return(false)

			err := runner.Exec(ctx, exec, mockEngine, inputEnv, inputArgs)
			Expect(err.Error()).To(ContainSubstring("compatible runner not found"))
		})

		It("should return error when execution times out", func() {
			ctx := &context.Context{}
			timeout := 250 * time.Millisecond
			exec := &executable.Executable{
				Name:    "test-exec",
				Timeout: &timeout,
			}
			inputEnv := make(map[string]string)
			inputArgs := make([]string, 0)

			mockRunner.EXPECT().IsCompatible(exec).Return(true)
			mockRunner.EXPECT().Exec(ctx, exec, mockEngine, inputEnv, inputArgs).DoAndReturn(
				func(
					_ *context.Context, _ *executable.Executable, _ engine.Engine, _ map[string]string, _ []string,
				) error {
					time.Sleep(2 * time.Second)
					return nil
				})

			err := runner.Exec(ctx, exec, mockEngine, inputEnv, inputArgs)
			Expect(err.Error()).To(ContainSubstring("timeout"))
		})
	})
})
