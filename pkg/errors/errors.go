package errors

import (
	"fmt"
)

type ExecutableNotFoundError struct {
	Ref string
}

func (e ExecutableNotFoundError) Error() string {
	return fmt.Sprintf("%s executable not found", e.Ref)
}

func NewExecutableNotFoundError(ref string) ExecutableNotFoundError {
	return ExecutableNotFoundError{Ref: ref}
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

type CacheUpdateError struct {
	Err error
}

func (e CacheUpdateError) Error() string {
	return fmt.Sprintf("unable to update cache - %v", e.Err)
}

func (e CacheUpdateError) Unwrap() error {
	return e.Err
}

func NewCacheUpdateError(err error) CacheUpdateError {
	return CacheUpdateError{Err: err}
}
