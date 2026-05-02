package fileparser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flowexec/flow/v2/types/executable"
)

type packageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

var packageJSONTags = []string{generatedTag, "npm"}

// ExecutablesFromPackageJSON parses package.json scripts and returns a list of Executables for them
func ExecutablesFromPackageJSON(wsPath, path string) (executable.ExecutableList, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to open package.json: %w", err)
	}
	defer f.Close()

	var pkg packageJSON
	dec := json.NewDecoder(f)
	if err := dec.Decode(&pkg); err != nil {
		return nil, fmt.Errorf("failed to decode package.json: %w", err)
	}

	execs := make(executable.ExecutableList, 0)
	dir := executable.Directory(shortenWsPath(wsPath, filepath.Dir(path)))

	// default npm install
	execs = append(execs, &executable.Executable{
		Verb:        executable.VerbInstall,
		Aliases:     []string{"npm"},
		Description: "Install npm dependencies",
		Tags:        packageJSONTags,
		Exec: &executable.ExecExecutableType{
			Dir: dir,
			Cmd: "npm install",
		},
	})

	for name, scriptCmd := range pkg.Scripts {
		verb := InferVerb(name)
		execName := NormalizeName(name, verb.String())
		e := &executable.Executable{
			Verb:        verb,
			Name:        execName,
			Description: fmt.Sprintf("Run npm script %s:\n`%s`", name, scriptCmd),
			Tags:        packageJSONTags,
			Exec: &executable.ExecExecutableType{
				Dir: dir,
				Cmd: fmt.Sprintf("npm run %s", name),
			},
		}
		execs = append(execs, e)
	}
	return execs, nil
}
