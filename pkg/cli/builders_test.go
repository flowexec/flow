package cli_test

import (
	stdCtx "context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

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
		r, w, _ := os.Pipe()
		ctx = context.NewContext(bkgCtx, cancelFunc, r, w)
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
})

var _ = Describe("RegisterAllCommands", func() {
	var ctx *context.Context
	var rootCmd *cobra.Command

	BeforeEach(func() {
		bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
		r, w, _ := os.Pipe()
		ctx = context.NewContext(bkgCtx, cancelFunc, r, w)
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
		Expect(commands).ToNot(BeEmpty())

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
