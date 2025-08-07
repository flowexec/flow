package builder

import (
	"github.com/flowexec/flow/types/executable"
)

func ParallelExecByRefConfig(opts ...Option) *executable.Executable {
	name := "parallel-config"
	e1 := SimpleExec(opts...)
	e2 := ExecWithArgs(opts...)
	e := &executable.Executable{
		Verb:       "start",
		Name:       name,
		Visibility: privateExecVisibility(),
		Parallel: &executable.ParallelExecutableType{
			Execs: []executable.ParallelRefConfig{
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

func ParallelExecWithExit(opts ...Option) *executable.Executable {
	name := "parallel-with-failure"
	e1 := ExecWithPauses(opts...)
	e2 := ExecWithExit(opts...)
	e3 := ExecWithPauses(opts...)
	ff := true
	e := &executable.Executable{
		Verb:       "start",
		Name:       name,
		Aliases:    []string{"parallel-exit"},
		Visibility: privateExecVisibility(),
		Parallel: &executable.ParallelExecutableType{
			FailFast: &ff,
			Execs:    executable.ParallelRefConfigList{{Ref: e1.Ref()}, {Ref: e2.Ref()}, {Ref: e3.Ref()}},
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ParallelExecWithMaxThreads(opts ...Option) *executable.Executable {
	e := ParallelExecByRefConfig(opts...)
	e.Parallel.MaxThreads = 1
	return e
}
