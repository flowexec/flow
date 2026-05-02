package fileparser_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/fileparser"
)

var _ = Describe("ExecutablesFromPackageJSON", func() {
	const pkgPath = "testdata/package.json"

	It("should parse executables from package.json", func() {
		execs, err := fileparser.ExecutablesFromPackageJSON("", pkgPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(execs).To(HaveLen(7))

		found := map[string]bool{
			"install":       false,
			"start dev":     false,
			"build":         false,
			"test":          false,
			"test watch":    false,
			"lint":          false,
			"start preview": false,
		}

		for _, e := range execs {
			Expect(e.Exec).NotTo(BeNil())
			Expect(e.Exec.Cmd).To(ContainSubstring("npm"))

			shortRef := strings.TrimSpace(fmt.Sprintf("%s %s", e.Verb, e.Name))
			if _, ok := found[shortRef]; ok {
				found[shortRef] = true
			}
		}
		for ref, found := range found {
			Expect(found).To(BeTrue(), "executable %s not found", ref)
		}
	})
})
