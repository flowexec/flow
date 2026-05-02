package fileparser

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/types/executable"
)

type composeFile struct {
	Services map[string]any `yaml:"services"`
}

var composeTags = []string{generatedTag, "docker-compose"}

// ExecutablesFromDockerCompose parses a docker-compose.yml and returns list of Executables for the services
func ExecutablesFromDockerCompose(wsPath, path string) (executable.ExecutableList, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to open docker compose file: %w", err)
	}
	defer f.Close()

	var cf composeFile
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cf); err != nil {
		return nil, fmt.Errorf("failed to decode docker compose file: %w", err)
	}

	execs := make(executable.ExecutableList, 0)
	dir := executable.Directory(shortenWsPath(wsPath, filepath.Dir(path)))
	// Per-service start/build
	for svc, data := range cf.Services {
		execs = append(execs, &executable.Executable{
			Name:        svc,
			Verb:        executable.VerbStart,
			Tags:        composeTags,
			Description: fmt.Sprintf("Start service %s via docker-compose", svc),
			Exec: &executable.ExecExecutableType{
				Dir: dir,
				Cmd: fmt.Sprintf("docker-compose up %s", svc),
			},
		})

		dataMap, ok := data.(map[string]any)
		if ok && dataMap["build"] != nil {
			execs = append(execs, &executable.Executable{
				Name:        svc,
				Verb:        executable.VerbBuild,
				Tags:        composeTags,
				Description: fmt.Sprintf("Build service %s via docker-compose", svc),
				Exec: &executable.ExecExecutableType{
					Dir: dir,
					Cmd: fmt.Sprintf("docker-compose build %s", svc),
				},
			})
		}
	}

	// start and stop all
	execs = append(execs, &executable.Executable{
		Verb:        executable.VerbStart,
		Aliases:     []string{"all", "services"},
		Tags:        composeTags,
		Description: "Start all services via docker-compose",
		Exec: &executable.ExecExecutableType{
			Dir: dir,
			Cmd: "docker-compose up",
		},
	})
	execs = append(execs, &executable.Executable{
		Verb:        executable.VerbStop,
		Aliases:     []string{"all", "services"},
		Tags:        composeTags,
		Description: "Stop all services via docker-compose",
		Exec: &executable.ExecExecutableType{
			Dir: dir,
			Cmd: "docker-compose down",
		},
	})

	return execs, nil
}
