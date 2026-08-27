package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Cobra's markdown is a flat dump: the command name is an h2 rather than a page
// title, every fenced block is untagged so nothing highlights, and the flag
// listings are fixed-width columns that overflow the site's content column and
// clip their own descriptions. polishCLIDocs rewrites the generated files into
// the same shape as the hand-written guides.
//
// It is deliberately conservative — anything it cannot parse is left as a code
// block rather than mangled into a broken table.

var (
	// Matches the leading `-x, --name` (or just `--name`) of a flag usage line.
	// Everything after it is the value placeholder and the description.
	flagHeadRe = regexp.MustCompile(`^\s+(?:-([a-zA-Z]), )?--([a-zA-Z0-9-]+)`)
	// Cobra aligns descriptions into a column, so two or more spaces separate
	// the placeholder from the description. A single space cannot be used: a
	// placeholder taken from backticks in the usage string may itself contain
	// one (e.g. `--spec flow logs`).
	columnGapRe = regexp.MustCompile(`\s{2,}`)
)

type flagDoc struct {
	short string
	name  string
	value string
	desc  string
}

func polishCLIDocs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		// index.md is hand-maintained, not emitted by GenMarkdownTree.
		if entry.Name() == "index.md" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		polished, ok := polishCLIDoc(string(raw))
		if !ok {
			continue
		}
		if err := os.WriteFile(path, []byte(polished), 0o600); err != nil {
			return err
		}
	}
	return nil
}

func polishCLIDoc(src string) (string, bool) {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")

	var (
		out     []string
		title   string
		section string
		i       int
	)

	for ; i < len(lines); i++ {
		if rest, ok := strings.CutPrefix(lines[i], "## "); ok {
			title = strings.TrimSpace(rest)
			i++
			break
		}
	}
	if title == "" {
		// Not a file this generator produced; leave it alone.
		return "", false
	}

	out = append(out, "---", "title: "+title)
	// Cobra puts the command's one-line summary directly under the heading; it
	// doubles as the page's meta description.
	if summary := leadParagraph(lines[i:]); summary != "" {
		out = append(out, "description: "+yamlString(summary))
	}
	out = append(out, "---", "", "# "+title)

	for ; i < len(lines); i++ {
		line := lines[i]

		switch {
		case strings.HasPrefix(line, "### "):
			section = strings.TrimSpace(line[len("### "):])
			out = append(out, "## "+sectionHeading(section))

		case strings.HasPrefix(strings.TrimSpace(line), "```"):
			var body []string
			for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "```"; i++ {
				body = append(body, lines[i])
			}
			out = append(out, renderBlock(section, body)...)

		case section == "SEE ALSO" && strings.HasPrefix(line, "* "):
			out = append(out, seeAlsoItem(line))

		default:
			out = append(out, line)
		}
	}

	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n", true
}

// leadParagraph returns the first line of prose before any heading or code
// block, which for a Cobra page is the command's short description.
func leadParagraph(lines []string) string {
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "#") || strings.HasPrefix(t, "```") {
			return ""
		}
		return t
	}
	return ""
}

func sectionHeading(section string) string {
	if section == "SEE ALSO" {
		return "See also"
	}
	return section
}

func isFlagSection(section string) bool {
	return strings.HasPrefix(section, "Options")
}

func renderBlock(section string, body []string) []string {
	if isFlagSection(section) {
		if flags, ok := parseFlagBlock(body); ok {
			return renderFlagTable(flags)
		}
	}
	return fence("shell", dedent(trimBlankEdges(body)))
}

func fence(lang string, body []string) []string {
	out := make([]string, 0, len(body)+2)
	out = append(out, "```"+lang)
	out = append(out, body...)
	return append(out, "```")
}

func renderFlagTable(flags []flagDoc) []string {
	out := []string{"| Flag | Type | Description |", "|------|------|-------------|"}
	for _, f := range flags {
		name := "--" + f.name
		if f.short != "" {
			name = "-" + f.short + ", " + name
		}
		value := ""
		if f.value != "" {
			value = "`" + f.value + "`"
		}
		out = append(out, "| `"+name+"` | "+value+" | "+escapeCell(f.desc)+" |")
	}
	return out
}

// parseFlagBlock turns Cobra's aligned flag usage listing into structured rows.
// It reports false if any line fails to parse, so the caller can fall back to
// rendering the block verbatim.
func parseFlagBlock(body []string) ([]flagDoc, bool) {
	var flags []flagDoc
	for _, raw := range body {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}

		m := flagHeadRe.FindStringSubmatchIndex(line)
		if m == nil {
			// A wrapped description continues the previous row.
			if len(flags) == 0 {
				return nil, false
			}
			flags[len(flags)-1].desc += " " + strings.TrimSpace(line)
			continue
		}

		rest := line[m[1]:]
		var value, desc string
		if gap := columnGapRe.FindStringIndex(rest); gap != nil {
			value = strings.TrimSpace(rest[:gap[0]])
			desc = strings.TrimSpace(rest[gap[1]:])
		} else {
			value = strings.TrimSpace(rest)
		}

		flags = append(flags, flagDoc{
			short: submatch(line, m, 1),
			name:  submatch(line, m, 2),
			value: value,
			desc:  desc,
		})
	}
	return flags, len(flags) > 0
}

func submatch(s string, m []int, group int) string {
	lo, hi := m[2*group], m[2*group+1]
	if lo < 0 {
		return ""
	}
	return s[lo:hi]
}

func escapeCell(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

// seeAlsoItem rewrites `* [flow](flow.md)\t - description` into a normal list
// item; the tab is a Cobra artifact and renders as a stray gap.
func seeAlsoItem(line string) string {
	s := strings.Join(strings.Fields(strings.TrimPrefix(line, "* ")), " ")
	return "- " + strings.Replace(s, ") - ", ") — ", 1)
}

func trimBlankEdges(body []string) []string {
	start, end := 0, len(body)
	for start < end && strings.TrimSpace(body[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(body[end-1]) == "" {
		end--
	}
	return body[start:end]
}

func dedent(body []string) []string {
	indent := -1
	for _, line := range body {
		if strings.TrimSpace(line) == "" {
			continue
		}
		n := len(line) - len(strings.TrimLeft(line, " "))
		if indent < 0 || n < indent {
			indent = n
		}
	}
	if indent <= 0 {
		return body
	}
	out := make([]string, len(body))
	for i, line := range body {
		if len(line) >= indent {
			out[i] = line[indent:]
		} else {
			out[i] = strings.TrimLeft(line, " ")
		}
	}
	return out
}
