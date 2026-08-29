package builder

import (
	"github.com/flowexec/flow/v2/types/executable"
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

// InheritedEnvChildExec echoes the values it resolved, so a parent's propagation can be
// asserted from the command output. It declares no step args of its own.
func InheritedEnvChildExec(opts ...Option) *executable.Executable {
	pos := 1
	e := &executable.Executable{
		Verb:       "run",
		Name:       "inherited-env-child",
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Args: executable.ArgumentList{{EnvKey: "OUTER", Pos: &pos, Default: "(unset)"}},
			Cmd:  "echo \"child OUTER=[$OUTER] OUTERP=[$OUTERP]\"",
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

// SerialExecWithInheritedEnv refs a child without declaring step args, the shape in which
// a parent's args and params used to stop reaching the child entirely.
func SerialExecWithInheritedEnv(opts ...Option) *executable.Executable {
	pos := 1
	e := &executable.Executable{
		Verb:       "run",
		Name:       "serial-inherited-env",
		Visibility: privateExecVisibility(),
		Serial: &executable.SerialExecutableType{
			Args:   executable.ArgumentList{{EnvKey: "OUTER", Pos: &pos, Default: "outer-value"}},
			Params: executable.ParameterList{{EnvKey: "OUTERP", Text: "outer-param"}},
			Execs:  []executable.SerialRefConfig{{Ref: "run examples:inherited-env-child"}},
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}
