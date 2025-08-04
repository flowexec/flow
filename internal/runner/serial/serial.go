package serial

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/services/expr"
	"github.com/flowexec/flow/internal/services/store"
	envUtils "github.com/flowexec/flow/internal/utils/env"
	execUtils "github.com/flowexec/flow/internal/utils/executables"
	"github.com/flowexec/flow/types/executable"
)

type serialRunner struct{}

func NewRunner() runner.Runner {
	return &serialRunner{}
}

func (r *serialRunner) Name() string {
	return "serial"
}

func (r *serialRunner) IsCompatible(executable *executable.Executable) bool {
	if executable == nil || executable.Serial == nil {
		return false
	}
	return true
}

func (r *serialRunner) Exec(
	ctx *context.Context,
	e *executable.Executable,
	eng engine.Engine,
	inputEnv map[string]string,
	inputArgs []string,
) error {
	serialSpec := e.Serial
	if err := envUtils.SetEnv(ctx.Config.CurrentVaultName(), e.Env(), inputArgs, inputEnv); err != nil {
		return errors.Wrap(err, "unable to set parameters to env")
	}

	if cb, err := envUtils.CreateTempEnvFiles(
		ctx.Config.CurrentVaultName(),
		e.FlowFilePath(),
		e.WorkspacePath(),
		e.Env(),
		inputArgs,
		inputEnv,
	); err != nil {
		ctx.AddCallback(cb)
		return errors.Wrap(err, "unable to create temporary env files")
	} else {
		ctx.AddCallback(cb)
	}

	if len(serialSpec.Execs) > 0 {
		str, err := store.NewStore(store.Path())
		if err != nil {
			return err
		}
		if err := str.CreateBucket(store.EnvironmentBucket()); err != nil {
			return err
		}
		cacheData, err := str.GetAll()
		if err != nil {
			return err
		}
		if err := str.Close(); err != nil {
			logger.Log().Error(err, "unable to close store")
		}

		return handleExec(ctx, e, eng, serialSpec, inputEnv, cacheData)
	}
	return fmt.Errorf("no serial executables to run")
}

func handleExec(
	ctx *context.Context,
	parent *executable.Executable,
	eng engine.Engine,
	serialSpec *executable.SerialExecutableType,
	inputEnv map[string]string,
	cacheData map[string]string,
) error {
	// Expand the directory of the serial execution. The root / parent's directory is used if one is not specified.
	var root *executable.Executable
	if ctx.RootExecutable != nil {
		root = ctx.RootExecutable
	} else {
		root = parent
	}
	if serialSpec.Dir == "" {
		serialSpec.Dir = executable.Directory(filepath.Dir(root.FlowFilePath()))
	}
	targetDir, isTmp, err := serialSpec.Dir.ExpandDirectory(
		root.WorkspacePath(),
		root.FlowFilePath(),
		ctx.ProcessTmpDir,
		inputEnv,
	)
	if err != nil {
		return errors.Wrap(err, "unable to expand directory")
	} else if isTmp && ctx.ProcessTmpDir == "" {
		ctx.ProcessTmpDir = targetDir
	}

	// Build the list of steps to execute
	var execs []engine.Exec
	conditionalData := expr.ExpressionEnv(ctx, parent, cacheData, inputEnv)

	for i, refConfig := range serialSpec.Execs {
		// Skip over steps that do not match the condition
		if refConfig.If != "" {
			truthy, err := expr.IsTruthy(refConfig.If, &conditionalData)
			if err != nil {
				return err
			}
			if !truthy {
				logger.Log().Debugf("skipping execution %d/%d", i+1, len(serialSpec.Execs))
				continue
			}
			logger.Log().Debugf("condition %s is true", refConfig.If)
		}

		// Get the executable for the step
		var exec *executable.Executable
		switch {
		case refConfig.Ref != "":
			var err error
			exec, err = execUtils.ExecutableForRef(ctx, refConfig.Ref)
			if err != nil {
				return err
			}
		case refConfig.Cmd != "":
			exec = execUtils.ExecutableForCmd(parent, refConfig.Cmd, i)
		default:
			return errors.New("serial executable must have a ref or cmd")
		}
		logger.Log().Debugf("executing %s (%d/%d)", exec.Ref(), i+1, len(serialSpec.Execs))

		// Prepare the environment and arguments for the child executable
		childEnv := make(map[string]string)
		childArgs := make([]string, 0)
		maps.Copy(childEnv, inputEnv)
		if len(refConfig.Args) > 0 {
			execEnv := exec.Env()
			if execEnv == nil || execEnv.Args == nil {
				logger.Log().Warnf(
					"executable %s has no arguments defined, skipping argument processing",
					exec.Ref().String(),
				)
			} else {
				for _, arg := range os.Environ() {
					kv := strings.SplitN(arg, "=", 2)
					if len(kv) == 2 {
						childEnv[kv[0]] = kv[1]
					}
				}

				if serialSpec.Args == nil {
					childArgs = refConfig.Args
				} else {
					childArgs = envUtils.BuildArgsFromEnv(execEnv.Args, childEnv)
					if len(childArgs) == 0 {
						childArgs = refConfig.Args // If no resolved args, fallback to original args
					}
				}

				a, err := envUtils.BuildArgsEnvMap(execEnv.Args, childArgs, childEnv)
				if err != nil {
					logger.Log().Error(err, "unable to process arguments")
				}
				maps.Copy(childEnv, a)
			}
		}

		// Set log fields and directory for the executable
		switch {
		case exec.Exec != nil:
			fields := map[string]interface{}{"step": exec.Ref().String()}
			exec.Exec.SetLogFields(fields)
			if serialSpec.Dir != "" && exec.Exec.Dir == "" {
				exec.Exec.Dir = serialSpec.Dir
			}
		case exec.Parallel != nil:
			if serialSpec.Dir != "" && exec.Parallel.Dir == "" {
				exec.Parallel.Dir = serialSpec.Dir
			}
		case exec.Serial != nil:
			if serialSpec.Dir != "" && exec.Serial.Dir == "" {
				exec.Serial.Dir = serialSpec.Dir
			}
		case exec.Request != nil:
			if exec.Request.ResponseFile != nil && serialSpec.Dir != "" && exec.Request.ResponseFile.Dir == "" {
				exec.Request.ResponseFile.Dir = serialSpec.Dir
			}
		case exec.Render != nil:
			if serialSpec.Dir != "" && exec.Render.Dir == "" {
				exec.Render.Dir = serialSpec.Dir
			}
		}

		runExec := func() error {
			return runSerialExecFunc(ctx, i, refConfig, exec, eng, childEnv, childArgs, serialSpec)
		}

		execs = append(execs, engine.Exec{ID: exec.Ref().String(), Function: runExec, MaxRetries: refConfig.Retries})
	}

	results := eng.Execute(ctx.Ctx, execs, engine.WithMode(engine.Serial), engine.WithFailFast(parent.Serial.FailFast))
	if results.HasErrors() {
		return errors.New(results.String())
	}
	return nil
}

func runSerialExecFunc(
	ctx *context.Context,
	step int,
	refConfig executable.SerialRefConfig,
	exec *executable.Executable,
	eng engine.Engine,
	childEnv map[string]string,
	childArgs []string,
	serialSpec *executable.SerialExecutableType,
) error {
	err := runner.Exec(ctx, exec, eng, childEnv, childArgs)
	if err != nil {
		return err
	}
	if step < len(serialSpec.Execs) && refConfig.ReviewRequired {
		logger.Log().Println("Do you want to proceed with the next execution? (y/n)")
		if !inputConfirmed(os.Stdin) {
			return fmt.Errorf("stopping runner early (%d/%d)", step+1, len(serialSpec.Execs))
		}
	}
	return nil
}

func inputConfirmed(in *os.File) bool {
	var response string
	_, _ = fmt.Fscanf(in, response)
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
