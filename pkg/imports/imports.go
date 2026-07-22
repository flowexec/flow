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
