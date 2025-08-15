//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/tests/utils"
)

var _ = Describe("exec e2e", func() {
	var (
		ctx *utils.Context
	)

	BeforeEach(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	DescribeTable("with dir example executables", func(ref string) {
		runner := utils.NewE2ECommandRunner()
		stdOut := ctx.StdOut()
		Expect(runner.Run(ctx.Context, "exec", ref, "--log-level", "debug")).To(Succeed())
		Expect(readFileContent(stdOut)).To(ContainSubstring("flow completed"))
	},
		Entry("print example", "examples:simple-print"),
		Entry("tmp dir example", "examples:with-tmp-dir"),
		Entry("nameless example", ""),
		Entry("request with transformation", "examples:request-with-transform"),
	)

	When("param overrides are provided", func() {
		It("should run the executable with the provided overrides", func() {
			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(
				ctx.Context, "exec", "examples:with-params",
				"--param", "PARAM1=value1", "--param", "PARAM2=value2", "--param", "PARAM3=value3",
			)).To(Succeed())
			out, _ := readFileContent(stdOut)
			Expect(out).To(ContainSubstring("value1"))
			Expect(out).To(ContainSubstring("value2"))
			Expect(out).To(ContainSubstring("value3"))
		})
	})

	Describe("file parameter and argument output files", func() {
		It("should create temporary files for file arguments", func() {
			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(ctx.Context, "exec", "examples:with-file-param", "argval", "--log-level", "debug")).To(Succeed())
			out, _ := readFileContent(stdOut)

			Expect(out).To(ContainSubstring("database:"))
			Expect(out).To(ContainSubstring("host: localhost"))
			Expect(out).To(ContainSubstring("port: 5432"))
			Expect(out).To(ContainSubstring("#!/bin/bash"))
			Expect(out).To(ContainSubstring("Hello from script"))

			// Should show passed in arg instead of default
			Expect(out).To(ContainSubstring("argval"))
			Expect(out).NotTo(ContainSubstring("notme"))

			Expect(out).To(ContainSubstring("flow completed"))
		})
	})

	Describe("workspace envFile feature", func() {
		It("should inherit environment variables from workspace .env file", func() {
			envFilePath := filepath.Join(ctx.WorkspaceDir(), ".env")
			envContent := "WORKSPACE_ENV_VAR=test_value_from_env"
			err := os.WriteFile(envFilePath, []byte(envContent), 0644)
			Expect(err).ToNot(HaveOccurred())

			runner := utils.NewE2ECommandRunner()
			stdOut := ctx.StdOut()
			Expect(runner.Run(ctx.Context, "exec", "examples:with-workspace-env", "--log-level", "debug")).To(Succeed())
			out, _ := readFileContent(stdOut)

			Expect(out).To(ContainSubstring("WORKSPACE_ENV_VAR=test_value_from_env"))
			Expect(out).To(ContainSubstring("flow completed"))
		})
	})
})
