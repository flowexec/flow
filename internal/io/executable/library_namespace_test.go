package executable_test

import (
	"slices"
	"testing"

	io "github.com/flowexec/flow/v2/internal/io/executable"
	"github.com/flowexec/flow/v2/types/executable"
)

func TestNamespaceRowsNestUnderParents(t *testing.T) {
	var execs executable.ExecutableList
	for _, ns := range []string{"", "api/v2", "api/v2", "api-x", "api/v2/beta", "api"} {
		e := &executable.Executable{Verb: "exec", Name: "e"}
		e.SetContext("ws", "/ws", ns, "/ws/f.flow")
		execs = append(execs, e)
	}

	got := io.NamespaceRowsForTest(execs, "ws")
	want := [][4]string{
		{"Root Namespace", "1", "", ""},
		{"api", "4", "api", "api/*"},
		{"  v2", "3", "api/v2", "api/v2/*"},
		{"    beta", "1", "api/v2/beta", "api/v2/beta"},
		{"api-x", "1", "api-x", "api-x"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("namespace rows:\n got %v\nwant %v", got, want)
	}
}
