package request

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/services/expr"
	"github.com/flowexec/flow/internal/services/rest"
	"github.com/flowexec/flow/internal/utils/env"
	"github.com/flowexec/flow/types/executable"
)

type requestRunner struct{}

func NewRunner() runner.Runner {
	return &requestRunner{}
}

func (r *requestRunner) Name() string {
	return "request"
}

func (r *requestRunner) IsCompatible(executable *executable.Executable) bool {
	if executable == nil || executable.Request == nil {
		return false
	}
	return true
}

func (r *requestRunner) Exec(
	ctx *context.Context,
	e *executable.Executable,
	_ engine.Engine,
	inputEnv map[string]string,
	inputArgs []string,
) error {
	requestSpec := e.Request
	envMap, err := env.BuildEnvMap(
		ctx.Config.CurrentVaultName(), e.Env(), inputArgs, inputEnv, env.DefaultEnv(ctx, e),
	)
	if err != nil {
		return errors.Wrap(err, "unable to set parameters to env")
	}

	url := expandEnvVars(envMap, requestSpec.URL)
	body := expandEnvVars(envMap, requestSpec.Body)
	for key, value := range requestSpec.Headers {
		requestSpec.Headers[key] = expandEnvVars(envMap, value)
	}
	restRequest := rest.Request{
		URL:     url,
		Method:  string(requestSpec.Method),
		Headers: requestSpec.Headers,
		Body:    body,
		Timeout: requestSpec.Timeout,
	}
	resp, err := rest.SendRequest(&restRequest, requestSpec.ValidStatusCodes)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}

	respStr := resp.Body
	if requestSpec.TransformResponse != "" {
		respStr, err = expr.EvaluateString(requestSpec.TransformResponse, resp)
		if err != nil {
			return errors.Wrap(err, "unable to transform response")
		}
	}

	log := logger.Log()
	if requestSpec.LogResponse {
		log.Infox(fmt.Sprintf("Successfully sent request to %s", requestSpec.URL), "response", respStr)
	} else {
		log.Infof("Successfully sent request to %s", requestSpec.URL)
	}

	if requestSpec.ResponseFile != nil && requestSpec.ResponseFile.Filename != "" {
		targetDir, isTmp, err := requestSpec.ResponseFile.Dir.ExpandDirectory(
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

		err = writeResponseToFile(
			respStr,
			filepath.Join(targetDir, requestSpec.ResponseFile.Filename),
			requestSpec.ResponseFile.SaveAs,
		)
		if err != nil {
			return errors.Wrap(err, "unable to save response")
		} else {
			log.Infof("Successfully saved response to %s", requestSpec.ResponseFile.Filename)
		}
	}

	return nil
}

func writeResponseToFile(resp, responseFile string, format executable.RequestResponseFileSaveAs) error {
	var formattedResp string
	var conversionErr error
	switch format {
	case "", executable.RequestResponseFileSaveAsRaw:
		formattedResp = resp
	case executable.RequestResponseFileSaveAsJson:
		var js interface{}
		if conversionErr = json.Unmarshal([]byte(resp), &js); conversionErr != nil {
			break
		}
		formattedResp = resp
	case executable.RequestResponseFileSaveAsIndentedJson, "formatted-json":
		var respMap interface{}
		conversionErr = json.Unmarshal([]byte(resp), &respMap)
		if conversionErr != nil {
			break
		}
		var formattedStr []byte
		formattedStr, conversionErr = json.MarshalIndent(respMap, "", "  ")
		if conversionErr != nil {
			break
		}
		formattedResp = string(formattedStr)
	case executable.RequestResponseFileSaveAsYaml, executable.RequestResponseFileSaveAsYml:
		var respMap interface{}
		conversionErr = json.Unmarshal([]byte(resp), &respMap)
		if conversionErr != nil {
			break
		}
		var yamlStr []byte
		yamlStr, conversionErr = yaml.Marshal(respMap)
		if conversionErr != nil {
			break
		}
		formattedResp = string(yamlStr)
	default:
		logger.Log().Warnf("unknown output format; skipping conversion")
		formattedResp = resp
	}

	if conversionErr != nil {
		logger.Log().Error(conversionErr, "unable to convert response")
	}

	file, err := os.Create(filepath.Clean(responseFile))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(formattedResp)
	if err != nil {
		return err
	}
	return nil
}

func expandEnvVars(envMap map[string]string, value string) string {
	if envMap == nil || value == "" {
		return value
	}
	return os.Expand(value, func(envVar string) string {
		return envMap[envVar]
	})
}
