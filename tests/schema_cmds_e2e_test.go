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

var _ = Describe("schema e2e", Ordered, func() {
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

	When("validating a valid flow file (flow schema validate)", func() {
		It("should succeed", func() {
			flowFile := filepath.Join(ctx.WorkspaceDir(), "examples.flow")
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "schema", "validate", flowFile)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("1 file(s) valid"))
		})
	})

	When("validating multiple valid flow files (flow schema validate)", func() {
		It("should validate all files", func() {
			file1 := filepath.Join(ctx.WorkspaceDir(), "examples.flow")
			file2 := filepath.Join(ctx.WorkspaceDir(), "root.flow")
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "schema", "validate", file1, file2)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("2 file(s) valid"))
		})
	})

	When("validating a valid workspace config (flow schema validate)", func() {
		It("should succeed", func() {
			wsConfig := filepath.Join(ctx.WorkspaceDir(), "flow.yaml")
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "schema", "validate", wsConfig)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("1 file(s) valid"))
		})
	})

	When("validating with strict mode (flow schema validate --strict)", func() {
		It("should succeed on a valid flow file with no extra keys", func() {
			flowFile := filepath.Join(ctx.WorkspaceDir(), "examples.flow")
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "schema", "validate", "--strict", flowFile)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("1 file(s) valid"))
		})
	})

	When("validating an invalid flow file (flow schema validate)", func() {
		It("should fail with validation errors for wrong types", func() {
			invalidFile := filepath.Join(ctx.WorkspaceDir(), "invalid.flow")
			Expect(os.WriteFile(invalidFile, []byte("namespace: 123\n"), 0600)).To(Succeed())

			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "schema", "validate", invalidFile)
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("validation failed")))
		})
	})

	When("validating with strict mode detects unknown keys", func() {
		It("should fail when extra keys are present", func() {
			fileWithExtraKey := filepath.Join(ctx.WorkspaceDir(), "extra.flow")
			content := "namespace: test\nunknownField: true\n"
			Expect(os.WriteFile(fileWithExtraKey, []byte(content), 0600)).To(Succeed())

			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "schema", "validate", "--strict", fileWithExtraKey)
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("additional properties")))
		})

		It("should pass without strict mode when extra keys are present", func() {
			fileWithExtraKey := filepath.Join(ctx.WorkspaceDir(), "extra.flow")
			content := "namespace: test\nunknownField: true\n"
			Expect(os.WriteFile(fileWithExtraKey, []byte(content), 0600)).To(Succeed())

			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "schema", "validate", fileWithExtraKey)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("1 file(s) valid"))
		})
	})

	When("validating with --type override (flow schema validate --type)", func() {
		It("should validate a workspace config with explicit type", func() {
			wsConfig := filepath.Join(ctx.WorkspaceDir(), "flow.yaml")
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "schema", "validate", "--type", "workspace", wsConfig)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("1 file(s) valid"))
		})

		It("should fail with an invalid --type value", func() {
			flowFile := filepath.Join(ctx.WorkspaceDir(), "examples.flow")
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "schema", "validate", "--type", "badtype", flowFile)
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("invalid --type")))
		})
	})

	When("validating with JSON output (flow schema validate --output json)", func() {
		It("should output valid JSON on success", func() {
			flowFile := filepath.Join(ctx.WorkspaceDir(), "examples.flow")
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "schema", "validate", "--output", "json", flowFile)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(`"message":"1 file(s) valid"`))
		})

		It("should output JSON error envelope on failure", func() {
			invalidFile := filepath.Join(ctx.WorkspaceDir(), "bad.flow")
			Expect(os.WriteFile(invalidFile, []byte("namespace: 123\n"), 0600)).To(Succeed())

			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "schema", "validate", "--output", "json", invalidFile)
			Expect(err).To(HaveOccurred())
			out, readErr := readFileContent(ctx.StdOut())
			Expect(readErr).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(`"code":"VALIDATION_FAILED"`))
		})
	})
})
