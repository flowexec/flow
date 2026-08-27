//go:build e2e

package tests_test

import (
	stdCtx "context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/tests/utils"
)

// canRunPython reports whether a python interpreter is reachable, mirroring the
// runner's own PATH precedence (see internal/services/run/python.go
// pathCandidates): Windows prefers "python" because "python3" there is usually
// the Microsoft Store alias stub rather than a real interpreter.
//
// CI pins an interpreter with actions/setup-python, so this guard should never
// skip there; it only spares a developer machine without python installed.
func canRunPython() bool {
	candidates := []string{"python3", "python"}
	if runtime.GOOS == "windows" {
		candidates = []string{"python", "python3"}
	}
	for _, name := range candidates {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

// pySpec builds a --spec argument for a transient python executable.
func pySpec(name string, e map[string]any) string {
	spec := map[string]any{"verb": "run", "name": name, "exec": e}
	out, err := json.Marshal(spec)
	Expect(err).NotTo(HaveOccurred())
	return string(out)
}

var _ = Describe("python exec e2e", func() {
	var ctx *utils.Context

	BeforeEach(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		if !canRunPython() {
			Skip("no python interpreter available on PATH")
		}
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	When("an executable sets interpreter: python", func() {
		It("runs the command as python rather than shell", func() {
			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(ctx.Context, "exec", "examples:with-python")).To(Succeed())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("hello from with-python"))
			// Proves a real interpreter ran: flow's shell interpreter would not
			// evaluate sys.version_info.
			Expect(out).To(ContainSubstring("py-major=3"))
		})

		It("reports a traceback with the failing line when the script raises", func() {
			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			ctx.ExpectFailure()
			err := runner.Run(ctx.Context, "exec", "--spec", pySpec("py-raise", map[string]any{
				"interpreter": "python",
				"cmd":         "x = 1\ny = 2\nraise ValueError('boom')\n",
			}))
			Expect(err).To(HaveOccurred())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("ValueError: boom"))
			// Running from a temp file preserves line numbers; `python -c` would not.
			Expect(out).To(ContainSubstring("line 3"))
		})

		It("passes flow-resolved parameters through the environment", func() {
			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(ctx.Context, "exec", "--spec", pySpec("py-params", map[string]any{
				"interpreter": "python",
				"params":      []map[string]any{{"envKey": "GREETING", "text": "from-params"}},
				"cmd":         "import os\nprint('greeting=' + os.environ['GREETING'])\n",
			}))).To(Succeed())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("greeting=from-params"))
		})
	})

	When("running an ad-hoc command with --interpreter", func() {
		It("runs the command as python", func() {
			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(ctx.Context, "exec",
				"--interpreter", "python",
				"--cmd", "import sys; print('adhoc py', sys.version_info[0])",
			)).To(Succeed())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("adhoc py 3"))
		})

		It("rejects an unknown interpreter instead of falling back to shell", func() {
			runner := utils.NewE2ECommandRunner()
			ctx.ExpectFailure()
			err := runner.Run(ctx.Context, "exec", "--interpreter", "ruby", "--cmd", "puts 1")
			Expect(err).To(HaveOccurred())
		})

		It("rejects --interpreter with multiple commands", func() {
			// Serial/parallel steps carry no interpreter, so only the single-command
			// form can honour the flag.
			runner := utils.NewE2ECommandRunner()
			ctx.ExpectFailure()
			err := runner.Run(ctx.Context, "exec", "--interpreter", "python",
				"--cmd", "print(1)", "--cmd", "print(2)")
			Expect(err).To(HaveOccurred())
		})
	})

	When("an executable runs a .py file", func() {
		It("infers the python interpreter from the extension", func() {
			dir := ctx.WorkspaceDir()
			Expect(os.WriteFile(
				filepath.Join(dir, "e2e_script.py"),
				[]byte("print('hello from py file')\n"), 0600,
			)).To(Succeed())

			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(ctx.Context, "exec", "--spec", pySpec("py-file", map[string]any{
				"file": "e2e_script.py",
				"dir":  filepath.ToSlash(dir),
			}))).To(Succeed())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("hello from py file"))
		})
	})

	When("the configured interpreter cannot be found", func() {
		It("fails with an actionable error naming what was tried", func() {
			GinkgoTB().Setenv("FLOW_PYTHON_BIN", filepath.Join(ctx.WorkspaceDir(), "not-python"))

			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			ctx.ExpectFailure()
			err := runner.Run(ctx.Context, "exec", "--spec", pySpec("py-missing", map[string]any{
				"interpreter": "python",
				"cmd":         "print(1)\n",
			}))
			Expect(err).To(HaveOccurred())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("no python interpreter found"))
		})
	})
})
