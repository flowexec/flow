package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

// maxListDescriptionRunes limits each description in the list tools' output.
const maxListDescriptionRunes = 160

var envKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// splitArgs splits an args string the way a POSIX shell would, without any expansion: whitespace
// separates words, single and double quotes group them, and a backslash escapes the next character.
//
// Outside single quotes a backslash only escapes whitespace, quotes, $, backticks, and itself;
// otherwise it stays literal so a Windows path like C:\tools survives unquoted.
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
	if r == '"' || r == '\\' || r == '$' || r == '`' {
		return true
	}
	return quote != '"' && (isArgSpace(r) || r == '\'')
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
		if !envKeyPattern.MatchString(key) {
			return nil, fmt.Errorf("%q is not a valid EnvKey", key)
		}
		flags = append(flags, "--param", key+"="+value)
	}
	return flags, nil
}

// executableArgs converts the execute tool's args string into the words handed to the executable.
// A leading `--` is dropped, since the caller of this function always adds its own. A later one
// stays, and ends flag parsing as it does on the command line.
func executableArgs(args string) ([]string, error) {
	words, err := splitArgs(args)
	if err != nil {
		return nil, err
	}
	if len(words) > 0 && words[0] == "--" {
		words = words[1:]
	}
	return words, nil
}

// validateExecutableRef rejects a verb or ID that flow's command line would read as one of its own
// flags, since both go ahead of the separator.
func validateExecutableRef(verb, id string) error {
	for _, value := range []string{verb, id} {
		if strings.HasPrefix(value, "-") {
			return fmt.Errorf("%q is not a valid executable verb or ID", value)
		}
	}
	return nil
}

// summarizeDescription returns the first line of a description, cut to maxListDescriptionRunes, for
// the list tools. The get tools return it in full.
func summarizeDescription(description string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(description), "\n")
	line = strings.TrimSpace(line)
	if runes := []rune(line); len(runes) > maxListDescriptionRunes {
		return string(runes[:maxListDescriptionRunes]) + "…"
	}
	return line
}

// dedupeExecutableDescription drops the description from get_executable's output when the rendered
// fullDescription already contains it, or an empty fullDescription. Output that isn't a JSON object
// passes through unchanged.
func dedupeExecutableDescription(output string) string {
	var fields map[string]any
	if err := json.Unmarshal([]byte(output), &fields); err != nil {
		return output
	}
	full, _ := fields["fullDescription"].(string)
	description, _ := fields["description"].(string)
	if strings.TrimSpace(full) == "" {
		delete(fields, "fullDescription")
	} else if strings.Contains(full, strings.TrimSpace(description)) {
		delete(fields, "description")
	}
	deduped, err := json.Marshal(fields)
	if err != nil {
		return output
	}
	return string(deduped)
}

// redactRemote strips credentials from a git remote URL before it reaches a client. SSH remotes and
// scp-like addresses only hold a login name, so they pass through.
func redactRemote(remote string) string {
	parsed, err := url.Parse(remote)
	if err != nil || parsed.User == nil || parsed.Scheme == "ssh" {
		return remote
	}
	parsed.User = nil
	return parsed.String()
}
