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

func (e ExecutableNotFoundError) Code() string { return ErrCodeNotFound }

func NewExecutableNotFoundError(ref string) ExecutableNotFoundError {
	return ExecutableNotFoundError{Ref: ref}
}

type WorkspaceNotFoundError struct {
	Workspace string
}

func (e WorkspaceNotFoundError) Error() string {
	return fmt.Sprintf("workspace %s not found", e.Workspace)
}

func (e WorkspaceNotFoundError) Code() string { return ErrCodeNotFound }

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

func (e ExecutableContextError) Code() string { return ErrCodeInvalidInput }

type CacheUpdateError struct {
	Err error
}

func (e CacheUpdateError) Error() string {
	return fmt.Sprintf("unable to update cache - %v", e.Err)
}

func (e CacheUpdateError) Unwrap() error {
	return e.Err
}

func (e CacheUpdateError) Code() string { return ErrCodeInternal }

func NewCacheUpdateError(err error) CacheUpdateError {
	return CacheUpdateError{Err: err}
}

// UsageError indicates the user supplied invalid flags/arguments or attempted an
// unsupported combination.
type UsageError struct {
	Msg string
}

func (e UsageError) Error() string { return e.Msg }

func (e UsageError) Code() string { return ErrCodeInvalidInput }

func NewUsageError(format string, args ...any) UsageError {
	return UsageError{Msg: fmt.Sprintf(format, args...)}
}

// ValidationError indicates a value failed semantic or schema validation.
type ValidationError struct {
	Msg     string
	Details map[string]any
}

func (e ValidationError) Error() string { return e.Msg }

func (e ValidationError) Code() string { return ErrCodeValidationFailed }

func NewValidationError(msg string, details map[string]any) ValidationError {
	return ValidationError{Msg: msg, Details: details}
}
