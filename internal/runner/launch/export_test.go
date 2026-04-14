package launch

// SetOpenFnForTest swaps the package-level open dispatcher for tests and
// returns a restore func to be called in teardown.
func SetOpenFnForTest(fn func(uri string) error) func() {
	prev := openFn
	openFn = fn
	return func() { openFn = prev }
}

// SetOpenWithFnForTest swaps the package-level open-with dispatcher for tests
// and returns a restore func to be called in teardown.
func SetOpenWithFnForTest(fn func(appName, uri string) error) func() {
	prev := openWithFn
	openWithFn = fn
	return func() { openWithFn = prev }
}
