package fileparser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flowexec/flow/v2/types/executable"
)

type makeTarget struct {
	name        string
	description string
}

// e.g. "target: dep1 dep2"
var (
	targetLine = regexp.MustCompile(`^([a-zA-Z0-9_.-]+)\s*:(.*)$`)
	makeTags   = []string{generatedTag, "make"}
)

// ExecutablesFromMakefile parses a Makefile and returns a list of Executables for each makeTarget
func ExecutablesFromMakefile(wsPath, path string) (executable.ExecutableList, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to open Makefile: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	targets := make(map[string]*makeTarget)
	var lastComment string

	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)
		if trim == "" {
			lastComment = ""
			continue
		}
		if strings.HasPrefix(trim, "#") {
			lastComment = appendComment(lastComment, strings.TrimSpace(strings.TrimPrefix(trim, "#")))
			continue
		}
		if m := targetLine.FindStringSubmatch(line); m != nil {
			name := m[1]

			// Skip special targets and pattern rules
			// TODO: add support for these targets
			if strings.HasPrefix(name, ".") || strings.Contains(name, "%") {
				continue
			}

			targets[name] = &makeTarget{name: name, description: lastComment}
			lastComment = ""
		}
	}

	execs := make(executable.ExecutableList, 0, len(targets))
	dir := executable.Directory(shortenWsPath(wsPath, filepath.Dir(path)))

	for _, t := range targets {
		verb := InferVerb(t.name)
		execName := NormalizeName(t.name, verb.String())
		e := &executable.Executable{
			Name:        execName,
			Verb:        verb,
			Description: t.description,
			Tags:        makeTags,
			Exec: &executable.ExecExecutableType{
				Dir: dir,
				Cmd: fmt.Sprintf("make %s", t.name),
			},
		}

		cfg, err := ExtractExecConfig(t.description, "")
		if err != nil {
			return nil, err
		}

		if len(cfg.SimpleFields) > 0 || len(cfg.Params) > 0 || len(cfg.Args) > 0 {
			e.Description = ""
			if err := ApplyExecConfig(e, cfg); err != nil {
				return nil, err
			}
		}

		execs = append(execs, e)
	}
	return execs, nil
}

func appendComment(s string, comment string) string {
	if s == "" {
		return comment
	}
	if comment != "" {
		return s + "\n" + comment
	}
	return s
}
