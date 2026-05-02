package fileparser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/fileparser"
	"github.com/flowexec/flow/v2/types/executable"
)

var _ = Describe("ExecutablesFromShFile", func() {
	const filePath = "testdata/simple.sh"

	It("should parse executables from sh file", func() {
		exec, err := fileparser.ExecutablesFromShFile("testdata", filePath)
		Expect(err).NotTo(HaveOccurred())
		Expect(exec).NotTo(BeNil())
		Expect(exec.Verb).To(Equal(executable.VerbShow))
		Expect(exec.Name).To(Equal("hello"))
		Expect(exec.Exec).NotTo(BeNil())
		Expect(exec.Exec.File).To(Equal("simple.sh"))
		Expect(exec.Exec.Dir).To(Equal(executable.Directory("//")))
	})
})
