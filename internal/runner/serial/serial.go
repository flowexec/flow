package serial

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/tuikit/io"
	"github.com/jahvon/expression"
	"github.com/pkg/errors"

	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	envUtils "github.com/flowexec/flow/internal/utils/env"
	execUtils "github.com/flowexec/flow/internal/utils/executables"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/pkg/store"
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
		return handleExec(ctx, e, eng, serialSpec, inputEnv)
	}
	return fmt.Errorf("no serial executables to run")
}

func handleExec(
	ctx *context.Context,
	parent *executable.Executable,
	eng engine.Engine,
	serialSpec *executable.SerialExecutableType,
	inputEnv map[string]string,
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

	// Resolve all executables first to count duplicate refs
	type resolvedExec struct {
		exec *executable.Executable
		ref  string
	}
	resolved := make([]resolvedExec, 0, len(serialSpec.Execs))
	refCounts := make(map[string]int)
	for i, refConfig := range serialSpec.Execs {
		var exec *executable.Executable
		switch {
		case refConfig.Ref != "":
			var err error
			exec, err = execUtils.ExecutableForRef(ctx, parent, refConfig.Ref)
			if err != nil {
				return err
			}
		case refConfig.Cmd != "":
			exec = execUtils.ExecutableForCmd(parent, refConfig.Cmd, i)
		default:
			return errors.New("serial executable must have a ref or cmd")
		}
		ref := exec.Ref().String()
		refCounts[ref]++
		resolved = append(resolved, resolvedExec{exec: exec, ref: ref})
	}
	refIdx := make(map[string]int)

	// Build the list of steps to execute
	tracker := io.NewTaskTracker()
	var execs []engine.Exec

	for i, refConfig := range serialSpec.Execs {
		exec := resolved[i].exec

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
					logger.Log().WrapError(err, "unable to process arguments")
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

		ref := resolved[i].ref
		refIdx[ref]++
		taskName := ref
		if refConfig.Name != "" {
			taskName = refConfig.Name
		} else if refCounts[ref] > 1 {
			taskName = fmt.Sprintf("%s · %d", ref, refIdx[ref])
		}
		runExec := func() error {
			task := tracker.StartTask(taskName)
			ctx.CurrentTask = task
			err := runSerialExecFunc(ctx, i, refConfig, exec, eng, childEnv, childArgs, serialSpec)
			if err != nil {
				tracker.CompleteTask(task, io.TaskFailed, err)
				return err
			}
			tracker.CompleteTask(task, io.TaskSuccess, nil)
			return nil
		}

		// Create condition function if needed
		var conditionFunc func() (bool, error)
		if refConfig.If != "" {
			ifCondition := refConfig.If
			stepNum := i + 1
			totalSteps := len(serialSpec.Execs)
			conditionFunc = func() (bool, error) {
				cacheData, err := ctx.DataStore.GetAllProcessVars(store.EnvironmentBucket())
				if err != nil {
					return false, err
				}

				conditionalData := runner.ExpressionEnv(ctx, parent, cacheData, inputEnv)
				truthy, err := expression.IsTruthy(ifCondition, conditionalData)
				if err != nil {
					return false, err
				}
				if !truthy {
					tracker.StartTask(taskName).Status = io.TaskSkipped
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

	parentTask := ctx.CurrentTask
	if parentTask == nil {
		if tal, ok := logger.Log().(io.TaskAwareLogger); ok {
			tal.BeginGroup(parent.Ref().String())
		}
	}
	results := eng.Execute(ctx, execs, engine.WithMode(engine.Serial), engine.WithFailFast(parent.Serial.FailFast))
	ctx.CurrentTask = nil
	if parentTask != nil {
		parentTask.Children = append(parentTask.Children, tracker.Tasks()...)
	} else {
		if tal, ok := logger.Log().(io.TaskAwareLogger); ok {
			tal.EndGroup()
			tal.PrintTaskSummary(tracker.Tasks())
		}
	}
	if results.HasErrors() {
		return fmt.Errorf("serial execution failed")
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
		if !inputConfirmed(ctx.StdIn()) {
			return fmt.Errorf("stopping runner early (%d/%d)", step+1, len(serialSpec.Execs))
		}
	}
	return nil
}

func inputConfirmed(in *os.File) bool {
	reader := bufio.NewReader(in)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
