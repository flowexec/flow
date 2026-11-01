package mcp

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// splitArgs splits an args string the way a POSIX shell would, without any expansion: whitespace
// separates words, single and double quotes group them, and a backslash escapes the next character.
//
// Outside single quotes a backslash only escapes whitespace, quotes, and itself; otherwise it stays
// literal so a Windows path like C:\tools survives unquoted.
func splitArgs(input string) ([]string, error) {
	var (
		words   []string
		current strings.Builder
		inWord  bool
		quote   rune
	)
	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\\' && i+1 < len(runes) && isEscapable(runes[i+1], quote):
			i++
			current.WriteRune(runes[i])
			inWord = true
		case quote == '"':
			if r == '"' {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			inWord = true
		case isArgSpace(r):
			if inWord {
				words = append(words, current.String())
				current.Reset()
				inWord = false
			}
		default:
			current.WriteRune(r)
			inWord = true
		}
	}
	if quote != 0 {
		return nil, errors.New("unterminated " + string(quote) + " quote")
	}
	if inWord {
		words = append(words, current.String())
	}
	return words, nil
}

func isEscapable(r, quote rune) bool {
	if quote == '"' {
		return r == '"' || r == '\\'
	}
	return isArgSpace(r) || r == '\'' || r == '"' || r == '\\'
}

func isArgSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// paramOverrideArgs converts the execute tool's params map into --param flags, sorted by key so the
// command line is deterministic.
func paramOverrideArgs(params any) ([]string, error) {
	if params == nil {
		return nil, nil
	}
	values, ok := params.(map[string]any)
	if !ok {
		return nil, errors.New("must be a map of EnvKey to string value")
	}
	flags := make([]string, 0, 2*len(values))
	for _, key := range slices.Sorted(maps.Keys(values)) {
		value, ok := values[key].(string)
		if !ok {
			return nil, fmt.Errorf("value for %q must be a string", key)
		}
		if key == "" || strings.Contains(key, "=") {
			return nil, fmt.Errorf("%q is not a valid EnvKey", key)
		}
		flags = append(flags, "--param", key+"="+value)
	}
	return flags, nil
}

// executableArgs converts the execute tool's args string into the words handed to the executable.
// One `--` is dropped if present, since the caller of this function always adds its own.
func executableArgs(args string) ([]string, error) {
	words, err := splitArgs(args)
	if err != nil {
		return nil, err
	}
	if i := slices.Index(words, "--"); i >= 0 {
		words = slices.Delete(words, i, i+1)
	}
	return words, nil
}
