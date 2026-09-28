package version_test

import (
	"runtime/debug"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/version"
)

var _ = Describe("Version", func() {
	var origReadBuildInfo func() (*debug.BuildInfo, bool)

	BeforeEach(func() {
		origReadBuildInfo = *version.ReadBuildInfo
	})

	AfterEach(func() {
		*version.ReadBuildInfo = origReadBuildInfo
		version.SetBuildVars("", "", "")
	})

	stubModuleVersion := func(v string) {
		*version.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Path: "github.com/flowexec/flow/v2", Version: v}}, true
		}
	}

	DescribeTable("falls back to the build info module version",
		func(moduleVersion, want string) {
			stubModuleVersion(moduleVersion)
			version.SetBuildVars("", "", "")
			version.Resolve()
			Expect(version.SemVer()).To(Equal(want))
		},
		Entry("release tag", "v2.2.2", "v2.2.2"),
		Entry("pre-release tag", "v2.3.0-rc.1", "v2.3.0-rc.1"),
		Entry("local build", "(devel)", ""),
		Entry("pseudo-version", "v2.2.3-0.20260901120000-abcdef123456", ""),
		Entry("dirty pseudo-version", "v2.2.3-0.20260901120000-abcdef123456+dirty", ""),
		Entry("empty", "", ""),
	)

	It("returns no version when build info is unavailable", func() {
		*version.ReadBuildInfo = func() (*debug.BuildInfo, bool) { return nil, false }
		version.SetBuildVars("", "", "")
		version.Resolve()
		Expect(version.SemVer()).To(BeEmpty())
		Expect(version.Short()).To(BeEmpty())
	})

	It("prefers the ldflags version over build info", func() {
		stubModuleVersion("v2.2.2")
		version.SetBuildVars("2.3.0", "abc123", "2026-09-01")
		version.Resolve()
		Expect(version.SemVer()).To(Equal("2.3.0"))
	})

	It("includes commit and build date when stamped", func() {
		version.SetBuildVars("2.3.0", "abc123", "2026-09-01")
		out := version.String()
		Expect(out).To(ContainSubstring("Version: 2.3.0"))
		Expect(out).To(ContainSubstring("Git Commit: abc123"))
		Expect(out).To(ContainSubstring("Build date: 2026-09-01"))
	})

	It("omits commit and build date when not stamped", func() {
		stubModuleVersion("v2.2.2")
		version.SetBuildVars("", "", "")
		version.Resolve()
		out := version.String()
		Expect(out).To(ContainSubstring("Version: v2.2.2"))
		Expect(out).NotTo(ContainSubstring("Git Commit"))
		Expect(out).NotTo(ContainSubstring("Build date"))
		Expect(out).To(ContainSubstring("Go version:"))
	})

	It("reports unknown when no version is available", func() {
		version.SetBuildVars("", "", "")
		Expect(version.String()).To(ContainSubstring("Version: unknown"))
	})
})
