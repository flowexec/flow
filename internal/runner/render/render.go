package render

import (
	"encoding/json"
	"fmt"
	stdio "io"
	"os"
	"path/filepath"

	"github.com/flowexec/tuikit/views"
	"github.com/jahvon/expression"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/utils/env"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
)

type renderRunner struct{}

func NewRunner() runner.Runner {
	return &renderRunner{}
}

func (r *renderRunner) Name() string {
	return "render"
}

func (r *renderRunner) IsCompatible(executable *executable.Executable) bool {
	if executable == nil || executable.Render == nil {
		return false
	}
	return true
}

func (r *renderRunner) Exec(
	ctx *context.Context,
	e *executable.Executable,
	_ engine.Engine,
	inputEnv map[string]string,
	inputArgs []string,
) error {
	if !ctx.Config.ShowTUI() {
		return fmt.Errorf("unable to render when interactive mode is disabled")
	}

	renderSpec := e.Render
	if err := env.SetEnv(ctx.Config.CurrentVaultName(), e.Env(), inputArgs, inputEnv); err != nil {
		return errors.Wrap(err, "unable to set parameters to env")
	}

	if cb, err := env.CreateTempEnvFiles(
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

	envMap, err := env.BuildEnvMap(
		ctx.Config.CurrentVaultName(), e.Env(), inputArgs, inputEnv, env.DefaultEnv(ctx, e),
	)
	if err != nil {
		return errors.Wrap(err, "unable to set parameters to env")
	}
	targetDir, isTmp, err := renderSpec.Dir.ExpandDirectory(
		e.WorkspacePath(),
		e.FlowFilePath(),
		ctx.ProcessTmpDir,
		envMap,
	)
	if err != nil {
		return errors.Wrap(err, "unable to expand directory")
	} else if isTmp {
		ctx.ProcessTmpDir = targetDir
	}

	contentFile := filepath.Clean(filepath.Join(targetDir, renderSpec.TemplateFile))
	var templateData interface{}
	if renderSpec.TemplateDataFile != "" {
		templateData, err = readDataFile(targetDir, renderSpec.TemplateDataFile)
		if err != nil {
			return err
		}
	}

	tmpl := expression.NewTemplate(filepath.Base(renderSpec.TemplateFile), map[string]interface{}{"data": templateData})
	if err = tmpl.ParseFile(contentFile); err != nil {
		return errors.Wrapf(err, "unable to parse template file %s", contentFile)
	}

	data, err := tmpl.ExecuteToString()
	if err != nil {
		return errors.Wrapf(err, "unable to parse template file %s", contentFile)
	}

	logger.Log().Infof("Rendering content from file %s", contentFile)

	if os.Getenv(io.DisableInteractiveEnvKey) != "" {
		logger.Log().Print("### Rendered Content Start ###")
		logger.Log().Print(data)
		logger.Log().Print("### Rendered Content End ###")
		return nil
	}

	if err := ctx.TUIContainer.Start(); err != nil {
		return errors.Wrapf(err, "unable to open viewer")
	}
	defer func() {
		ctx.TUIContainer.WaitForExit()
	}()

	filename := filepath.Base(contentFile)
	ctx.TUIContainer.SetState("file", filename)
	return ctx.TUIContainer.SetView(views.NewMarkdownView(ctx.TUIContainer.RenderState(), data))
}

func readDataFile(dir, path string) (interface{}, error) {
	var templateData interface{}
	dataFilePath := filepath.Clean(filepath.Join(dir, path))
	if _, err := os.Stat(dataFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template data file %s does not exist", dataFilePath)
	}
	reader, err := os.Open(dataFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open template data file %s", dataFilePath)
	}
	defer reader.Close()
	data, err := stdio.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read template data file %s", dataFilePath)
	}
	extension := filepath.Ext(dataFilePath)
	switch extension {
	case ".json":
		if err = json.Unmarshal(data, &templateData); err != nil {
			return nil, errors.Wrapf(err, "unable to unmarshal template data file %s", dataFilePath)
		}
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(data, &templateData); err != nil {
			return nil, errors.Wrapf(err, "unable to unmarshal template data file %s", dataFilePath)
		}
	}
	return templateData, nil
}
