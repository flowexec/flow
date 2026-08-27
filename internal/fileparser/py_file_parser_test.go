package fileparser_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/fileparser"
	"github.com/flowexec/flow/v2/types/executable"
)

var _ = Describe("ExecutablesFromPyFile", func() {
	const wsPath = "testdata"

	It("should parse a simple py file", func() {
		exec, err := fileparser.ExecutablesFromPyFile(wsPath, "testdata/simple.py")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec).NotTo(BeNil())
		Expect(exec.Name).To(Equal("hello"))
		Expect(exec.Verb).To(Equal(executable.VerbShow))
		Expect(exec.Exec).NotTo(BeNil())
		Expect(exec.Exec.File).To(Equal("simple.py"))
		Expect(exec.Exec.Dir).To(Equal(executable.Directory("//")))
		Expect(exec.Tags).To(ContainElement("generated"))
	})

	It("should leave the interpreter unset, since .py already implies it", func() {
		exec, err := fileparser.ExecutablesFromPyFile(wsPath, "testdata/simple.py")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec.Exec.InterpreterIsSet()).To(BeFalse())
		// The extension is what routes it, so the executable still runs as python.
		Expect(exec.Exec.InterpreterForFile()).To(Equal(executable.InterpreterPython))
	})

	It("should parse a complex py file with all metadata", func() {
		exec, err := fileparser.ExecutablesFromPyFile(wsPath, "testdata/complex.py")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec).NotTo(BeNil())
		Expect(exec.Name).To(Equal("deploy"))
		Expect(exec.Verb).To(Equal(executable.VerbDeploy))
		Expect(exec.Description).To(Equal("Deploy to production"))
		Expect(exec.Tags).To(ContainElements("production", "critical", "generated"))
		expectedTimeout := 10 * time.Minute
		Expect(exec.Timeout).To(Equal(&expectedTimeout))
	})

	It("should parse params from a py file", func() {
		exec, err := fileparser.ExecutablesFromPyFile(wsPath, "testdata/params.py")
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

	It("should read metadata that follows a shebang line", func() {
		// A shebang is idiomatic in python scripts, so it must not shadow the
		// metadata comments beneath it.
		exec, err := fileparser.ExecutablesFromPyFile(wsPath, "testdata/shebang.py")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec.Name).To(Equal("with-shebang"))
		Expect(exec.Verb).To(Equal(executable.VerbRun))
	})

	It("should error on a missing file", func() {
		_, err := fileparser.ExecutablesFromPyFile(wsPath, "testdata/does-not-exist.py")
		Expect(err).To(HaveOccurred())
	})
})
