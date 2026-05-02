package fileparser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/flow/v2/types/executable"
)

func ExecutablesFromPs1File(wsPath, filePath string) (*executable.Executable, error) {
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

	// PowerShell uses # for line comments, same as shell scripts.
	// Normalise Windows line endings so the line-based parser works correctly.
	content := strings.ReplaceAll(string(fileBytes), "\r\n", "\n")
	cfg, err := ExtractExecConfig(content, "# ")
	if err != nil {
		return nil, err
	}
	if err := ApplyExecConfig(exec, cfg); err != nil {
		return nil, err
	}

	exec.Tags = append(exec.Tags, generatedTag)
	return exec, nil
}
