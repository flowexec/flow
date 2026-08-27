//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os/exec"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/tests/utils"
)

// canRunLinuxContainers reports whether a runtime capable of running the Linux
// test image is available and actually reachable. Windows hosts are excluded:
// docker.exe may be on PATH there (e.g. Windows CI), but it cannot run a Linux
// image, so the test would fail rather than exercise the feature.
//
// Mirrors the runner's own "auto" precedence (docker before podman, see
// internal/services/run/container.go ResolveRuntime): a binary on PATH isn't
// enough (Docker Desktop can be installed but not running), so the runtime
// that "auto" would actually pick also has to respond before we call it
// available — otherwise the test fails instead of skipping.
func canRunLinuxContainers() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	for _, rt := range []string{"docker", "podman"} {
		if _, err := exec.LookPath(rt); err != nil {
			continue
		}
		return runtimeIsLive(rt)
	}
	return false
}

func runtimeIsLive(rt string) bool {
	ctx, cancel := stdCtx.WithTimeout(stdCtx.Background(), 3*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, rt, "info").Run() == nil
}

var _ = Describe("container exec e2e", func() {
	var ctx *utils.Context

	BeforeEach(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	When("a container runtime is available", func() {
		BeforeEach(func() {
			if !canRunLinuxContainers() {
				Skip("no Linux-capable container runtime available (docker/podman)")
			}
		})

		It("runs python inside the container using the image's interpreter", func() {
			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(ctx.Context, "exec", "examples:with-python-container",
				"--log-level", "debug")).To(Succeed())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("hello from with-python-container"))
			Expect(out).To(ContainSubstring("in-container=true"))
			// A real interpreter ran, and it was the image's - the host venv, if
			// any, is deliberately not carried in.
			Expect(out).To(ContainSubstring("py-major=3"))
		})

		It("runs the command inside the container", func() {
			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(ctx.Context, "exec", "examples:with-container", "--log-level", "debug")).To(Succeed())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("hello from with-container"))
			Expect(out).To(ContainSubstring("in-container=true"))
			// The workspace is auto-mounted at /workspace and used as the workdir.
			Expect(out).To(ContainSubstring("/workspace"))
		})
	})
})
