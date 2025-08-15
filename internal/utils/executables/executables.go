package executables

import (
	"fmt"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
)

func ExecutableForRef(ctx *context.Context, ref executable.Ref) (*executable.Executable, error) {
	executableRef := context.ExpandRef(ctx, ref)
	exec, err := ctx.ExecutableCache.GetExecutableByRef(executableRef)
	if err != nil {
		return nil, err
	} else if exec == nil {
		return nil, fmt.Errorf("executable missing ref='%s'", ref)
	}

	if exec.Exec != nil {
		fields := map[string]interface{}{
			"executable": exec.Ref().String(),
		}
		exec.Exec.SetLogFields(fields)
	}

	return exec, nil
}

func ExecutableForCmd(parent *executable.Executable, cmd string, _ int) *executable.Executable {
	vis := executable.ExecutableVisibility(common.VisibilityInternal)
	exec := &executable.Executable{
		Verb:       parent.Verb,
		Name:       parent.Name,
		Visibility: &vis,
		Exec: &executable.ExecExecutableType{
			Cmd: cmd,
		},
	}
	fields := map[string]interface{}{"executable": exec.Ref().String()}
	exec.Exec.SetLogFields(fields)
	exec.SetContext(parent.Workspace(), parent.WorkspacePath(), parent.Namespace(), parent.FlowFilePath())
	return exec
}
