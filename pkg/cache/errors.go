package cache

import (
	"fmt"
)

type ExecutableNotFoundError struct {
	Ref string
}

func (e ExecutableNotFoundError) Error() string {
	return fmt.Sprintf("unable to find executable with reference %s", e.Ref)
}

func (e ExecutableNotFoundError) Unwrap() error {
	return fmt.Errorf("executable not found")
}

func NewExecutableNotFoundError(ref string) ExecutableNotFoundError {
	return ExecutableNotFoundError{Ref: ref}
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
