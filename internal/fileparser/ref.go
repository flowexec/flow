package fileparser

import (
	"regexp"
	"strings"

	"github.com/flowexec/flow/types/executable"
)

var verbPatterns = []struct {
	verb  executable.Verb
	regex *regexp.Regexp
}{
	{executable.VerbStart, regexp.MustCompile(`^(start|dev|serve|watch|run|preview|storybook)[\s:_-]?`)},
	{executable.VerbBuild, regexp.MustCompile(`^(build|compile|bundle|transpile)[\s:_-]?`)},
	{executable.VerbTest, regexp.MustCompile(`^(test|coverage|check|ci|e2e|unit)[\s:_-]?`)},
	{executable.VerbLint, regexp.MustCompile(`^(lint|format|fmt|prettier|eslint|stylelint)[\s:_-]?`)},
	{executable.VerbClean, regexp.MustCompile(`^(clean|reset|purge|clear)[\s:_-]?`)},
	{executable.VerbDeploy, regexp.MustCompile(`^(deploy|publish|release|push)[\s:_-]?`)},
	{executable.VerbInstall, regexp.MustCompile(`^(install|bootstrap|setup)[\s:_-]?`)},
	{executable.VerbRemove, regexp.MustCompile(`^(remove|uninstall|delete)[\s:_-]?`)},
	{executable.VerbUpdate, regexp.MustCompile(`^(update|upgrade)[\s:_-]?`)},
	{executable.VerbAnalyze, regexp.MustCompile(`^(analyze|audit|inspect|scan)[\s:_-]?`)},
	{executable.VerbConfigure, regexp.MustCompile(`^(configure|setup)[\s:_-]?`)},
	{executable.VerbGenerate, regexp.MustCompile(`^(generate|gen)[\s:_-]?`)},
}

// InferVerb infers the most likely Executable verb from a script or makeTarget name.
func InferVerb(name string) executable.Verb {
	lower := strings.ToLower(name)
	verb := executable.Verb(lower)
	if verb.Validate() == nil {
		return verb
	}
	for _, vp := range verbPatterns {
		if vp.regex.MatchString(lower) {
			return vp.verb
		}
	}
	// Substring match (lower priority)
	for _, vp := range verbPatterns {
		if vp.regex.FindStringIndex(lower) != nil {
			return vp.verb
		}
	}
	return executable.VerbExec
}

// NormalizeName strips any character that is not a letter, number, dash, or underscore,
// and also removes the verb prefix from the name if present.
func NormalizeName(name, verb string) string {
	name = strings.TrimPrefix(name, verb)
	name = strings.TrimPrefix(name, ":")
	name = strings.TrimPrefix(name, "-")
	name = strings.TrimPrefix(name, "_")

	return regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(name, "-")
}
