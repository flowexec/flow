package common_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flowexec/flow/internal/io/common"
)

func TestShortenPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"single component", "foo", "foo"},
		{"two components", "foo/bar", "foo/bar"},
		{"shallow path under limit", "a/b/c/d/e", "a/b/c/d/e"},
		{"deep relative path", "a/b/c/d/e/f", "…/c/d/e/f"},
		{"deep absolute path", "/a/b/c/d/e/f", "…/c/d/e/f"},
		{"home substitution short", filepath.Join(home, "a/b/c"), "~/a/b/c"},
		{"home substitution deep", filepath.Join(home, "a/b/c/d/e"), "…/b/c/d/e"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := common.ShortenPath(tc.input); got != tc.want {
				t.Fatalf("ShortenPath(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
