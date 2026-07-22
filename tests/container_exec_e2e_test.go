//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os/exec"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/tests/utils"
)

// canRunLinuxContainers reports whether a runtime capable of running the Linux
// test image is available. Windows hosts are excluded: docker.exe may be on PATH
// there (e.g. Windows CI), but it cannot run a Linux image, so the test would
// fail rather than exercise the feature.
func canRunLinuxContainers() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	for _, rt := range []string{"docker", "podman"} {
		if _, err := exec.LookPath(rt); err == nil {
			return true
		}
	}
	return false
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
