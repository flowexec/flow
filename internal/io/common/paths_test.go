package common_test

import (
	"testing"

	"github.com/flowexec/flow/internal/io/common"
)

func TestShortenPath(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"single component", "foo", "foo"},
		{"two components", "foo/bar", "foo/bar"},
		{"three components", "a/b/c", "…/b/c"},
		{"deep unix path", "/Users/jahvon/workspaces/flow/internal/io", "…/internal/io"},
		{"trailing slash", "a/b/c/", "…/c/"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := common.ShortenPath(tc.input); got != tc.want {
				t.Fatalf("ShortenPath(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
