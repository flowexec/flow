package utils

import (
	"fmt"

	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/types/executable"
)

type refMatcher struct {
	ref executable.Ref
}

func (m *refMatcher) Matches(x any) bool {
	e, ok := x.(*executable.Executable)
	if !ok {
		return false
	}
	return e.Ref().Equals(m.ref)
}

func (m *refMatcher) String() string {
	return fmt.Sprintf("has ref %s", m.ref)
}

func ExecWithRef(ref executable.Ref) gomock.Matcher {
	return &refMatcher{ref}
}

type cmdMatcher struct {
	cmd string
}

func (m *cmdMatcher) Matches(x any) bool {
	e, ok := x.(*executable.Executable)
	if !ok {
		return false
	}
	if e.Exec == nil {
		return false
	}
	return e.Exec.Cmd == m.cmd
}

func (m *cmdMatcher) String() string {
	return fmt.Sprintf("has command %s", m.cmd)
}

func ExecWithCmd(cmd string) gomock.Matcher {
	return &cmdMatcher{cmd}
}
