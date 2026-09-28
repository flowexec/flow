package imports

import (
	"github.com/flowexec/flow/v2/internal/fileparser"
	"github.com/flowexec/flow/v2/types/executable"
)

// ExecutablesFromImports generates the executables declared by flowFile.Imports.
func ExecutablesFromImports(
	wsName string, flowFile *executable.FlowFile,
) (executable.ExecutableList, error) {
	return fileparser.ExecutablesFromImports(wsName, flowFile)
}

// InferVerb infers the verb for an executable generated from a task, target, or script name, the
// way flow names the executables it imports. It falls back to exec.
func InferVerb(name string) executable.Verb {
	return fileparser.InferVerb(name)
}

// NormalizeName turns a task, target, or script name into a valid executable name, dropping a
// leading verb so the ref doesn't repeat it. A name that is only the verb normalizes to "".
func NormalizeName(name, verb string) string {
	return fileparser.NormalizeName(name, verb)
}

// ShortenWsPath returns dir relative to the workspace root as a `//` directory, or dir unchanged
// when it is outside the workspace.
func ShortenWsPath(wsPath, dir string) executable.Directory {
	return fileparser.ShortenWsPath(wsPath, dir)
}
