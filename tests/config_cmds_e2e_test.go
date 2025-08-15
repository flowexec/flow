//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/tests/utils"
)

var _ = Describe("config e2e", Ordered, func() {
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

	When("getting configuration (flow config get)", func() {
		It("should display configuration in yaml format", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "config", "get", "-o", "yaml")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("currentWorkspace:"))
		})
	})

	When("setting namespace (flow config set namespace)", func() {
		It("should set the namespace successfully", func() {
			Expect(run.Run(ctx.Context, "config", "set", "namespace", "test-namespace")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Namespace set to test-namespace"))
		})
	})

	When("setting workspace mode (flow config set workspace-mode)", func() {
		It("should set workspace mode to fixed", func() {
			Expect(run.Run(ctx.Context, "config", "set", "workspace-mode", "fixed")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Workspace mode set to 'fixed'"))
		})

		It("should set workspace mode to dynamic", func() {
			Expect(run.Run(ctx.Context, "config", "set", "workspace-mode", "dynamic")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Workspace mode set to 'dynamic'"))
		})
	})

	When("setting log mode (flow config set log-mode)", func() {
		It("should set log mode to logfmt", func() {
			Expect(run.Run(ctx.Context, "config", "set", "log-mode", "logfmt")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Default log mode set to 'logfmt'"))
		})

		It("should set log mode to json", func() {
			Expect(run.Run(ctx.Context, "config", "set", "log-mode", "json")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Default log mode set to 'json'"))
		})

		It("should set log mode to text", func() {
			Expect(run.Run(ctx.Context, "config", "set", "log-mode", "text")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Default log mode set to 'text'"))
		})

		It("should set log mode to hidden", func() {
			Expect(run.Run(ctx.Context, "config", "set", "log-mode", "hidden")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Default log mode set to 'hidden'"))
		})
	})

	When("setting TUI (flow config set tui)", func() {
		It("should enable TUI", func() {
			Expect(run.Run(ctx.Context, "config", "set", "tui", "true")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Interactive UI enabled"))
		})

		It("should disable TUI", func() {
			Expect(run.Run(ctx.Context, "config", "set", "tui", "false")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Interactive UI disabled"))
		})
	})

	When("setting notifications (flow config set notifications)", func() {
		It("should enable notifications", func() {
			Expect(run.Run(ctx.Context, "config", "set", "notifications", "true")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Notifications enabled"))
		})

		It("should disable notifications", func() {
			Expect(run.Run(ctx.Context, "config", "set", "notifications", "false")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Notifications disabled"))
		})
	})

	When("setting theme (flow config set theme)", func() {
		It("should set theme to light", func() {
			Expect(run.Run(ctx.Context, "config", "set", "theme", "light")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Theme set to light"))
		})

		It("should set theme to dark", func() {
			Expect(run.Run(ctx.Context, "config", "set", "theme", "dark")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Theme set to dark"))
		})

		It("should set theme to default", func() {
			Expect(run.Run(ctx.Context, "config", "set", "theme", "default")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Theme set to default"))
		})
	})

	When("setting timeout (flow config set timeout)", func() {
		It("should set timeout to a valid duration", func() {
			Expect(run.Run(ctx.Context, "config", "set", "timeout", "30s")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Default timeout set to 30s"))
		})

		It("should set timeout to minutes", func() {
			Expect(run.Run(ctx.Context, "config", "set", "timeout", "5m")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Default timeout set to 5m"))
		})
	})

	When("resetting configuration (flow config reset)", func() {
		It("should prompt for confirmation and reset config", func() {
			reader, writer, err := os.Pipe()
			Expect(err).NotTo(HaveOccurred())
			_, err = writer.Write([]byte("true\n"))
			Expect(err).ToNot(HaveOccurred())

			ctx.SetIO(reader, ctx.StdOut())
			Expect(run.Run(ctx.Context, "config", "reset")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Restored flow configurations"))
		})

		It("should abort reset when confirmation is false", func() {
			reader, writer, err := os.Pipe()
			Expect(err).NotTo(HaveOccurred())
			_, err = writer.Write([]byte("false\n"))
			Expect(err).ToNot(HaveOccurred())

			ctx.SetIO(reader, ctx.StdOut())
			Expect(run.Run(ctx.Context, "config", "reset")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Aborting"))
		})
	})
})
