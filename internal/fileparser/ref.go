package fileparser

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flowexec/flow/v2/types/executable"
)

var verbPatterns = []struct {
	verb  executable.Verb
	regex *regexp.Regexp
}{
	{executable.VerbStart, regexp.MustCompile(`^(start|dev|serve|watch|run|preview|storybook)[\s:_-]?`)},
	{executable.VerbBuild, regexp.MustCompile(`^(build|compile|bundle|transpile)[\s:_-]?`)},
	{executable.VerbTest, regexp.MustCompile(`^(test|coverage|check|ci|e2e|unit)[\s:_-]?`)},
	{executable.VerbLint, regexp.MustCompile(`^(lint|format|fmt|prettier|eslint|stylelint|vet)[\s:_-]?`)},
	{executable.VerbClean, regexp.MustCompile(`^(clean|reset|purge|clear)[\s:_-]?`)},
	{executable.VerbDeploy, regexp.MustCompile(`^(deploy|publish|release|push)[\s:_-]?`)},
	{executable.VerbInstall, regexp.MustCompile(`^(install|bootstrap|setup)[\s:_-]?`)},
	{executable.VerbRemove, regexp.MustCompile(`^(remove|uninstall|delete)[\s:_-]?`)},
	{executable.VerbUpdate, regexp.MustCompile(`^(update|upgrade)[\s:_-]?`)},
	{executable.VerbAnalyze, regexp.MustCompile(`^(analyze|audit|inspect|scan)[\s:_-]?`)},
	{executable.VerbConfigure, regexp.MustCompile(`^(configure|setup)[\s:_-]?`)},
	{executable.VerbGenerate, regexp.MustCompile(`^(generate|gen)[\s:_-]?`)},
}

var (
	nameSanitizer  = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	wordSeparators = regexp.MustCompile(`[\s:_.-]+`)
)

// InferVerb infers the most likely Executable verb from a script or makeTarget name. A name that
// is itself a verb is used as is; otherwise it is matched as a prefix, then word by word (so
// `docker-build` is a build), falling back to exec.
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
	for _, word := range wordSeparators.Split(lower, -1) {
		if word == "" {
			continue
		}
		for _, vp := range verbPatterns {
			if vp.regex.FindString(word) == word {
				return vp.verb
			}
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

	return nameSanitizer.ReplaceAllString(name, "-")
}

// ShortenWsPath returns path relative to the workspace root as a `//` directory. The relative
// part uses forward slashes on every platform, the form ExpandDirectory reads.
func ShortenWsPath(wsPath string, path string) executable.Directory {
	if strings.HasPrefix(path, wsPath) {
		rel := strings.TrimPrefix(path[len(wsPath):], string(filepath.Separator))
		return executable.Directory("//" + filepath.ToSlash(rel))
	}

	return executable.Directory(path)
}

// scriptName is the name a script's executable is derived from: its file name without the
// extension.
func scriptName(filePath string) string {
	fn := filepath.Base(filePath)
	return strings.TrimSuffix(fn, filepath.Ext(fn))
}
