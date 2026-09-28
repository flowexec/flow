package fileparser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/internal/fileparser"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
)

func TestFileParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileParser Suite")
}

var _ = Describe("ExecutablesFromImports", func() {
	var (
		ctrl       *gomock.Controller
		mockLogger *mocks.MockLogger
		flowFile   *executable.FlowFile
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockLogger = mocks.NewMockLogger(ctrl)
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		wd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		ff := filepath.Join(wd, "testdata", "test"+executable.FlowFileExt)
		flowFile = &executable.FlowFile{Imports: make(executable.Imports, 0)}
		flowFile.SetContext("ws", filepath.Join(wd, "testdata"), ff)
	})

	It("should return executables from imports", func() {
		flowFile.Imports = append(
			flowFile.Imports,
			"Makefile",
			"package.json",
			"docker-compose.yml",
			"complex.sh",
		)

		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result)).To(BeNumerically(">", 10))

		for _, e := range result {
			Expect(e.Exec).ToNot(BeNil())
			Expect(e.Exec.Dir).To(Equal(executable.Directory("//")))
		}
	})

	It("records the imported file each executable came from", func() {
		flowFile.Imports = append(flowFile.Imports, "Makefile", "complex.sh")

		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())

		wsPath := flowFile.WorkspacePath()
		for _, e := range result {
			Expect(e.FlowFilePath()).To(Equal(flowFile.ConfigPath()))
			Expect(e.SourceFilePath()).To(BeElementOf(
				filepath.Join(wsPath, "Makefile"),
				filepath.Join(wsPath, "complex.sh"),
			))
		}
	})

	It("should return executables from bat file imports", func() {
		flowFile.Imports = append(flowFile.Imports, "simple.bat")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Exec.File).To(Equal("simple.bat"))
	})

	It("should return executables from ps1 file imports", func() {
		flowFile.Imports = append(flowFile.Imports, "simple.ps1")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Exec.File).To(Equal("simple.ps1"))
	})

	It("should return executables from py file imports", func() {
		flowFile.Imports = append(flowFile.Imports, "simple.py")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Exec.File).To(Equal("simple.py"))
	})

	It("should log a warning for invalid file type", func() {
		mockLogger.EXPECT().Warn(gomock.Any(), "file", "invalidfile").AnyTimes()
		flowFile.Imports = append(flowFile.Imports, "invalidfile")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	It("should log an error for dir instead of file", func() {
		mockLogger.EXPECT().Error(gomock.Any(), "err", "invaliddir is not a file").AnyTimes()
		flowFile.Imports = append(flowFile.Imports, "invaliddir")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	It("should log an error for non-existent file", func() {
		mockLogger.EXPECT().WrapError(gomock.Any(), gomock.Any()).AnyTimes()
		flowFile.Imports = append(flowFile.Imports, "nonexistent.sh")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	It("should log an error when configuration key is not recognized", func() {
		mockLogger.EXPECT().WrapError(gomock.Any(), gomock.Any()).AnyTimes()
		flowFile.Imports = append(flowFile.Imports, "unknownkey.sh")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).ToNot(BeNil())
	})
})

var _ = Describe("InferVerb", func() {
	DescribeTable("infers the verb from a name",
		func(name string, expected executable.Verb) {
			Expect(fileparser.InferVerb(name)).To(Equal(expected))
		},
		Entry("a verb", "lint", executable.VerbLint),
		Entry("a verb prefix", "build-app", executable.VerbBuild),
		Entry("a verb as a later word", "docker-build", executable.VerbBuild),
		Entry("a verb after a colon", "db:migrate:test", executable.VerbTest),
		Entry("vet", "vet", executable.VerbLint),
		Entry("a verb inside a word", "rebuilder", executable.VerbExec),
		Entry("no verb", "docker", executable.VerbExec),
		Entry("a verb in camelCase", "prodBuild", executable.VerbBuild),
		Entry("any valid verb as a word", "db:migrate", executable.VerbMigrate),
		Entry("any valid verb as the first word", "backup-db", executable.VerbBackup),
		Entry("a synonym", "docker-up", executable.VerbStart),
		Entry("an npm pre hook", "prebuild", executable.VerbBuild),
		Entry("an npm post hook", "postinstall", executable.VerbInstall),
		Entry("a noun-like verb as a word", "search-index", executable.VerbExec),
		Entry("a noun-like verb as the whole name", "index", executable.VerbIndex),
		Entry("nouns before a verb-like word", "package-lock", executable.VerbExec),
	)

	It("infers a custom verb from a word", func() {
		executable.RegisterCustomVerbs("sync")
		DeferCleanup(executable.RegisterCustomVerbs)
		Expect(fileparser.InferVerb("sync-assets")).To(Equal(executable.Verb("sync")))
	})

	// The word fallback only runs where the pattern passes fall back to exec, so names they
	// already matched keep their verb even when a later word is a more specific one.
	DescribeTable("keeps the verb the pattern passes infer",
		func(name string, expected executable.Verb) {
			Expect(fileparser.InferVerb(name)).To(Equal(expected))
		},
		Entry(nil, "compile-assets", executable.VerbBuild),
		Entry(nil, "publish-docs", executable.VerbDeploy),
		Entry(nil, "reset-db", executable.VerbClean),
		Entry(nil, "buildProd", executable.VerbBuild),
		Entry(nil, "test:e2e", executable.VerbTest),
		Entry(nil, "preview", executable.VerbStart),
		Entry(nil, "docker-push", executable.VerbDeploy),
	)
})

var _ = Describe("NormalizeName", func() {
	DescribeTable("normalizes a name",
		func(name, verb, expected string) {
			Expect(fileparser.NormalizeName(name, verb)).To(Equal(expected))
		},
		Entry("drops the verb prefix", "build-app", "build", "app"),
		Entry("keeps a later verb", "docker-build", "build", "docker-build"),
		Entry("replaces invalid characters", "db:migrate", "exec", "db-migrate"),
		Entry("a name that is only the verb", "lint", "lint", ""),
	)
})

var _ = Describe("ShortenWsPath", func() {
	It("returns a path inside the workspace relative to its root", func() {
		ws := filepath.Join("home", "ws")
		Expect(fileparser.ShortenWsPath(ws, filepath.Join(ws, "sub", "dir"))).
			To(Equal(executable.Directory("//sub/dir")))
	})

	It("returns a path outside the workspace unchanged", func() {
		other := filepath.Join("home", "other")
		Expect(fileparser.ShortenWsPath(filepath.Join("home", "ws"), other)).
			To(Equal(executable.Directory(other)))
	})
})
