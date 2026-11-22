package digest

import (
	"strings"
)

// Text renders the document as Markdown: a `## ` section per source, its metadata as a list,
// then each part of its content under a `### ` heading, with logs and files fenced.
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
		for _, p := range s.parts {
			if md := p.markdown("###"); md != "" {
				b.WriteString("\n" + md + "\n")
			}
		}
		blocks = append(blocks, b.String())
	}
	return strings.Join(blocks, "\n")
}
