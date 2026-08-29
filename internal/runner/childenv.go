package runner

import (
	"maps"

	envUtils "github.com/flowexec/flow/v2/internal/utils/env"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
)

// ChildEnvAndArgs builds the environment and argument list for one child of a serial or
// parallel executable. The parent's resolved environment always reaches the child: it is
// the only path for a child that does not run a subprocess, since a request or render
// runner builds its own map and never reads the process environment.
//
// Arguments written on the step take precedence over inherited values. When the step
// declares none, the child's arguments are rebuilt from the parent environment by
// matching envKeys.
func ChildEnvAndArgs(
	parentEnv map[string]string,
	refArgs []string,
	child *executable.Executable,
) (map[string]string, []string) {
	childEnv := maps.Clone(parentEnv)
	if childEnv == nil {
		childEnv = make(map[string]string)
	}

	childArgs := make([]string, 0)
	execEnv := child.Env()
	if execEnv == nil || len(execEnv.Args) == 0 {
		if len(refArgs) > 0 {
			logger.Log().Warnf(
				"executable %s has no arguments defined, skipping argument processing",
				child.Ref().String(),
			)
		}
		return childEnv, childArgs
	}

	buildEnvMap := envUtils.BuildArgsEnvMap
	if len(refArgs) > 0 {
		for _, arg := range refArgs {
			childArgs = append(childArgs, envUtils.ExpandAuthored(arg, childEnv))
		}
		buildEnvMap = envUtils.BuildChildArgsEnvMap
	} else {
		childArgs = envUtils.BuildArgsFromEnv(execEnv.Args, childEnv)
	}

	argEnv, err := buildEnvMap(execEnv.Args, childArgs, childEnv)
	if err != nil {
		logger.Log().WrapError(err, "unable to process arguments")
	}
	maps.Copy(childEnv, argEnv)

	return childEnv, childArgs
}
