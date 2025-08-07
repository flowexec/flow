//nolint:lll
package builder

import (
	"fmt"
	"strings"
	"time"

	tuikitIO "github.com/flowexec/tuikit/io"

	"github.com/flowexec/flow/types/executable"
)

func SimpleExec(opts ...Option) *executable.Executable {
	name := "simple-print"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Cmd: fmt.Sprintf("echo 'hello from %s'", name),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func NamelessExec(opts ...Option) *executable.Executable {
	e := &executable.Executable{
		Verb:       "run",
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Cmd: "echo 'hello from nameless'",
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ExecWithPauses(opts ...Option) *executable.Executable {
	name := "with-pauses"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Cmd: fmt.Sprintf(
				"echo 'hello from %[1]s'; sleep 1; echo 'hello from %[1]s'; sleep 1; echo 'hello from %[1]s'",
				name,
			),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ExecWithExit(opts ...Option) *executable.Executable {
	name := "with-exit"
	exitCode := 1
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Cmd: fmt.Sprintf("exit %d", exitCode),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ExecWithTimeout(opts ...Option) *executable.Executable {
	name := "with-timeout"
	timeout := 3 * time.Second
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Timeout:    &timeout,
		Exec: &executable.ExecExecutableType{
			Cmd: fmt.Sprintf("sleep %d", int(timeout.Seconds()+10)),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ExecWithTmpDir(opts ...Option) *executable.Executable {
	name := "with-tmp-dir"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Dir: executable.Directory(executable.TmpDirLabel),
			Cmd: fmt.Sprintf("echo 'hello from %[1]s';mkdir %[1]s; cd %[1]s; pwd", name),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ExecWithArgs(opts ...Option) *executable.Executable {
	name := "with-args"
	pos := 1
	args := executable.ArgumentList{{EnvKey: "ARG1", Pos: &pos}, {EnvKey: "ARG2", Flag: "x", Default: "yz"}}
	var argCmds []string
	for _, arg := range args {
		if arg.Pos != nil && *arg.Pos > 0 {
			argCmds = append(argCmds, fmt.Sprintf("echo 'pos=%d, key=%s'", arg.Pos, arg.EnvKey))
		} else if arg.Flag != "" {
			argCmds = append(argCmds, fmt.Sprintf("echo 'flag=%s, key=%s'", arg.Flag, arg.EnvKey))
		}
	}
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Args: args,
			Cmd:  fmt.Sprintf("echo 'hello from %s'; %s", name, strings.Join(argCmds, "; ")),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ExecWithParams(opts ...Option) *executable.Executable {
	name := "with-params"
	params := executable.ParameterList{
		{EnvKey: "PARAM1", Text: "value1"},
		{EnvKey: "PARAM2", SecretRef: "flow-example-secret"},
		{EnvKey: "PARAM3", Prompt: "Enter a value"},
	}
	var paramCmds []string
	for _, param := range params {
		switch {
		case param.Text != "":
			paramCmds = append(paramCmds, fmt.Sprintf(`echo "key=%s, value=$%[1]s"`, param.EnvKey))
		case param.SecretRef != "":
			paramCmds = append(
				paramCmds,
				fmt.Sprintf(`echo "key=%s, secret=%s, value=$%[1]s"`, param.EnvKey, param.SecretRef),
			)
		case param.Prompt != "":
			paramCmds = append(
				paramCmds,
				fmt.Sprintf(`echo "key=%s, prompt=%s value=$%[1]s"`, param.EnvKey, param.Prompt),
			)
		}
	}
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Params: params,
			Cmd:    fmt.Sprintf("echo 'hello from %s'; %s", name, strings.Join(paramCmds, "; ")),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ExecWithLogMode(opts ...Option) *executable.Executable {
	name := "with-plaintext"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			LogMode: tuikitIO.Text,
			Cmd: fmt.Sprintf(
				"echo 'hello from %s'; echo 'line 2'; echo 'line 3'; echo 'line 4'",
				name,
			),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func ExecWithEnvOutputFiles(opts ...Option) *executable.Executable {
	name := "with-file-param"
	params := executable.ParameterList{
		{Text: "database:\n  host: localhost\n  port: 5432", OutputFile: "test-config.yaml"},
		{Text: "#!/bin/bash\necho 'Hello from script'", OutputFile: "test-script.sh"},
	}
	pos := 1
	args := executable.ArgumentList{
		{Pos: &pos, Default: "notme", OutputFile: "output.txt"},
	}
	cmds := []string{
		`echo "param config file content:"; cat test-config.yaml; echo`,
		`echo "param script file content:"; cat test-script.sh; echo`,
		`echo "arg txt file content:"; cat output.txt; echo`,
	}
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Exec: &executable.ExecExecutableType{
			Params: params,
			Args:   args,
			Cmd:    fmt.Sprintf("echo 'hello from %s'; %s", name, strings.Join(cmds, "; ")),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}
