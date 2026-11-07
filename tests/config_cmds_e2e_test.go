//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/tests/utils"
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

		It("should accept a nested namespace", func() {
			Expect(run.Run(ctx.Context, "config", "set", "namespace", "parent/child")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Namespace set to parent/child"))
		})

		It("should reject an invalid namespace", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "config", "set", "namespace", "parent//child")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("invalid namespace")))
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

	When("setting an invalid timeout (flow config set timeout bogus)", func() {
		It("fatals with an invalid duration message", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "config", "set", "timeout", "bogus")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("invalid duration")))
		})
	})

	When("setting log retention (flow config set log-retention)", func() {
		It("sets the given limits and keeps the rest", func() {
			Expect(run.Run(ctx.Context, "config", "set", "log-retention", "--max-age", "168h", "--max-count", "0")).
				To(Succeed())
			utils.ResetTestContext(ctx, GinkgoTB())
			Expect(run.Run(ctx.Context, "config", "set", "log-retention", "--max-size-mb", "100")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(
				"Log retention set: max age 168h0m0s, max size 100 MB, max count unlimited"))
		})

		It("fatals on an invalid duration", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "config", "set", "log-retention", "--max-age", "bogus")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("invalid --max-age")))
		})

		It("fatals on a negative limit", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "config", "set", "log-retention", "--max-count", "-1")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("must not be negative")))
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

	When("setting the container runtime (flow config set container-runtime)", func() {
		It("sets what auto uses on this machine", func() {
			Expect(run.Run(ctx.Context, "config", "set", "container-runtime", "podman")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Containers set: runtime podman"))
		})

		It("rejects an unknown runtime", func() {
			ctx.ExpectFailure()
			Expect(run.Run(ctx.Context, "config", "set", "container-runtime", "lxc")).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("invalid container runtime")))
		})
	})

	When("setting a runtime's command (flow config set container-command)", func() {
		It("stores global flags given after --", func() {
			Expect(run.Run(ctx.Context,
				"config", "set", "container-command", "docker", "docker", "--", "--context", "colima")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("docker runs `docker --context colima`"))
		})

		// Runs after the spec above (the suite is Ordered), so it clears the flags set there; run
		// alone it still checks the result, since each spec can run one command.
		It("goes back to the default when set to the runtime's own name", func() {
			Expect(run.Run(ctx.Context, "config", "set", "container-command", "docker", "docker")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			lines := strings.Split(strings.TrimSpace(out), "\n")
			last := lines[len(lines)-1]
			Expect(last).To(ContainSubstring("Containers set:"))
			Expect(last).NotTo(ContainSubstring("docker runs"))
		})

		It("points at -- when a runtime flag is given without it", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "config", "set", "container-command", "docker", "docker", "--context", "colima")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("after `--`")))
		})

		It("rejects an unknown runtime family", func() {
			ctx.ExpectFailure()
			Expect(run.Run(ctx.Context, "config", "set", "container-command", "lxc", "lxc")).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("invalid container runtime")))
		})
	})
})
