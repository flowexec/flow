//go:build e2e

package tests_test

import (
	stdCtx "context"
	"fmt"
	stdIO "io"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/flowexec/tuikit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	execIO "github.com/flowexec/flow/v2/internal/io/executable"
	"github.com/flowexec/flow/v2/tests/utils"
	"github.com/flowexec/flow/v2/types/executable"
)

var _ = Describe("browse TUI", func() {
	var (
		ctx       *utils.Context
		container *tuikit.Container

		runChan chan string
		runFunc func(ref string) error
	)

	BeforeEach(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		runChan = make(chan string, 1)
		runFunc = func(ref string) error {
			runChan <- ref
			return nil
		}

		container = newTUIContainer(ctx)
		ctx.SetTUIContainer(container)
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	Specify("library snapshot", func() {
		tm := teatest.NewTestModel(GinkgoTB(), container, teatest.WithInitialTermSize(80, 25))
		container.Program().SetTeaProgram(tm.GetProgram())
		container.SetSendFunc(tm.Send)

		wsList, err := ctx.WorkspacesCache.GetWorkspaceConfigList()
		Expect(err).NotTo(HaveOccurred())
		execList, err := ctx.ExecutableCache.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())

		libraryView := execIO.NewLibraryView(
			ctx.Context, wsList, execList,
			execIO.Filter{},
			runFunc,
		)
		Expect(container.SetView(libraryView)).To(Succeed())

		container.Send(tea.Quit(), 250*time.Millisecond)
		tm.WaitFinished(GinkgoTB(), teatest.WithFinalTimeout(500*time.Millisecond))
		out, err := stdIO.ReadAll(tm.FinalOutput(GinkgoTB()))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).NotTo(BeEmpty())
	})

	Specify("list snapshot", func() {
		tm := teatest.NewTestModel(GinkgoTB(), container, teatest.WithInitialTermSize(80, 25))
		container.Program().SetTeaProgram(tm.GetProgram())
		container.SetSendFunc(tm.Send)
		fmt.Println("Running executable list snapshot test...")

		execList, err := ctx.ExecutableCache.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())
		listView := execIO.NewExecutableListView(ctx.Context, execList, runFunc)
		Expect(container.SetView(listView)).To(Succeed())

		container.Send(tea.Quit(), 250*time.Millisecond)
		tm.WaitFinished(GinkgoTB(), teatest.WithFinalTimeout(500*time.Millisecond))
		out, err := stdIO.ReadAll(tm.FinalOutput(GinkgoTB()))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).NotTo(BeEmpty())
	})

	Specify("exec snapshot", func() {
		path := filepath.Join(ctx.WorkspaceDir(), "snapshot.flow")
		exec := &executable.Executable{
			Verb: "show",
			Name: "snapshot",
			Exec: &executable.ExecExecutableType{Cmd: "echo 'Hello, world! This is a snapshot test.'"},
		}
		exec.SetContext(utils.TestWorkspaceName, ctx.WorkspaceDir(), "", path)

		tm := teatest.NewTestModel(GinkgoTB(), container, teatest.WithInitialTermSize(80, 25))
		container.Program().SetTeaProgram(tm.GetProgram())
		container.SetSendFunc(tm.Send)

		execView := execIO.NewExecutableView(ctx.Context, exec, runFunc)
		Expect(container.SetView(execView)).To(Succeed())

		container.Send(tea.Quit(), 250*time.Millisecond)
		tm.WaitFinished(GinkgoTB(), teatest.WithFinalTimeout(500*time.Millisecond))
		out, err := stdIO.ReadAll(tm.FinalOutput(GinkgoTB()))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).NotTo(BeEmpty())
	})
})

var _ = Describe("browse e2e", Ordered, func() {
	var (
		ctx *utils.Context
		run *utils.CommandRunner
	)

	BeforeAll(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		run = utils.NewE2ECommandRunner()
	})

	BeforeEach(func() {
		utils.ResetTestContext(ctx, GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	DescribeTable("browse list with various filters produces YAML output",
		func(args []string) {
			stdOut := ctx.StdOut()
			cmdArgs := append([]string{"browse", "--list"}, args...)
			Expect(run.Run(ctx.Context, cmdArgs...)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("executables:"))
		},
		Entry("no filter", []string{}),
		Entry("workspace filter", []string{"--workspace", "."}),
		Entry("namespace filter", []string{"--namespace", "."}),
		Entry("all namespaces", []string{"--all"}),
		Entry("verb filter", []string{"--verb", "exec"}),
		Entry("tag filter", []string{"--tag", "test"}),
		Entry("substring filter", []string{"--filter", "print"}),
		Entry("multiple filters", []string{"--verb", "exec", "--workspace", ".", "--namespace", "."}),
	)

	It("should show executable details by verb and name", func() {
		stdOut := ctx.StdOut()
		Expect(run.Run(ctx.Context, "browse", "exec", "examples:simple-print")).To(Succeed())
		out, err := readFileContent(stdOut)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("name: simple-print"))
	})

	When("browsing with an invalid visibility filter", func() {
		It("fatals with an invalid visibility message", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "browse", "--list", "--visibility", "bogus")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("invalid visibility")))
		})
	})

	When("browsing an executable that does not exist", func() {
		It("fatals with a not-found message", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "browse", "exec", "examples:doesnotexist")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).NotTo(BeEmpty())
		})
	})
})
