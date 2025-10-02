package templates

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jahvon/expression"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/filesystem"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/utils"
	argUtils "github.com/flowexec/flow/internal/utils/env"
	execUtils "github.com/flowexec/flow/internal/utils/executables"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

func ProcessTemplate(
	ctx *context.Context,
	template *executable.Template,
	ws *workspace.Workspace,
	flowfileName, flowfileDir string,
) error {
	if flowfileName == "" {
		flowfileName = fmt.Sprintf("executables_%s", time.Now().Format("20060102150405"))
	}
	flowfileName = strings.ReplaceAll(strings.ToLower(flowfileName), " ", "_")
	if !executable.HasFlowFileExt(flowfileName) {
		flowfileName += executable.FlowFileExt
	}

	formMap := make(map[string]string)
	if template.Form != nil {
		if err := showForm(ctx, template.Form); err != nil {
			return err
		}
		formMap = template.Form.ValueMap()
	}

	env := os.Environ()
	envMap := make(map[string]string)
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		envMap[pair[0]] = pair[1]
	}
	flowfileDir = utils.ExpandDirectory(flowfileDir, ws.Location(), template.Location(), envMap)
	fullPath := filepath.Join(flowfileDir, flowfileName)
	logger.Log().Debugx(
		fmt.Sprintf("processing %s template", flowfileName),
		"template", template.Location(), "output", fullPath,
	)

	dataMap := newExpressionData(
		ws.AssignedName(), ws.Location(),
		flowfileName, flowfileDir, fullPath, template.Location(),
		envMap, formMap,
	)

	if err := runExecutables(
		ctx, ws, "pre-run", filepath.Dir(template.Location()), template.PreRun, dataMap,
	); err != nil {
		return err
	}
	if err := copyAllArtifacts(
		template.Artifacts,
		ws.Location(),
		filepath.Dir(template.Location()),
		flowfileDir,
		dataMap,
	); err != nil {
		return err
	}

	if template.Template != "" {
		flowfile, err := templateToFlowfile(template, dataMap)
		if err != nil {
			return err
		}

		if _, e := os.Stat(fullPath); e == nil {
			// TODO: Add a flag to overwrite existing files
			logger.Log().Warnx("Overwriting existing file", "dst", fullPath)
		}

		if err := filesystem.WriteFlowFile(fullPath, flowfile); err != nil {
			return errors.Wrap(err, fmt.Sprintf("unable to write flowfile %s from template", flowfileName))
		}
	}
	if err := runExecutables(ctx, ws, "post-run", flowfileDir, template.PostRun, dataMap); err != nil {
		return err
	}

	return nil
}

//nolint:gocognit
func runExecutables(
	ctx *context.Context,
	ws *workspace.Workspace,
	stage, flowfileDir string,
	execs []executable.TemplateRefConfig,
	templateData expressionData,
) error {
	logger.Log().Debugf("running %d %s executables", len(execs), stage)
	for i, e := range execs {
		if e.If != "" {
			eval, err := expression.IsTruthy(e.If, templateData)
			if err != nil {
				return errors.Wrap(err, "unable to evaluate if condition")
			}
			if !eval {
				logger.Log().Debugf("skipping %s executable %d", stage, i)
				return nil
			}
		}
		var exec *executable.Executable
		switch {
		case e.Ref != "":
			var err error
			ref, err := processAsGoTemplate(flowfileDir, string(e.Ref), templateData)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("unable to process %s executable %d", stage, i))
			}
			exec, err = execUtils.ExecutableForRef(ctx, nil, executable.Ref(ref.String()))
			if err != nil {
				return err
			}
		case e.Cmd != "":
			cmd, err := processAsGoTemplate(flowfileDir, e.Cmd, templateData)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("unable to process %s executable %d", stage, i))
			}
			exec = execUtils.ExecutableForCmd(templateParent(ws.AssignedName(), ws.Location(), flowfileDir), cmd.String(), i)
		default:
			return errors.New("post-run executable must have a ref or cmd")
		}
		inputEnv := make(map[string]string)
		inputArgs := make([]string, 0)
		ee := expressionEnv(templateData)
		maps.Copy(inputEnv, ee)
		//nolint:nestif
		if len(e.Args) > 0 {
			for _, arg := range e.Args {
				a, err := processAsGoTemplate(flowfileDir, arg, templateData)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("unable to process %s executable %d", stage, i))
				}
				inputArgs = append(inputArgs, a.String())
			}
			execEnv := exec.Env()
			if execEnv == nil || execEnv.Args == nil {
				logger.Log().Warnf(
					"executable %s has no arguments defined, skipping argument processing",
					exec.Ref().String(),
				)
			} else {
				a, err := argUtils.BuildArgsEnvMap(execEnv.Args, inputArgs, ee)
				if err != nil {
					logger.Log().Error(err, "unable to process arguments")
				}
				maps.Copy(inputEnv, a)
			}
		}
		if exec.Exec != nil {
			exec.Exec.SetLogFields(map[string]interface{}{
				"stage": stage,
				"step":  i + 1,
			})
		}
		if err := runner.Exec(ctx, exec, engine.NewExecEngine(), inputEnv, inputArgs); err != nil {
			return errors.Wrap(err, fmt.Sprintf("unable to execute %s executable %d", stage, i))
		}
	}
	return nil
}

func parseSourcePath(
	name, flowFileSrc, wsDir string,
	artifact executable.Artifact,
	templateData expressionData,
) (string, error) {
	var err error
	if artifact.SrcDir != "" {
		flowFileSrc = utils.ExpandDirectory(artifact.SrcDir, wsDir, flowFileSrc, expressionEnv(templateData))
	}
	var sb *bytes.Buffer
	sb, err = processAsGoTemplate(name, filepath.Join(flowFileSrc, artifact.SrcName), templateData)
	if err != nil {
		return "", errors.Wrap(err, "unable to process artifact as template")
	}
	return sb.String(), nil
}

func parseDestinationPath(
	name, dstDir, flowFileSrc, wsDir string,
	artifact executable.Artifact,
	templateData expressionData,
) (string, error) {
	var err error
	if artifact.DstDir != "" {
		dstDir = utils.ExpandDirectory(artifact.DstDir, wsDir, flowFileSrc, expressionEnv(templateData))
	}
	dstName := artifact.DstName
	var db *bytes.Buffer
	db, err = processAsGoTemplate(name, dstName, templateData)
	if err != nil {
		return "", errors.Wrap(err, "unable to process artifact as template")
	}
	dstName = db.String()
	return filepath.Join(dstDir, dstName), nil
}

func templateToFlowfile(
	t *executable.Template,
	templateData expressionData,
) (*executable.FlowFile, error) {
	buf, err := processAsGoTemplate(t.Name(), t.Template, templateData)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("flowfile template %s", t.Name()))
	}

	flowfile := &executable.FlowFile{}
	if err := yaml.NewDecoder(buf).Decode(flowfile); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to decode %s flowfile template", t.Name()))
	}

	return flowfile, nil
}

func processAsGoTemplate(fileName, txt string, data expressionData) (*bytes.Buffer, error) {
	tmpl := expression.NewTemplate(fileName, data)
	if err := tmpl.Parse(txt); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to parse %s template", fileName))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to execute %s template", fileName))
	}

	return &buf, nil
}

// templateParent returns a pseudo-executable that can be used as a parent for other executables. It simply includes
// the executable context that is derived from the rendered template.
func templateParent(ws, wsPath, flowfilePath string) *executable.Executable {
	e := &executable.Executable{}
	e.SetContext(ws, wsPath, "flow-internal", flowfilePath)
	return e
}
