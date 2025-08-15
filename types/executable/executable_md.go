//nolint:goconst
package executable

import (
	"fmt"
	"path/filepath"
	"strings"
)

func execMarkdown(e *Executable) string {
	var mkdwn string
	mkdwn += fmt.Sprintf("# [Executable] %s\n", e.Ref())
	mkdwn += execDescriptionMarkdown(e, true)
	if e.Visibility != nil {
		mkdwn += fmt.Sprintf("**Visibility:** %s\n", *e.Visibility)
	}
	if e.Timeout != nil {
		mkdwn += fmt.Sprintf("**Timeout:** %s\n", e.Timeout.String())
	}
	if len(e.Aliases) > 0 {
		mkdwn += "**Aliases**\n"
		for _, alias := range e.Aliases {
			mkdwn += fmt.Sprintf("- `%s`\n", alias)
		}
		mkdwn += "\n"
	}

	if len(e.Tags) > 0 {
		mkdwn += "**Tags**\n"
		for _, tag := range e.Tags {
			mkdwn += fmt.Sprintf("- `%s`\n", tag)
		}
		mkdwn += "\n"
	}

	mkdwn += execTypeMarkdown(e)
	mkdwn += fmt.Sprintf("\n\n_Executable can be found in_ [%s](%s)\n", e.flowFilePath, e.flowFilePath)
	return mkdwn
}

func execDescriptionMarkdown(e *Executable, withPrefix bool) string {
	if e.Description == "" && e.inheritedDescription == "" {
		return ""
	}
	var mkdwn string

	prefix := ""
	if withPrefix {
		prefix = "│ "
	}
	if d := strings.TrimSpace(e.Description); d != "" {
		mkdwn += prefix + "\n"
		mkdwn += addPrefx(d, prefix)
	}
	if d := strings.TrimSpace(e.inheritedDescription); d != "" {
		mkdwn += prefix + "\n"
		mkdwn += addPrefx(d, prefix)
	}
	if mkdwn != "" {
		mkdwn += prefix + "\n"
	}
	mkdwn += "\n"
	return mkdwn
}

func execTypeMarkdown(spec *Executable) string {
	var mkdwn string
	switch {
	case spec == nil:
		mkdwn += "No executable type found\n"
	case spec.Exec != nil:
		mkdwn += shellExecMarkdown(spec.Env(), spec.Exec)
	case spec.Launch != nil:
		mkdwn += launchExecMarkdown(spec.Env(), spec.Launch)
	case spec.Request != nil:
		mkdwn += requestExecMarkdown(spec.Env(), spec.Request)
	case spec.Render != nil:
		mkdwn += renderExecMarkdown(spec.Env(), spec.Render)
	case spec.Serial != nil:
		mkdwn += serialExecMarkdown(spec.Env(), spec.Serial)
	case spec.Parallel != nil:
		mkdwn += parallelExecMarkdown(spec.Env(), spec.Parallel)
	default:
		mkdwn += "**generated markdown not supported for type**\n"
	}
	return mkdwn
}

func shellExecMarkdown(e *ExecutableEnvironment, s *ExecExecutableType) string {
	if s == nil {
		return ""
	}
	mkdwn := "## Shell Configuration\n"
	if s.Dir != "" {
		mkdwn += fmt.Sprintf("**Executed from:** `%s`\n", s.Dir)
	}
	if s.LogMode != "" {
		mkdwn += fmt.Sprintf("**Log Mode:** %s\n", s.LogMode)
	}
	if s.Cmd != "" {
		mkdwn += fmt.Sprintf("**Command**\n```sh\n%s\n```\n", s.Cmd)
	} else if s.File != "" {
		mkdwn += fmt.Sprintf("**File:** `%s`\n", s.File)
	}
	mkdwn += execEnvTable(e)

	return mkdwn
}

func launchExecMarkdown(e *ExecutableEnvironment, l *LaunchExecutableType) string {
	if l == nil {
		return ""
	}
	mkdwn := "## Launch Configuration\n"
	if l.App != "" {
		mkdwn += fmt.Sprintf("**App:** `%s`\n", l.App)
	}
	if l.URI != "" {
		mkdwn += fmt.Sprintf("**URI:** [%s](%s)\n", l.URI, l.URI)
	}
	mkdwn += execEnvTable(e)
	return mkdwn
}

func requestExecMarkdown(e *ExecutableEnvironment, r *RequestExecutableType) string {
	if r == nil {
		return ""
	}
	mkdwn := "## Request Configuration\n"
	mkdwn += fmt.Sprintf("**Method:** %s\n", r.Method)
	mkdwn += fmt.Sprintf("**URL:** [%s](%s)\n", r.URL, r.URL)

	if r.Timeout != 0 {
		mkdwn += fmt.Sprintf("**Request Timeout:** %s\n", r.Timeout)
	}
	if r.LogResponse {
		mkdwn += "**Log Response:** enabled\n"
	}
	if r.Body != "" {
		mkdwn += fmt.Sprintf("**Body:**\n```\n%s\n```\n", r.Body)
	}

	if len(r.Headers) > 0 {
		mkdwn += "\n**Headers**\n"
		for k, v := range r.Headers {
			mkdwn += fmt.Sprintf("- %s: %s\n", k, v)
		}
	}
	if len(r.ValidStatusCodes) > 0 {
		mkdwn += "**Accepted Status Codes**\n"
		for _, code := range r.ValidStatusCodes {
			mkdwn += fmt.Sprintf("- %d\n", code)
		}
	}

	if r.ResponseFile != nil {
		mkdwn += fmt.Sprintf("**Resonse Saved To:** %s\n", r.ResponseFile.Filename)
		if r.ResponseFile.SaveAs != "" {
			mkdwn += fmt.Sprintf("**Response Saved As:** %s\n", r.ResponseFile.SaveAs)
		}
	}
	if r.TransformResponse != "" {
		mkdwn += fmt.Sprintf("**Transformation Expression:**\n ```\n%s\n```\n", r.TransformResponse)
	}

	mkdwn += execEnvTable(e)
	return mkdwn
}

func renderExecMarkdown(e *ExecutableEnvironment, r *RenderExecutableType) string {
	if r == nil {
		return ""
	}

	mkdwn := "## Render Configuration\n"
	if r.Dir != "" {
		mkdwn += fmt.Sprintf("**Executed from:** `%s`\n", r.Dir)
	}
	if r.TemplateFile != "" {
		mkdwn += fmt.Sprintf("**Template File:** `%s`\n", r.TemplateFile)
	}
	if r.TemplateDataFile != "" {
		mkdwn += fmt.Sprintf("**Template Store File:** `%s`\n", r.TemplateDataFile)
	}
	mkdwn += execEnvTable(e)
	return mkdwn
}

func serialExecMarkdown(e *ExecutableEnvironment, s *SerialExecutableType) string {
	if s == nil {
		return ""
	}
	mkdwn := "## Serial Configuration\n"
	if s.FailFast != nil && *s.FailFast {
		mkdwn += "**Fail Fast:** enabled\n"
	} else if s.FailFast != nil && !*s.FailFast {
		mkdwn += "**Fail Fast:** disabled\n"
	}
	mkdwn += "**Executables**\n"
	for i, refCfg := range s.Execs {
		if refCfg.Ref != "" {
			mkdwn += fmt.Sprintf("%d. ref: %s\n", i+1, refCfg.Ref)
		} else if refCfg.Cmd != "" {
			mkdwn += fmt.Sprintf("%d. cmd: \n```sh\n%s\n```\n", i+1, refCfg.Cmd)
		}
		if refCfg.Retries > 0 {
			mkdwn += fmt.Sprintf("  - **Retries:** %d\n", refCfg.Retries)
		}
		if refCfg.ReviewRequired {
			mkdwn += fmt.Sprintf("  - **Review Required:** %v\n", refCfg.ReviewRequired)
		}
		if len(refCfg.Args) > 0 {
			mkdwn += "  - **Arguments**\n"
			for _, arg := range refCfg.Args {
				mkdwn += fmt.Sprintf("    - %s\n", arg)
			}
		}
	}
	mkdwn += execEnvTable(e)
	return mkdwn
}

func parallelExecMarkdown(e *ExecutableEnvironment, p *ParallelExecutableType) string {
	if p == nil {
		return ""
	}
	mkdwn := "## Parallel Configuration\n"
	if p.MaxThreads > 0 {
		mkdwn += fmt.Sprintf("**Max Threads:** %d\n", p.MaxThreads)
	}
	if p.FailFast != nil && *p.FailFast {
		mkdwn += "**Fail Fast:** enabled\n"
	} else if p.FailFast != nil && !*p.FailFast {
		mkdwn += "**Fail Fast:** disabled\n"
	}
	mkdwn += "**Executables**\n"
	for i, refCfg := range p.Execs {
		if refCfg.Ref != "" {
			mkdwn += fmt.Sprintf("%d. ref: %s\n", i+1, refCfg.Ref)
		} else if refCfg.Cmd != "" {
			mkdwn += fmt.Sprintf("%d. cmd: \n```sh\n%s\n```\n", i+1, refCfg.Cmd)
		}
		if refCfg.Retries > 0 {
			mkdwn += fmt.Sprintf("  - **Retries:** %d\n", refCfg.Retries)
		}
		if len(refCfg.Args) > 0 {
			mkdwn += "  - **Arguments**\n"
			for _, arg := range refCfg.Args {
				mkdwn += fmt.Sprintf("    - %s\n", arg)
			}
		}
	}
	mkdwn += execEnvTable(e)
	return mkdwn
}

func execEnvTable(env *ExecutableEnvironment) string {
	if env == nil {
		return ""
	}
	var table string
	if len(env.Params) > 0 {
		table += "### Parameters\n"
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
		table += "### Arguments\n"
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

func templateMarkdown(t *Template) string {
	mkdwn := fmt.Sprintf("# [Template] %s\n", t.Name())
	mkdwn += templateFormMarkdown(t)
	mkdwn += templateArtifactsMarkdown(t)
	if len(t.PreRun) > 0 {
		mkdwn += "## Pre-Run\n"
		for i, e := range t.PreRun {
			if e.Ref != "" {
				mkdwn += fmt.Sprintf("%d. ref: %s\n", i+1, e.Ref)
			} else if e.Cmd != "" {
				mkdwn += fmt.Sprintf("%d. cmd: \n```sh\n%s\n```\n", i+1, e.Cmd)
			}
			if len(e.Args) > 0 {
				mkdwn += "  - **Arguments**\n"
				for _, arg := range e.Args {
					mkdwn += fmt.Sprintf("    - %s\n", arg)
				}
			}
		}
	}
	if len(t.PostRun) > 0 {
		mkdwn += "## Post-Run\n"
		for i, e := range t.PostRun {
			if e.Ref != "" {
				mkdwn += fmt.Sprintf("%d. ref: %s\n", i+1, e.Ref)
			} else if e.Cmd != "" {
				mkdwn += fmt.Sprintf("%d. cmd: \n```sh\n%s\n```\n", i+1, e.Cmd)
			}
			if len(e.Args) > 0 {
				mkdwn += "  - **Arguments**\n"
				for _, arg := range e.Args {
					mkdwn += fmt.Sprintf("    - %s\n", arg)
				}
			}
		}
	}
	mkdwn += fmt.Sprintf("## Flow File\n```yaml\n%s\n```\n", t.Template)
	mkdwn += fmt.Sprintf("\n\n_Template can be found in_ [%s](%s)\n", t.Name(), t.Location())
	return mkdwn
}

func templateArtifactsMarkdown(t *Template) string {
	if len(t.Artifacts) == 0 {
		return ""
	}
	mkdwn := "## Artifacts\n"
	for _, a := range t.Artifacts {
		mkdwn += fmt.Sprintf("- Source: `%s`\n", filepath.Join(a.SrcDir, a.SrcName))
		if a.DstDir != "" {
			mkdwn += fmt.Sprintf("  Destination: `%s`\n", filepath.Join(a.DstDir, a.DstName))
		} else if a.DstName != "" {
			mkdwn += fmt.Sprintf("  Destination: `%s`\n", a.DstName)
		}
		if a.If != "" {
			mkdwn += fmt.Sprintf("  Conditional: `%s`\n", a.If)
		}
		mkdwn += fmt.Sprintf("  Rendered as a template: %t\n", a.AsTemplate)
	}
	return mkdwn
}

func templateFormMarkdown(t *Template) string {
	if len(t.Form) == 0 {
		return ""
	}
	mkdwn := "## Form Fields\n"
	mkdwn += "| Field | Prompt | Description | Default Value | Required |\n"
	for _, f := range t.Form {
		mkdwn += fmt.Sprintf(
			"| %s | %s | %s | %s | %t |\n",
			f.Key, f.Prompt, f.Description, f.Default, f.Required,
		)
	}
	return mkdwn
}

func addPrefx(s, prefix string) string {
	lines := strings.Split(s, "\n")
	var final string
	for _, line := range lines {
		final += prefix + line + "\n"
	}
	return final
}
