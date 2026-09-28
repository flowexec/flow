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

// verbSynonyms maps words that aren't verbs, and that no pattern matches, to a verb.
var verbSynonyms = map[string]executable.Verb{
	"bench":     executable.VerbBenchmark,
	"conf":      executable.VerbConfigure,
	"del":       executable.VerbRemove,
	"deps":      executable.VerbInstall,
	"down":      executable.VerbStop,
	"ls":        executable.VerbList,
	"rm":        executable.VerbRemove,
	"typecheck": executable.VerbCheck,
	"up":        executable.VerbStart,
	"vendor":    executable.VerbInstall,
}

// nounVerbs are verbs more often used as nouns in a task name (`search-index`, `new-relic`), so
// they're only inferred when they are the whole name.
var nounVerbs = map[executable.Verb]bool{
	executable.VerbIndex:   true,
	executable.VerbLock:    true,
	executable.VerbNew:     true,
	executable.VerbOpen:    true,
	executable.VerbPackage: true,
	executable.VerbPlan:    true,
	executable.VerbProfile: true,
	executable.VerbQueue:   true,
	executable.VerbSet:     true,
	executable.VerbTag:     true,
	executable.VerbView:    true,
}

var hookPrefixes = []string{"pre", "post"}

var (
	nameSanitizer  = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	wordSeparators = regexp.MustCompile(`[\s:_.-]+`)
	camelBoundary  = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	wordPattern    = regexp.MustCompile(`[^\s:_.-]+`)
)

// InferVerb infers the most likely Executable verb from a script or makeTarget name. A name that
// is itself a verb is used as is; otherwise it is matched as a prefix, then word by word (so
// `docker-build` is a build). A name neither matches is split further and matched against every
// valid verb before falling back to exec.
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
	return inferVerbFromWords(name)
}

func inferVerbFromWords(name string) executable.Verb {
	split := strings.ToLower(camelBoundary.ReplaceAllString(name, "$1 $2"))
	for _, word := range wordSeparators.Split(split, -1) {
		if verb, ok := wordVerb(word); ok {
			return verb
		}
		for _, prefix := range hookPrefixes {
			if rest, found := strings.CutPrefix(word, prefix); found {
				if verb, ok := wordVerb(rest); ok {
					return verb
				}
			}
		}
	}
	return executable.VerbExec
}

func wordVerb(word string) (executable.Verb, bool) {
	if word == "" {
		return "", false
	}
	for _, vp := range verbPatterns {
		if vp.regex.FindString(word) == word {
			return vp.verb, true
		}
	}
	if verb, ok := verbSynonyms[word]; ok {
		return verb, true
	}
	if verb := executable.Verb(word); !nounVerbs[verb] && verb.Validate() == nil {
		return verb, true
	}
	return "", false
}

// NormalizeName strips any character that is not a letter, number, dash, or underscore,
// and also removes the verb from the name so the ref doesn't repeat it: as a prefix
// (`build-app` → `app`), or otherwise as a later word (`db-migrate` → `db`).
func NormalizeName(name, verb string) string {
	stripped := strings.TrimPrefix(name, verb)
	if stripped == name {
		stripped = removeVerbWord(name, verb)
	}
	stripped = strings.TrimPrefix(stripped, ":")
	stripped = strings.TrimPrefix(stripped, "-")
	stripped = strings.TrimPrefix(stripped, "_")

	return nameSanitizer.ReplaceAllString(stripped, "-")
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

func removeVerbWord(name, verb string) string {
	if verb == "" || verb == executable.VerbExec.String() {
		return name
	}
	lower := strings.ToLower(name)
	for _, vp := range verbPatterns {
		if vp.regex.MatchString(lower) {
			return name
		}
	}
	spans := wordSpans(name)
	if len(spans) < 2 {
		return name
	}
	for _, span := range spans[1:] {
		start, end := span[0], span[1]
		if !strings.EqualFold(name[start:end], verb) {
			continue
		}
		if wordSeparators.MatchString(name[start-1 : start]) {
			start--
		}
		return name[:start] + name[end:]
	}
	return name
}

func wordSpans(name string) [][2]int {
	var spans [][2]int
	for _, loc := range wordPattern.FindAllStringIndex(name, -1) {
		start := loc[0]
		for _, b := range camelBoundary.FindAllStringIndex(name[loc[0]:loc[1]], -1) {
			split := loc[0] + b[0] + 1
			spans = append(spans, [2]int{start, split})
			start = split
		}
		spans = append(spans, [2]int{start, loc[1]})
	}
	return spans
}
