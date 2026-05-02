package fileparser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/flow/v2/types/executable"
)

// batCommentPrefixes are the comment markers recognised in batch files for flow configuration.
var batCommentPrefixes = []string{"REM ", ":: "}

func ExecutablesFromBatFile(wsPath, filePath string) (*executable.Executable, error) {
	fn := filepath.Base(filePath)
	verb := InferVerb(fn)
	execName := NormalizeName(fn, verb.String())
	dir := executable.Directory(shortenWsPath(wsPath, filepath.Dir(filePath)))
	exec := &executable.Executable{
		Verb: verb,
		Name: execName,
		Exec: &executable.ExecExecutableType{
			Dir:  dir,
			File: filepath.Base(filePath),
		},
	}

	fileBytes, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}

	cfg, err := extractBatConfig(string(fileBytes))
	if err != nil {
		return nil, err
	}
	if err := ApplyExecConfig(exec, cfg); err != nil {
		return nil, err
	}

	exec.Tags = append(exec.Tags, generatedTag)
	return exec, nil
}

// extractBatConfig extracts flow configuration from batch file comments.
// It tries each known batch comment prefix and returns the first successful parse.
func extractBatConfig(data string) (*ParseResult, error) {
	// Normalise Windows line endings so the line-based parser works correctly.
	data = strings.ReplaceAll(data, "\r\n", "\n")

	for _, prefix := range batCommentPrefixes {
		result, err := ExtractExecConfig(data, prefix)
		if err != nil {
			return nil, err
		}
		if len(result.SimpleFields) > 0 || len(result.Params) > 0 || len(result.Args) > 0 {
			return result, nil
		}
	}
	// No config found with any prefix — return empty result.
	return &ParseResult{
		SimpleFields: make(map[string]string),
		Params:       make(executable.ParameterList, 0),
		Args:         make(executable.ArgumentList, 0),
	}, nil
}
