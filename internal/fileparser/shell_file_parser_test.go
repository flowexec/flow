package fileparser_test

import (
	"os"
	"path/filepath"

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

	It("uses forward slashes for a nested directory on every platform", func() {
		ws := GinkgoT().TempDir()
		dir := filepath.Join(ws, "scripts", "ci")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		path := filepath.Join(dir, "simple.sh")
		Expect(os.WriteFile(path, []byte("#!/bin/sh\necho hi\n"), 0o600)).To(Succeed())

		exec, err := fileparser.ExecutablesFromShFile(ws, path)
		Expect(err).NotTo(HaveOccurred())
		Expect(exec.Exec.Dir).To(Equal(executable.Directory("//scripts/ci")))
	})

	It("names the executable from the file name without its extension", func() {
		ws := GinkgoT().TempDir()
		for file, ref := range map[string]string{
			"lint-go.sh":      "lint go",
			"docker-build.sh": "build docker-build",
		} {
			path := filepath.Join(ws, file)
			Expect(os.WriteFile(path, []byte("#!/bin/sh\necho hi\n"), 0o600)).To(Succeed())

			exec, err := fileparser.ExecutablesFromShFile(ws, path)
			Expect(err).NotTo(HaveOccurred())
			Expect(exec.Verb.String() + " " + exec.Name).To(Equal(ref))
			Expect(exec.Exec.File).To(Equal(file))
		}
	})
})
