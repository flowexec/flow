package mcp

import (
	"slices"
	"testing"
)

func TestExecutableArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"blank", "  \t\n ", nil},
		{"words", "one  two\tthree", []string{"one", "two", "three"}},
		{"flag", "--node workbox", []string{"--node", "workbox"}},
		{"leading separator", "-- --node workbox", []string{"--node", "workbox"}},
		{"bare separator", "--", []string{}},
		{"separator mid-string", "pos -- --node x", []string{"pos", "--node", "x"}},
		{"only the first separator goes", "-- a -- b", []string{"a", "--", "b"}},
		{"quoted separator is still a separator word", "'--' a", []string{"a"}},
		{"single quotes", "devbox 'uname -a'", []string{"devbox", "uname -a"}},
		{"double quotes", `devbox "uname -a"`, []string{"devbox", "uname -a"}},
		{"quotes inside a word", `--msg="hello world"`, []string{"--msg=hello world"}},
		{"adjacent quotes join", `'a b'"c d"e`, []string{"a bc de"}},
		{"empty quoted word kept", `a '' b`, []string{"a", "", "b"}},
		{"escaped space", `a\ b c`, []string{"a b", "c"}},
		{"escaped quote", `it\'s`, []string{"it's"}},
		{"escaped quote in double quotes", `"say \"hi\""`, []string{`say "hi"`}},
		{"backslash literal in single quotes", `'a\b'`, []string{`a\b`}},
		{"windows path", `C:\tools\bin`, []string{`C:\tools\bin`}},
		{"no expansion", `$HOME *`, []string{"$HOME", "*"}},
		{"unicode", "héllo 'wör ld'", []string{"héllo", "wör ld"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executableArgs(tt.input)
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecutableArgs_UnterminatedQuote(t *testing.T) {
	for _, input := range []string{`don't`, `"open`, `a 'b c`} {
		if got, err := executableArgs(input); err == nil {
			t.Errorf("executableArgs(%q) = %q, want an error", input, got)
		}
	}
}
