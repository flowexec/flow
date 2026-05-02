package fileparser

import (
	"os"
	"path/filepath"

	"github.com/flowexec/flow/v2/types/executable"
)

func ExecutablesFromShFile(wsPath, filePath string) (*executable.Executable, error) {
	fn := filepath.Base(filepath.Base(filePath)) // remove the ext and the path
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
	cfg, err := ExtractExecConfig(string(fileBytes), "# ")
	if err != nil {
		return nil, err
	}
	if err := ApplyExecConfig(exec, cfg); err != nil {
		return nil, err
	}

	exec.Tags = append(exec.Tags, generatedTag)
	return exec, nil
}
