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

// NamespaceRowsForTest exposes namespaceChildren as (label, count, namespace, filter) tuples.
func NamespaceRowsForTest(execs executable.ExecutableList, wsName string) [][4]string {
	rows := namespaceChildren(execs, wsName, Filter{Namespace: executable.WildcardNamespace})
	out := make([][4]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, [4]string{r.Data[0], r.Data[1], r.Data[wsRowCellNsName], r.Data[wsRowCellNsFilt]})
	}
	return out
}
