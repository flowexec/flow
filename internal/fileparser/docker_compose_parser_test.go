package fileparser_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/fileparser"
)

var _ = Describe("ExecutablesFromDockerCompose", func() {
	const composePath = "testdata/docker-compose.yml"

	It("should parse docker-compose.yml", func() {
		execs, err := fileparser.ExecutablesFromDockerCompose("", composePath)
		Expect(err).NotTo(HaveOccurred())
		Expect(execs).To(HaveLen(6))

		found := map[string]bool{
			"start":       false,
			"stop":        false,
			"start app":   false,
			"start db":    false,
			"start redis": false,
			"build app":   false,
		}

		for _, e := range execs {
			Expect(e.Exec).NotTo(BeNil())
			Expect(e.Exec.Cmd).To(ContainSubstring("docker-compose"))

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
