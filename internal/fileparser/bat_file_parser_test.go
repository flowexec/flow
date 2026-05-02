package fileparser_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/fileparser"
	"github.com/flowexec/flow/v2/types/executable"
)

var _ = Describe("ExecutablesFromBatFile", func() {
	const wsPath = "testdata"

	It("should parse a simple bat file with REM comments", func() {
		exec, err := fileparser.ExecutablesFromBatFile(wsPath, "testdata/simple.bat")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec).NotTo(BeNil())
		Expect(exec.Name).To(Equal("hello"))
		Expect(exec.Verb).To(Equal(executable.VerbShow))
		Expect(exec.Exec).NotTo(BeNil())
		Expect(exec.Exec.File).To(Equal("simple.bat"))
		Expect(exec.Exec.Dir).To(Equal(executable.Directory("//")))
		Expect(exec.Tags).To(ContainElement("generated"))
	})

	It("should parse a complex bat file with :: comments", func() {
		exec, err := fileparser.ExecutablesFromBatFile(wsPath, "testdata/complex.bat")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec).NotTo(BeNil())
		Expect(exec.Name).To(Equal("deploy"))
		Expect(exec.Verb).To(Equal(executable.VerbDeploy))
		Expect(exec.Description).To(Equal("Deploy to production"))
		Expect(exec.Tags).To(ContainElements("production", "critical", "generated"))
		expectedTimeout := 10 * time.Minute
		Expect(exec.Timeout).To(Equal(&expectedTimeout))
	})

	It("should parse params from a bat file", func() {
		exec, err := fileparser.ExecutablesFromBatFile(wsPath, "testdata/params.bat")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec).NotTo(BeNil())
		Expect(exec.Name).To(Equal("test-params"))
		Expect(exec.Exec.Params).To(HaveLen(3))
		Expect(exec.Exec.Params[0].SecretRef).To(Equal("my-secret"))
		Expect(exec.Exec.Params[0].EnvKey).To(Equal("SECRET_VAR"))
		Expect(exec.Exec.Params[1].Prompt).To(Equal("Enter name"))
		Expect(exec.Exec.Params[1].EnvKey).To(Equal("NAME_VAR"))
		Expect(exec.Exec.Params[2].Text).To(Equal("default-value"))
		Expect(exec.Exec.Params[2].EnvKey).To(Equal("DEFAULT_VAR"))
	})
})
