package executable

import "github.com/flowexec/flow/v2/types/executable"

// Test seams for the browse rendering helpers, so the external test package can
// assert what the library and detail views show without exporting them.

// ExecTypeNameForTest exposes execTypeName.
func ExecTypeNameForTest(e *executable.Executable) string {
	return execTypeName(e)
}

// ExecBodyMarkdownForTest exposes execBodyMarkdown.
func ExecBodyMarkdownForTest(e *executable.Executable) string {
	return execBodyMarkdown(e)
}
