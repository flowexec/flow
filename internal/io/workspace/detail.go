package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/types/workspace"
)

func workspaceDetailOpts(ws *workspace.Workspace) views.DetailContentOpts {
	title := ws.AssignedName()
	if ws.DisplayName != "" {
		title = ws.DisplayName
	}

	return views.DetailContentOpts{
		Title:    title,
		Subtitle: "Workspace",
		Tags:     common.ColorizeTags(ws.Tags),
		Metadata: workspaceMetadataFields(ws),
		Body:     workspaceBodyMarkdown(ws),
		Footer:   fmt.Sprintf("_Located in %s_", common.ShortenPath(ws.Location())),
		Entity:   ws,
	}
}

func workspaceMetadataFields(ws *workspace.Workspace) []views.DetailField {
	const maxFields = 2
	var candidates []views.DetailField

	// Show assigned name only when it differs from the display title
	if ws.DisplayName != "" && ws.DisplayName != ws.AssignedName() {
		candidates = append(candidates, views.DetailField{Key: "Name", Value: ws.AssignedName()})
	}

	if len(ws.EnvFiles) > 0 {
		candidates = append(candidates, views.DetailField{Key: "Env Files", Value: strings.Join(ws.EnvFiles, ", ")})
	}

	if len(candidates) > maxFields {
		candidates = candidates[:maxFields]
	}
	return candidates
}

func workspaceBodyMarkdown(ws *workspace.Workspace) string {
	var sections []string

	if desc := wsDescription(ws); desc != "" {
		sections = append(sections, desc)
	}

	if ws.VerbAliases != nil && len(*ws.VerbAliases) > 0 {
		md := "## Verb Aliases\n"
		for verb, mapped := range *ws.VerbAliases {
			md += fmt.Sprintf("- **%s** → %s\n", verb, strings.Join(mapped, ", "))
		}
		sections = append(sections, md)
	}

	if ws.Executables != nil {
		var lines []string
		if len(ws.Executables.Included) > 0 {
			lines = append(lines, "**Included**")
			for _, p := range ws.Executables.Included {
				lines = append(lines, fmt.Sprintf("- `%s`", p))
			}
		}
		if len(ws.Executables.Excluded) > 0 {
			lines = append(lines, "**Excluded**")
			for _, p := range ws.Executables.Excluded {
				lines = append(lines, fmt.Sprintf("- `%s`", p))
			}
		}
		if len(lines) > 0 {
			sections = append(sections, "## Executable Filters\n"+strings.Join(lines, "\n"))
		}
	}

	if len(sections) == 0 {
		return "*No description available.*"
	}
	return strings.Join(sections, "\n\n")
}

func wsDescription(ws *workspace.Workspace) string {
	var parts []string
	if d := strings.TrimSpace(ws.Description); d != "" {
		parts = append(parts, d)
	}
	if ws.DescriptionFile != "" {
		wsFile := filepath.Join(ws.Location(), ws.DescriptionFile)
		mdBytes, err := os.ReadFile(filepath.Clean(wsFile))
		if err != nil {
			parts = append(parts, fmt.Sprintf("**Error loading description file:** %s", err))
		} else if d := strings.TrimSpace(string(mdBytes)); d != "" {
			parts = append(parts, d)
		}
	}
	return strings.Join(parts, "\n\n")
}
