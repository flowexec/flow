package builder

import (
	"github.com/flowexec/flow/types/executable"
)

func SerialExecByRefConfig(opts ...Option) *executable.Executable {
	name := "serial-config"
	e1 := SimpleExec(opts...)
	e2 := ExecWithArgs(opts...)
	e := &executable.Executable{
		Verb:       "start",
		Name:       name,
		Visibility: privateExecVisibility(),
		Serial: &executable.SerialExecutableType{
			Execs: []executable.SerialRefConfig{
				{Ref: e1.Ref()},
				{Ref: e2.Ref(), Args: []string{"hello", "x=123"}},
				{Cmd: "echo 'hello from serial command'"},
			},
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func SerialExecWithExit(opts ...Option) *executable.Executable {
	name := "serial-with-failure"
	e1 := SimpleExec(opts...)
	e2 := ExecWithExit(opts...)
	e3 := SimpleExec(opts...)
	ff := true
	e := &executable.Executable{
		Verb:       "start",
		Name:       name,
		Aliases:    []string{"serial-exit"},
		Visibility: privateExecVisibility(),
		Serial: &executable.SerialExecutableType{
			FailFast: &ff,
			Execs:    []executable.SerialRefConfig{{Ref: e1.Ref()}, {Ref: e2.Ref()}, {Ref: e3.Ref()}},
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}
