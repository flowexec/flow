package run

import "sync"

// Test seams for the container backend. These expose unexported helpers and the
// PATH-lookup seam so the external _test package can assert argv construction and
// runtime detection without spawning real containers.

// BuildRunArgsForTest exposes buildRunArgs.
func BuildRunArgsForTest(spec ContainerSpec, envFile string) []string {
	return buildRunArgs(spec, envFile)
}

// WriteEnvFileForTest exposes writeEnvFile.
func WriteEnvFileForTest(env map[string]string) (string, error) {
	return writeEnvFile(env)
}

// SetLookPathForTest swaps the PATH-lookup seam and returns a restore func.
func SetLookPathForTest(fn func(string) (string, error)) func() {
	prev := lookPath
	lookPath = fn
	return func() { lookPath = prev }
}

// ResetRuntimeCacheForTest clears the memoized "auto" runtime resolution so each
// test observes a fresh detection.
func ResetRuntimeCacheForTest() {
	autoRuntimeOnce = sync.Once{}
	autoRuntime = ""
	errAutoRuntime = nil
}
