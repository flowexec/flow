package digest

import (
	"strings"
)

// Text renders the document as Markdown: a `## ` section per source, its metadata as a list,
// then its content.
func (d *Document) Text() string {
	blocks := make([]string, 0, len(d.Sections))
	for _, s := range d.Sections {
		var b strings.Builder
		b.WriteString("## " + s.Label + "\n")
		if len(s.fields) > 0 {
			b.WriteString("\n")
			for _, f := range s.fields {
				b.WriteString("- " + f.key + ": " + f.value + "\n")
			}
		}
		if s.Content != "" {
			b.WriteString("\n")
			if s.textHeading != "" {
				b.WriteString("### " + s.textHeading + "\n\n")
			}
			b.WriteString(s.Content + "\n")
		}
		if s.Truncated {
			b.WriteString("\n_(truncated; raise --max-chars or --max-lines for more)_\n")
		}
		blocks = append(blocks, b.String())
	}
	return strings.Join(blocks, "\n")
}
