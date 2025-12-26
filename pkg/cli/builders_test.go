package cli_test

import (
	stdCtx "context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/pkg/cli"
	"github.com/flowexec/flow/pkg/context"
)

func TestBuilders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Builders Suite")
}

var _ = Describe("BuildRootCommand", func() {
	var ctx *context.Context

	BeforeEach(func() {
		bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
		ctx = context.NewContext(bkgCtx, cancelFunc, io.Stdin, io.Stdout)
	})

	AfterEach(func() {
		if ctx != nil {
			ctx.Finalize()
		}
	})

	It("should create a root command with default configuration", func() {
		rootCmd := cli.BuildRootCommand(ctx)

		Expect(rootCmd).NotTo(BeNil())
		Expect(rootCmd.Use).To(Equal("flow"))
		Expect(rootCmd.Short).NotTo(BeEmpty())
	})

	It("should apply custom use", func() {
		rootCmd := cli.BuildRootCommand(ctx, cli.WithUse("mycli"))

		Expect(rootCmd.Use).To(Equal("mycli"))
	})

	It("should apply custom short description", func() {
		rootCmd := cli.BuildRootCommand(ctx, cli.WithShort("Custom short"))

		Expect(rootCmd.Short).To(Equal("Custom short"))
	})

	It("should apply custom long description", func() {
		rootCmd := cli.BuildRootCommand(ctx, cli.WithLong("Custom long"))

		Expect(rootCmd.Long).To(Equal("Custom long"))
	})

	It("should apply custom version", func() {
		rootCmd := cli.BuildRootCommand(ctx, cli.WithVersion("1.0.0-custom"))

		Expect(rootCmd.Version).To(Equal("1.0.0-custom"))
	})

	It("should apply multiple options", func() {
		rootCmd := cli.BuildRootCommand(ctx,
			cli.WithUse("mycli"),
			cli.WithShort("Custom short"),
			cli.WithVersion("2.0.0"),
		)

		Expect(rootCmd.Use).To(Equal("mycli"))
		Expect(rootCmd.Short).To(Equal("Custom short"))
		Expect(rootCmd.Version).To(Equal("2.0.0"))
	})

	It("should apply custom PersistentPreRun hook", func() {
		hookCalled := false
		rootCmd := cli.BuildRootCommand(ctx,
			cli.WithPersistentPreRun(func(cmd *cobra.Command, args []string) {
				hookCalled = true
			}),
		)

		Expect(rootCmd.PersistentPreRun).NotTo(BeNil())
		rootCmd.PersistentPreRun(rootCmd, []string{})
		Expect(hookCalled).To(BeTrue())
	})

	It("should apply custom PersistentPostRun hook", func() {
		hookCalled := false
		rootCmd := cli.BuildRootCommand(ctx,
			cli.WithPersistentPostRun(func(cmd *cobra.Command, args []string) {
				hookCalled = true
			}),
		)

		Expect(rootCmd.PersistentPostRun).NotTo(BeNil())
		rootCmd.PersistentPostRun(rootCmd, []string{})
		Expect(hookCalled).To(BeTrue())
	})
})

var _ = Describe("RegisterAllCommands", func() {
	var ctx *context.Context
	var rootCmd *cobra.Command

	BeforeEach(func() {
		bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
		ctx = context.NewContext(bkgCtx, cancelFunc, io.Stdin, io.Stdout)
		rootCmd = cli.BuildRootCommand(ctx)
	})

	AfterEach(func() {
		if ctx != nil {
			ctx.Finalize()
		}
	})

	It("should register all commands", func() {
		cli.RegisterAllCommands(ctx, rootCmd)

		commands := rootCmd.Commands()
		Expect(len(commands)).To(BeNumerically(">", 0))

		// Check for key commands
		commandNames := make(map[string]bool)
		for _, cmd := range commands {
			commandNames[cmd.Name()] = true
		}

		Expect(commandNames).To(HaveKey("exec"))
		Expect(commandNames).To(HaveKey("workspace"))
		Expect(commandNames).To(HaveKey("config"))
	})
})

var _ = Describe("Individual Command Builders", func() {
	var ctx *context.Context

	BeforeEach(func() {
		bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
		ctx = context.NewContext(bkgCtx, cancelFunc, io.Stdin, io.Stdout)
	})

	AfterEach(func() {
		if ctx != nil {
			ctx.Finalize()
		}
	})

	It("should build exec command", func() {
		cmd := cli.BuildExecCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("exec"))
	})

	It("should build browse command", func() {
		cmd := cli.BuildBrowseCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("browse"))
	})

	It("should build config command", func() {
		cmd := cli.BuildConfigCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("config"))
	})

	It("should build secret command", func() {
		cmd := cli.BuildSecretCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("secret"))
	})

	It("should build vault command", func() {
		cmd := cli.BuildVaultCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("vault"))
	})

	It("should build cache command", func() {
		cmd := cli.BuildCacheCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("cache"))
	})

	It("should build workspace command", func() {
		cmd := cli.BuildWorkspaceCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("workspace"))
	})

	It("should build template command", func() {
		cmd := cli.BuildTemplateCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("template"))
	})

	It("should build logs command", func() {
		cmd := cli.BuildLogsCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("logs"))
	})

	It("should build sync command", func() {
		cmd := cli.BuildSyncCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("sync"))
	})

	It("should build mcp command", func() {
		cmd := cli.BuildMCPCommand(ctx)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.Name()).To(Equal("mcp"))
	})
})
