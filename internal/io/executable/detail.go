package executable

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/v2/internal/io/common"
	"github.com/flowexec/flow/v2/types/executable"
)

const mdArgsHeader = "   - **Arguments**\n"

func executableDetailOpts(exec *executable.Executable) views.DetailContentOpts {
	return views.DetailContentOpts{
		Title:    exec.Ref().String(),
		Subtitle: execTypeName(exec),
		Tags:     common.ColorizeTags(exec.Tags),
		Metadata: execMetadataFields(exec),
		Body:     execBodyMarkdown(exec),
		Footer:   fmt.Sprintf("_Located in %s_", common.ShortenPath(exec.FlowFilePath())),
		Entity:   exec,
	}
}

func execTypeName(exec *executable.Executable) string {
	switch {
	case exec.Exec != nil:
		return "Shell Executable"
	case exec.Launch != nil:
		return "Launch Executable"
	case exec.Request != nil:
		return "Request Executable"
	case exec.Render != nil:
		return "Render Executable"
	case exec.Serial != nil:
		return "Serial Executable"
	case exec.Parallel != nil:
		return "Parallel Executable"
	default:
		return "Executable"
	}
}

func execMetadataFields(exec *executable.Executable) []views.DetailField {
	const maxFields = 2

	// Collect all candidate fields in priority order:
	// 1. Visibility (always present)
	// 2. Aliases (if set)
	// 3. Verb Aliases (if set)
	// 4. Executed From / dir (if set)
	// 5. Timeout (if set)
	var candidates []views.DetailField

	visibility := "private"
	if exec.Visibility != nil {
		visibility = string(*exec.Visibility)
	}
	candidates = append(candidates, views.DetailField{Key: "Visibility", Value: visibility})

	if len(exec.Aliases) > 0 {
		candidates = append(candidates, views.DetailField{Key: "Aliases", Value: strings.Join(exec.Aliases, ", ")})
	}
	if len(exec.VerbAliases) > 0 {
		verbs := make([]string, len(exec.VerbAliases))
		for i, v := range exec.VerbAliases {
			verbs[i] = string(v)
		}
		candidates = append(candidates, views.DetailField{Key: "Verb Aliases", Value: strings.Join(verbs, ", ")})
	}
	if dir := execDir(exec); dir != "" {
		candidates = append(candidates, views.DetailField{Key: "Executed From", Value: dir})
	}
	if exec.Timeout != nil {
		candidates = append(candidates, views.DetailField{Key: "Timeout", Value: exec.Timeout.String()})
	}

	if len(candidates) > maxFields {
		candidates = candidates[:maxFields]
	}
	return candidates
}

// execDir returns the working directory for exec/render types, empty otherwise.
func execDir(exec *executable.Executable) string {
	switch {
	case exec.Exec != nil && exec.Exec.Dir != "":
		return string(exec.Exec.Dir)
	case exec.Render != nil && exec.Render.Dir != "":
		return string(exec.Render.Dir)
	default:
		return ""
	}
}

func execBodyMarkdown(exec *executable.Executable) string {
	var sections []string

	// Description
	if desc := execDescription(exec); desc != "" {
		sections = append(sections, desc)
	}

	// Type-specific configuration
	if cfg := execTypeConfig(exec); cfg != "" {
		sections = append(sections, cfg)
	}

	return strings.Join(sections, "\n\n")
}

func execDescription(exec *executable.Executable) string {
	var parts []string
	if d := strings.TrimSpace(exec.Description); d != "" {
		parts = append(parts, d)
	}
	return strings.Join(parts, "\n\n")
}

func execTypeConfig(spec *executable.Executable) string {
	switch {
	case spec.Exec != nil:
		return shellExecConfig(spec.Env(), spec.Exec)
	case spec.Launch != nil:
		return launchExecConfig(spec.Env(), spec.Launch)
	case spec.Request != nil:
		return requestExecConfig(spec.Env(), spec.Request)
	case spec.Render != nil:
		return renderExecConfig(spec.Env(), spec.Render)
	case spec.Serial != nil:
		return serialExecConfig(spec.Env(), spec.Serial)
	case spec.Parallel != nil:
		return parallelExecConfig(spec.Env(), spec.Parallel)
	default:
		return ""
	}
}

func shellExecConfig(e *executable.ExecutableEnvironment, s *executable.ExecExecutableType) string {
	if s == nil {
		return ""
	}
	md := "## Shell Configuration\n"
	if s.LogMode != "" {
		md += fmt.Sprintf("**Log Mode:** %s\n\n", s.LogMode)
	}
	if s.Cmd != "" {
		md += fmt.Sprintf("**Command**\n```sh\n%s\n```\n", s.Cmd)
	} else if s.File != "" {
		md += fmt.Sprintf("**File:** `%s`\n\n", s.File)
	}
	md += envTable(e)
	return md
}

func launchExecConfig(e *executable.ExecutableEnvironment, l *executable.LaunchExecutableType) string {
	if l == nil {
		return ""
	}
	md := "## Launch Configuration\n"
	if l.App != "" {
		md += fmt.Sprintf("**App:** `%s`\n\n", l.App)
	}
	if l.URI != "" {
		md += fmt.Sprintf("**URI:** [%s](%s)\n\n", l.URI, l.URI)
	}
	md += envTable(e)
	return md
}

func requestExecConfig(e *executable.ExecutableEnvironment, r *executable.RequestExecutableType) string {
	if r == nil {
		return ""
	}
	md := "## Request Configuration\n"
	md += fmt.Sprintf("**Method:** %s\n\n", r.Method)
	md += fmt.Sprintf("**URL:** [%s](%s)\n\n", r.URL, r.URL)

	if r.Timeout != 0 {
		md += fmt.Sprintf("**Request Timeout:** %s\n\n", r.Timeout)
	}
	if r.LogResponse {
		md += "**Log Response:** enabled\n\n"
	}
	if r.Body != "" {
		md += fmt.Sprintf("**Body:**\n```\n%s\n```\n", r.Body)
	}
	if len(r.Headers) > 0 {
		md += "\n**Headers**\n"
		for k, v := range r.Headers {
			md += fmt.Sprintf("- %s: %s\n", k, v)
		}
		md += "\n"
	}
	if len(r.ValidStatusCodes) > 0 {
		md += "**Accepted Status Codes**\n"
		for _, code := range r.ValidStatusCodes {
			md += fmt.Sprintf("- %d\n", code)
		}
		md += "\n"
	}
	if r.ResponseFile != nil {
		md += fmt.Sprintf("**Response Saved To:** %s\n\n", r.ResponseFile.Filename)
		if r.ResponseFile.SaveAs != "" {
			md += fmt.Sprintf("**Response Saved As:** %s\n\n", r.ResponseFile.SaveAs)
		}
	}
	if r.TransformResponse != "" {
		md += fmt.Sprintf("**Transformation Expression:**\n```\n%s\n```\n", r.TransformResponse)
	}
	md += envTable(e)
	return md
}

func renderExecConfig(e *executable.ExecutableEnvironment, r *executable.RenderExecutableType) string {
	if r == nil {
		return ""
	}
	md := "## Render Configuration\n"
	if r.TemplateFile != "" {
		md += fmt.Sprintf("**Template File:** `%s`\n\n", r.TemplateFile)
	}
	if r.TemplateDataFile != "" {
		md += fmt.Sprintf("**Template Store File:** `%s`\n\n", r.TemplateDataFile)
	}
	md += envTable(e)
	return md
}

func serialExecConfig(e *executable.ExecutableEnvironment, s *executable.SerialExecutableType) string {
	if s == nil {
		return ""
	}
	md := "## Serial Configuration\n"
	if s.FailFast != nil && *s.FailFast {
		md += "**Fail Fast:** enabled\n\n"
	} else if s.FailFast != nil && !*s.FailFast {
		md += "**Fail Fast:** disabled\n\n"
	}
	md += "**Executables**\n"
	for i, refCfg := range s.Execs {
		if refCfg.Ref != "" {
			md += fmt.Sprintf("%d. ref: %s\n", i+1, refCfg.Ref)
		} else if refCfg.Cmd != "" {
			md += fmt.Sprintf("%d. cmd:\n```sh\n%s\n```\n", i+1, refCfg.Cmd)
		}
		if refCfg.Retries > 0 {
			md += fmt.Sprintf("   - **Retries:** %d\n", refCfg.Retries)
		}
		if refCfg.ReviewRequired {
			md += fmt.Sprintf("   - **Review Required:** %v\n", refCfg.ReviewRequired)
		}
		if len(refCfg.Args) > 0 {
			md += mdArgsHeader
			for _, arg := range refCfg.Args {
				md += fmt.Sprintf("     - %s\n", arg)
			}
		}
	}
	md += envTable(e)
	return md
}

func parallelExecConfig(e *executable.ExecutableEnvironment, p *executable.ParallelExecutableType) string {
	if p == nil {
		return ""
	}
	md := "## Parallel Configuration\n"
	if p.MaxThreads > 0 {
		md += fmt.Sprintf("**Max Threads:** %d\n\n", p.MaxThreads)
	}
	if p.FailFast != nil && *p.FailFast {
		md += "**Fail Fast:** enabled\n\n"
	} else if p.FailFast != nil && !*p.FailFast {
		md += "**Fail Fast:** disabled\n\n"
	}
	md += "**Executables**\n"
	for i, refCfg := range p.Execs {
		if refCfg.Ref != "" {
			md += fmt.Sprintf("%d. ref: %s\n", i+1, refCfg.Ref)
		} else if refCfg.Cmd != "" {
			md += fmt.Sprintf("%d. cmd:\n```sh\n%s\n```\n", i+1, refCfg.Cmd)
		}
		if refCfg.Retries > 0 {
			md += fmt.Sprintf("   - **Retries:** %d\n", refCfg.Retries)
		}
		if len(refCfg.Args) > 0 {
			md += mdArgsHeader
			for _, arg := range refCfg.Args {
				md += fmt.Sprintf("     - %s\n", arg)
			}
		}
	}
	md += envTable(e)
	return md
}

func envTable(env *executable.ExecutableEnvironment) string {
	if env == nil {
		return ""
	}
	var table string
	if len(env.Params) > 0 {
		table += "\n### Parameters\n"
		table += "| Env Key | Type | Value |\n| --- | --- | --- |\n"
		for _, p := range env.Params {
			var valueType, valueInput string
			switch {
			case p.Text != "":
				valueType = "text"
				valueInput = p.Text
			case p.SecretRef != "":
				valueType = "secret"
				valueInput = p.SecretRef
			case p.Prompt != "":
				valueType = "prompt"
				valueInput = p.Prompt
			}
			table += fmt.Sprintf("| `%s` | %s | %s |\n", p.EnvKey, valueType, valueInput)
		}
	}

	if len(env.Args) > 0 {
		table += "\n### Arguments\n"
		table += "| Env Key | Arg Type | Input Type | Default | Required |\n| --- | --- | --- | --- | --- |\n"
		for _, a := range env.Args {
			var argType string
			switch {
			case a.Pos != nil && *a.Pos > 0:
				argType = "positional"
			case a.Flag != "":
				argType = "flag"
			}
			table += fmt.Sprintf(
				"| `%s` | %s | %s | %s | %t |\n",
				a.EnvKey, argType, a.Type, a.Default, a.Required,
			)
		}
	}
	return table
}

// templateDetailOpts builds detail view options for a template.
func templateDetailOpts(t *executable.Template) views.DetailContentOpts {
	return views.DetailContentOpts{
		Title:    t.Name(),
		Subtitle: "Template",
		Metadata: templateMetadataFields(t),
		Body:     templateBodyMarkdown(t),
		Footer:   fmt.Sprintf("_Located in %s_", common.ShortenPath(t.Location())),
		Entity:   t,
	}
}

func templateMetadataFields(t *executable.Template) []views.DetailField {
	const maxFields = 2
	var candidates []views.DetailField

	if len(t.Form) > 0 {
		candidates = append(candidates, views.DetailField{Key: "Form Fields", Value: fmt.Sprintf("%d", len(t.Form))})
	}
	if len(t.Artifacts) > 0 {
		candidates = append(candidates, views.DetailField{Key: "Artifacts", Value: fmt.Sprintf("%d", len(t.Artifacts))})
	}
	if len(t.PreRun) > 0 || len(t.PostRun) > 0 {
		steps := len(t.PreRun) + len(t.PostRun)
		candidates = append(candidates, views.DetailField{Key: "Run Steps", Value: fmt.Sprintf("%d", steps)})
	}

	if len(candidates) > maxFields {
		candidates = candidates[:maxFields]
	}
	return candidates
}

func templateBodyMarkdown(t *executable.Template) string {
	var sections []string

	if form := templateFormMarkdown(t); form != "" {
		sections = append(sections, form)
	}
	if artifacts := templateArtifactsMarkdown(t); artifacts != "" {
		sections = append(sections, artifacts)
	}
	if len(t.PreRun) > 0 {
		sections = append(sections, execRefsMarkdown("Pre-Run", t.PreRun))
	}
	if len(t.PostRun) > 0 {
		sections = append(sections, execRefsMarkdown("Post-Run", t.PostRun))
	}
	sections = append(sections, fmt.Sprintf("## Flow File\n```yaml\n%s\n```", t.Template))

	return strings.Join(sections, "\n\n")
}

func templateFormMarkdown(t *executable.Template) string {
	if len(t.Form) == 0 {
		return ""
	}
	md := "## Form Fields\n"
	md += "| Field | Prompt | Description | Default | Required |\n"
	md += "| --- | --- | --- | --- | --- |\n"
	for _, f := range t.Form {
		md += fmt.Sprintf("| %s | %s | %s | %s | %t |\n",
			f.Key, f.Prompt, f.Description, f.Default, f.Required)
	}
	return md
}

func templateArtifactsMarkdown(t *executable.Template) string {
	if len(t.Artifacts) == 0 {
		return ""
	}
	md := "## Artifacts\n"
	for _, a := range t.Artifacts {
		md += fmt.Sprintf("- Source: `%s`\n", filepath.Join(a.SrcDir, a.SrcName))
		if a.DstDir != "" {
			md += fmt.Sprintf("  Destination: `%s`\n", filepath.Join(a.DstDir, a.DstName))
		} else if a.DstName != "" {
			md += fmt.Sprintf("  Destination: `%s`\n", a.DstName)
		}
		if a.If != "" {
			md += fmt.Sprintf("  Conditional: `%s`\n", a.If)
		}
		md += fmt.Sprintf("  Rendered as template: %t\n", a.AsTemplate)
	}
	return md
}

func execRefsMarkdown(title string, refs []executable.TemplateRefConfig) string {
	md := fmt.Sprintf("## %s\n", title)
	for i, e := range refs {
		if e.Ref != "" {
			md += fmt.Sprintf("%d. ref: %s\n", i+1, e.Ref)
		} else if e.Cmd != "" {
			md += fmt.Sprintf("%d. cmd:\n```sh\n%s\n```\n", i+1, e.Cmd)
		}
		if len(e.Args) > 0 {
			md += mdArgsHeader
			for _, arg := range e.Args {
				md += fmt.Sprintf("     - %s\n", arg)
			}
		}
	}
	return md
}
