package parallel

import (
	stdCtx "context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/jahvon/expression"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/services/store"
	envUtils "github.com/flowexec/flow/internal/utils/env"
	execUtils "github.com/flowexec/flow/internal/utils/executables"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
)

type parallelRunner struct{}

func NewRunner() runner.Runner {
	return &parallelRunner{}
}

func (r *parallelRunner) Name() string {
	return "parallel"
}

func (r *parallelRunner) IsCompatible(executable *executable.Executable) bool {
	if executable == nil || executable.Parallel == nil {
		return false
	}
	return true
}

func (r *parallelRunner) Exec(
	ctx *context.Context,
	e *executable.Executable,
	eng engine.Engine,
	inputEnv map[string]string,
	inputArgs []string,
) error {
	parallelSpec := e.Parallel
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

	if len(parallelSpec.Execs) > 0 {
		return handleExec(ctx, e, eng, parallelSpec, inputEnv)
	}

	return fmt.Errorf("no parallel executables to run")
}

func handleExec(
	ctx *context.Context, parent *executable.Executable,
	eng engine.Engine,
	parallelSpec *executable.ParallelExecutableType,
	inputEnv map[string]string,
) error {
	groupCtx, cancel := stdCtx.WithCancel(ctx)
	defer cancel()
	group, _ := errgroup.WithContext(groupCtx)
	limit := parallelSpec.MaxThreads
	if limit == 0 {
		limit = len(parallelSpec.Execs)
	}
	group.SetLimit(limit)

	// Expand the directory of the parallel execution. The root / parent's directory is used if one is not specified.
	var root *executable.Executable
	if ctx.RootExecutable != nil {
		root = ctx.RootExecutable
	} else {
		root = parent
	}
	if parallelSpec.Dir == "" {
		parallelSpec.Dir = executable.Directory(filepath.Dir(root.FlowFilePath()))
	}
	targetDir, isTmp, err := parallelSpec.Dir.ExpandDirectory(
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

	for i, refConfig := range parallelSpec.Execs {
		// Get the executable for the step
		var exec *executable.Executable
		switch {
		case len(refConfig.Ref) > 0:
			var err error
			exec, err = execUtils.ExecutableForRef(ctx, parent, refConfig.Ref)
			if err != nil {
				return err
			}
		case refConfig.Cmd != "":
			exec = execUtils.ExecutableForCmd(parent, refConfig.Cmd, i)
		default:
			return errors.New("parallel executable must have a ref or cmd")
		}

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

				if parallelSpec.Args == nil {
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
			if parallelSpec.Dir != "" && exec.Exec.Dir == "" {
				exec.Exec.Dir = parallelSpec.Dir
			}
		case exec.Parallel != nil:
			if parallelSpec.Dir != "" && exec.Parallel.Dir == "" {
				exec.Parallel.Dir = parallelSpec.Dir
			}
		case exec.Serial != nil:
			if parallelSpec.Dir != "" && exec.Serial.Dir == "" {
				exec.Serial.Dir = parallelSpec.Dir
			}
		case exec.Request != nil:
			if exec.Request.ResponseFile != nil && parallelSpec.Dir != "" && exec.Request.ResponseFile.Dir == "" {
				exec.Request.ResponseFile.Dir = parallelSpec.Dir
			}
		case exec.Render != nil:
			if parallelSpec.Dir != "" && exec.Render.Dir == "" {
				exec.Render.Dir = parallelSpec.Dir
			}
		}

		runExec := func() error {
			err := runner.Exec(ctx, exec, eng, childEnv, childArgs)
			if err != nil {
				return err
			}
			return nil
		}

		// Create condition function if needed
		var conditionFunc func() (bool, error)
		if refConfig.If != "" {
			ifCondition := refConfig.If
			stepNum := i + 1
			totalSteps := len(parallelSpec.Execs)
			conditionFunc = func() (bool, error) {
				str, err := store.NewStore(store.Path())
				if err != nil {
					return false, err
				}
				if _, err := str.CreateAndSetBucket(store.EnvironmentBucket()); err != nil {
					_ = str.Close()
					return false, err
				}
				cacheData, err := str.GetAll()
				if err != nil {
					_ = str.Close()
					return false, err
				}
				if err := str.Close(); err != nil {
					logger.Log().Error(err, "unable to close store")
				}

				conditionalData := runner.ExpressionEnv(ctx, parent, cacheData, inputEnv)
				truthy, err := expression.IsTruthy(ifCondition, conditionalData)
				if err != nil {
					return false, err
				}
				if !truthy {
					logger.Log().Debugf("skipping execution %d/%d", stepNum, totalSteps)
				} else {
					logger.Log().Debugf("condition %s is true", ifCondition)
				}
				return truthy, nil
			}
		}

		execs = append(execs, engine.Exec{
			ID:         exec.Ref().String(),
			Function:   runExec,
			Condition:  conditionFunc,
			MaxRetries: refConfig.Retries,
		})
	}

	results := eng.Execute(
		ctx, execs,
		engine.WithMode(engine.Parallel),
		engine.WithFailFast(parent.Parallel.FailFast),
		engine.WithMaxThreads(parent.Parallel.MaxThreads),
	)
	if results.HasErrors() {
		return errors.New(results.String())
	}
	return nil
}
