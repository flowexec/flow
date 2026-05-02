package fileparser_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/fileparser"
)

var _ = Describe("ExecutablesFromMakefile", func() {
	const makefile = "testdata/Makefile"

	It("should parse Makefile", func() {
		execs, err := fileparser.ExecutablesFromMakefile("", makefile)
		Expect(err).NotTo(HaveOccurred())
		Expect(execs).To(HaveLen(4))

		found := map[string]bool{
			"build":       false,
			"test":        false,
			"deploy":      false,
			"run program": false,
		}
		expectedDesc := map[string]string{
			"build":       "Build the application binary",
			"test":        "Run all tests with coverage",
			"deploy":      "Deploy to production environment\nDepends on build and test",
			"run program": "Run main.go",
		}

		for _, e := range execs {
			Expect(e.Exec).NotTo(BeNil())
			Expect(e.Exec.Cmd).To(ContainSubstring("make"))

			shortRef := strings.TrimSpace(fmt.Sprintf("%s %s", e.Verb, e.Name))
			if _, ok := found[shortRef]; ok {
				found[shortRef] = true
			}
			Expect(e.Description).To(Equal(expectedDesc[shortRef]))
		}

		for ref, found := range found {
			Expect(found).To(BeTrue(), "executable %s not found", ref)
		}
	})
})
