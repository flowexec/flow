package internal

import (
	"fmt"
	"maps"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/internal/services/git"
	"github.com/flowexec/flow/v2/internal/utils/env"
	"github.com/flowexec/flow/v2/pkg/context"
	flowErrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/types/executable"
)

// dryRunPlan is what a run would do, resolved without running anything: no command, no secret
// lookup, no prompt, no temp directory. That makes it safe for a workspace nobody has trusted.
type dryRunPlan struct {
	Ref       string            `yaml:"ref"`
	Workspace string            `yaml:"workspace"`
	Type      string            `yaml:"type"`
	Dir       string            `yaml:"dir,omitempty"`
	Cmd       string            `yaml:"cmd,omitempty"`
	File      string            `yaml:"file,omitempty"`
	Steps     []string          `yaml:"steps,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	// MissingPrompts lists prompt params that have no value; a real run would ask for them.
	MissingPrompts []string `yaml:"missingPrompts,omitempty"`
}

func printDryRun(
	ctx *context.Context, cmd *cobra.Command, e *executable.Executable, ref executable.Ref, execArgs []string,
) {
	supplied := make(map[string]string)
	if wsData, err := ctx.WorkspacesCache.GetWorkspaceConfigList(); err == nil {
		if wsCfg := wsData.FindByName(e.Workspace()); wsCfg != nil {
			applyWorkspaceParameterOverrides(wsCfg, supplied)
		}
	}
	applyParameterOverrides(flags.ValueFor[[]string](cmd, *flags.ParameterValueFlag, false), supplied)

	plan := dryRunPlan{Ref: ref.String(), Workspace: e.Workspace(), Env: map[string]string{}}
	maps.Copy(plan.Env, supplied)

	execEnv := e.Env()
	if execEnv == nil {
		execEnv = &executable.ExecutableEnvironment{}
	}
	for _, param := range execEnv.Params {
		if _, set := plan.Env[param.EnvKey]; set && param.EnvKey != "" {
			continue
		}
		switch {
		case param.Text != "":
			plan.Env[param.EnvKey] = param.Text
		case param.SecretRef != "":
			plan.Env[param.EnvKey] = fmt.Sprintf("<secret %s>", param.SecretRef)
		case param.Prompt != "":
			plan.MissingPrompts = append(plan.MissingPrompts, fmt.Sprintf("%s (%q)", param.EnvKey, param.Prompt))
			plan.Env[param.EnvKey] = "<prompt>"
		case param.EnvFile != "":
			loaded, _ := env.LoadEnvFromFiles(
				[]string{param.EnvFile}, filepath.Dir(e.FlowFilePath()),
			)
			for k, v := range loaded {
				if param.EnvKey == "" || param.EnvKey == k {
					plan.Env[k] = v
				}
			}
		}
	}

	argEnv, err := env.BuildArgsEnvMap(execEnv.Args, execArgs, plan.Env)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, flowErrors.NewUsageError("%v", err))
		return
	}
	maps.Copy(plan.Env, argEnv)

	resolveEnv := env.DefaultEnv(ctx, e)
	maps.Copy(resolveEnv, plan.Env)
	if checkout := resolveEnv[git.CheckoutEnvKey]; checkout != "" {
		plan.Env[git.CheckoutEnvKey] = checkout
	}
	describeExecutableType(ctx, cmd, e, resolveEnv, &plan)

	out, err := yaml.Marshal(plan)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}
	_, _ = ctx.StdOut().Write(out)
}

const dryRunTypeExec = "exec"

func describeExecutableType(
	ctx *context.Context, cmd *cobra.Command, e *executable.Executable, resolveEnv map[string]string, plan *dryRunPlan,
) {
	expandDir := func(dir executable.Directory) string {
		if dir == executable.TmpDirLabel {
			return "<temporary directory>"
		}
		resolved, _, err := dir.ExpandDirectory(e.WorkspacePath(), e.FlowFilePath(), "", resolveEnv)
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, flowErrors.NewUsageError("%v", err))
		}
		return resolved
	}

	switch {
	case e.Exec != nil:
		plan.Type = dryRunTypeExec
		plan.Dir = expandDir(e.Exec.Dir)
		plan.Cmd = e.Exec.Cmd
		plan.File = e.Exec.File
	case e.Serial != nil:
		plan.Type = "serial"
		plan.Dir = expandDir(e.Serial.Dir)
		for _, step := range e.Serial.Execs {
			plan.Steps = append(plan.Steps, describeStep(string(step.Ref), step.Cmd, step.Args))
		}
	case e.Parallel != nil:
		plan.Type = "parallel"
		plan.Dir = expandDir(e.Parallel.Dir)
		for _, step := range e.Parallel.Execs {
			plan.Steps = append(plan.Steps, describeStep(string(step.Ref), step.Cmd, step.Args))
		}
	case e.Request != nil:
		plan.Type = "request"
	case e.Launch != nil:
		plan.Type = "launch"
	case e.Render != nil:
		plan.Type = "render"
	}
}

func describeStep(ref, cmd string, args []string) string {
	if ref == "" {
		return "cmd: " + cmd
	}
	step := "ref: " + ref
	for _, arg := range args {
		step += " " + arg
	}
	return step
}
