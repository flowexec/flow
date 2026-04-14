package launch_test

import (
	stdCtx "context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine/mocks"
	"github.com/flowexec/flow/internal/runner/launch"
	testUtils "github.com/flowexec/flow/tests/utils"
	"github.com/flowexec/flow/types/executable"
)

func TestLaunch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Launch Suite")
}

var _ = Describe("Launch Runner", func() {
	var (
		launchRnr  runner.Runner
		ctx        *testUtils.ContextWithMocks
		mockEngine *mocks.MockEngine

		openCalls     []string
		openWithCalls [][2]string
		openErr       error
		openWithErr   error

		restoreOpen     func()
		restoreOpenWith func()
	)

	BeforeEach(func() {
		ctx = testUtils.NewContextWithMocks(stdCtx.Background(), GinkgoTB())
		launchRnr = launch.NewRunner()
		ctrl := gomock.NewController(GinkgoT())
		mockEngine = mocks.NewMockEngine(ctrl)

		openCalls = nil
		openWithCalls = nil
		openErr = nil
		openWithErr = nil
		restoreOpen = launch.SetOpenFnForTest(func(uri string) error {
			openCalls = append(openCalls, uri)
			return openErr
		})
		restoreOpenWith = launch.SetOpenWithFnForTest(func(app, uri string) error {
			openWithCalls = append(openWithCalls, [2]string{app, uri})
			return openWithErr
		})
	})

	AfterEach(func() {
		restoreOpen()
		restoreOpenWith()
	})

	Context("Name", func() {
		It("returns 'launch'", func() {
			Expect(launchRnr.Name()).To(Equal("launch"))
		})
	})

	Context("IsCompatible", func() {
		It("is false when executable is nil", func() {
			Expect(launchRnr.IsCompatible(nil)).To(BeFalse())
		})
		It("is false when Launch spec is nil", func() {
			Expect(launchRnr.IsCompatible(&executable.Executable{})).To(BeFalse())
		})
		It("is true when Launch spec is set", func() {
			e := &executable.Executable{Launch: &executable.LaunchExecutableType{}}
			Expect(launchRnr.IsCompatible(e)).To(BeTrue())
		})
	})

	Describe("Exec", func() {
		It("dispatches URIs with a scheme through Open", func() {
			e := &executable.Executable{Launch: &executable.LaunchExecutableType{URI: "https://example.com"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(launchRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(openCalls).To(ConsistOf("https://example.com"))
			Expect(openWithCalls).To(BeEmpty())
		})

		It("uses OpenWith when an App is specified", func() {
			e := &executable.Executable{Launch: &executable.LaunchExecutableType{
				URI: "https://example.com",
				App: "Firefox",
			}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(launchRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(openWithCalls).To(Equal([][2]string{{"Firefox", "https://example.com"}}))
			Expect(openCalls).To(BeEmpty())
		})

		It("expands environment variables in the URI", func() {
			GinkgoTB().Setenv("LAUNCH_TARGET", "https://expanded.example.com")
			e := &executable.Executable{Launch: &executable.LaunchExecutableType{URI: "$LAUNCH_TARGET"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			Expect(launchRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
			Expect(openCalls).To(ConsistOf("https://expanded.example.com"))
		})

		It("surfaces errors from the opener", func() {
			openErr = errors.New("boom")
			e := &executable.Executable{Launch: &executable.LaunchExecutableType{URI: "https://example.com"}}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			err := launchRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)
			Expect(err).To(MatchError("boom"))
		})
	})
})
