package mcp

import (
	"slices"
	"strings"
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
		{"separator mid-string stays", "pos -- --node x", []string{"pos", "--", "--node", "x"}},
		{"escaped dollar", `\$HOME "\$HOME" '\$HOME'`, []string{"$HOME", "$HOME", `\$HOME`}},
		{"escaped backtick", "\\`a\\`", []string{"`a`"}},
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

func TestParamOverrideArgs_RejectsInvalidKeys(t *testing.T) {
	for _, key := range []string{"", "BAD KEY", "A=B", "1A", "--flag"} {
		if _, err := paramOverrideArgs(map[string]any{key: "v"}); err == nil {
			t.Errorf("key %q accepted", key)
		}
	}
	if _, err := paramOverrideArgs(map[string]any{"GOOD_KEY1": "v"}); err != nil {
		t.Errorf("valid key rejected: %v", err)
	}
}

func TestValidateExecutableRef(t *testing.T) {
	for _, ref := range [][2]string{{"--help", ""}, {"run", "--background"}, {"run", "-x"}} {
		if err := validateExecutableRef(ref[0], ref[1]); err == nil {
			t.Errorf("%v accepted", ref)
		}
	}
	if err := validateExecutableRef("run", "ws/ns:name"); err != nil {
		t.Errorf("valid ref rejected: %v", err)
	}
}

func TestSummarizeDescription(t *testing.T) {
	long := strings.Repeat("é", maxListDescriptionRunes+5)
	tests := map[string]string{
		"":                  "",
		"short":             "short",
		"  first\nsecond\n": "first",
		long:                strings.Repeat("é", maxListDescriptionRunes) + "…",
	}
	for input, want := range tests {
		if got := summarizeDescription(input); got != want {
			t.Errorf("summarizeDescription(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDedupeExecutableDescription(t *testing.T) {
	tests := map[string]string{
		`not json`: `not json`,
		`{"description":"d","fullDescription":"# t\n\nd\n"}`: `{"fullDescription":"# t\n\nd\n"}`,
		`{"description":"d","fullDescription":"\n"}`:         `{"description":"d"}`,
		`{"description":"d","fullDescription":"other"}`:      `{"description":"d","fullDescription":"other"}`,
	}
	for input, want := range tests {
		if got := dedupeExecutableDescription(input); got != want {
			t.Errorf("dedupeExecutableDescription(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestRedactRemote(t *testing.T) {
	tests := map[string]string{
		"https://tok:s3cret@example.com/r.git": "https://example.com/r.git",
		"https://ghp_token@example.com/r.git":  "https://example.com/r.git",
		"https://example.com/r.git":            "https://example.com/r.git",
		"ssh://git@example.com/r.git":          "ssh://git@example.com/r.git",
		"git@example.com:org/r.git":            "git@example.com:org/r.git",
	}
	for input, want := range tests {
		if got := redactRemote(input); got != want {
			t.Errorf("redactRemote(%q) = %q, want %q", input, got, want)
		}
	}
}
