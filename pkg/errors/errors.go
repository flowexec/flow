package errors

import (
	"fmt"
)

type ExecutableNotFoundError struct {
	Verb string
	Name string
}

func (e ExecutableNotFoundError) Error() string {
	return fmt.Sprintf("%s executable %s not found", e.Verb, e.Name)
}

type WorkspaceNotFoundError struct {
	Workspace string
}

func (e WorkspaceNotFoundError) Error() string {
	return fmt.Sprintf("workspace %s not found", e.Workspace)
}

type ExecutableContextError struct {
	Workspace, Namespace, WorkspacePath, FlowFile string
}

func (e ExecutableContextError) Error() string {
	return fmt.Sprintf(
		"invalid context - %s/%s from (%s,%s)",
		e.Workspace,
		e.Namespace,
		e.WorkspacePath,
		e.FlowFile,
	)
}
