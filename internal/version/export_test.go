package version

var (
	ReadBuildInfo = &readBuildInfo
	Resolve       = resolve
)

// SetBuildVars sets the values normally injected via ldflags.
func SetBuildVars(v, commit, date string) {
	version, gitCommit, buildDate = v, commit, date
}
