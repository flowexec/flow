//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/tests/utils"
	"github.com/flowexec/flow/v2/types/executable"
)

const statusFlowFile = `executables:
  - verb: status
    name: services
    exec:
      cmd: echo "all services healthy"
`

var _ = Describe("custom verbs e2e", func() {
	var (
		ctx *utils.Context
		run *utils.CommandRunner
	)

	BeforeEach(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		run = utils.NewE2ECommandRunner()
	})

	AfterEach(func() {
		executable.RegisterCustomVerbs()
		ctx.Finalize()
	})

	// setup runs a command whose output isn't asserted on. Each run finalizes the context's IO, so
	// it is reset for the next command.
	setup := func(args ...string) {
		Expect(run.Run(ctx.Context, args...)).To(Succeed())
		utils.ResetTestContext(ctx, GinkgoTB())
	}

	When("adding a verb (flow config add verb)", func() {
		It("persists the verb to the user config", func() {
			Expect(run.Run(ctx.Context, "config", "add", "verb", "status", "health")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Registered verbs: status, health"))

			cfg, err := filesystem.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.CustomVerbs).To(Equal([]string{"status", "health"}))
		})

		It("skips verbs that are already registered", func() {
			setup("config", "add", "verb", "status")
			Expect(run.Run(ctx.Context, "config", "add", "verb", "status")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Verbs already registered: status"))

			cfg, err := filesystem.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.CustomVerbs).To(Equal([]string{"status"}))
		})

		DescribeTable("rejects invalid verb names",
			func(name, msg string) {
				ctx.ExpectFailure()
				Expect(run.Run(ctx.Context, "config", "add", "verb", name)).To(HaveOccurred())
				Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring(msg)))
			},
			Entry("uppercase", "Status", "invalid verb"),
			Entry("built-in verb", "build", "already a built-in verb"),
			Entry("flow command name", "logs", "conflicts with the flow logs command"),
			Entry("flow command alias", "cfg", "conflicts with the flow cfg command"),
		)
	})

	When("the config file has a verb that collides with a flow command", func() {
		It("warns and keeps the flow command reachable", func() {
			ctx.Config.CustomVerbs = []string{"workspace"}
			Expect(filesystem.WriteConfig(ctx.Config)).To(Succeed())
			executable.RegisterCustomVerbs(ctx.Config.CustomVerbs...)

			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "workspace", "get", utils.TestWorkspaceName)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(`custom verb "workspace" conflicts with the flow workspace command`))
			Expect(out).To(ContainSubstring(utils.TestWorkspaceName))
		})
	})

	When("removing a verb (flow config remove verb)", func() {
		It("removes the verb from the user config", func() {
			setup("config", "add", "verb", "status", "health")
			Expect(run.Run(ctx.Context, "config", "remove", "verb", "status")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Removed verbs: status"))

			cfg, err := filesystem.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.CustomVerbs).To(Equal([]string{"health"}))
		})

		It("rejects a verb that is not registered", func() {
			ctx.ExpectFailure()
			Expect(run.Run(ctx.Context, "config", "remove", "verb", "status")).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("not a registered custom verb")))
		})
	})

	When("running an executable with a custom verb", func() {
		BeforeEach(func() {
			path := filepath.Join(ctx.WorkspaceDir(), "status.flow")
			Expect(os.WriteFile(path, []byte(statusFlowFile), 0600)).To(Succeed())
		})

		It("runs the executable once the verb is registered", func() {
			setup("config", "add", "verb", "status")
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "status", "services")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("all services healthy"))
		})

		It("does not recognize the verb before it is registered", func() {
			setup("sync")
			ctx.ExpectFailure()
			Expect(run.Run(ctx.Context, "status", "services")).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring(`unknown command "status"`)))
		})

		It("stops running the executable after the verb is removed", func() {
			setup("config", "add", "verb", "status")
			setup("config", "remove", "verb", "status")
			ctx.ExpectFailure()
			Expect(run.Run(ctx.Context, "status", "services")).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring(`unknown command "status"`)))
		})
	})
})
