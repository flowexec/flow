package env

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/filesystem"
	"github.com/flowexec/flow/internal/utils"
	"github.com/flowexec/flow/types/executable"
)

// SetEnv sets environment variables based on the parameters and arguments defined in the executable environment.
func SetEnv(
	currentVault string,
	exec *executable.ExecutableEnvironment,
	inputArgs []string,
	inputEnv map[string]string,
) error {
	var errs []error

	envMap := make(map[string]string)
	maps.Copy(envMap, inputEnv)

	for _, param := range exec.Params {
		if param.OutputFile != "" {
			// CreateTempEnvFiles will handle outputFile parameters
			continue
		}
		val, err := ResolveParameterValue(currentVault, param, envMap)
		if err != nil {
			errs = append(errs, err)
		}

		if err := os.Setenv(param.EnvKey, val); err != nil {
			errs = append(errs, fmt.Errorf("failed to set env %s: %w", param.EnvKey, err))
		}
		envMap[param.EnvKey] = val
	}

	argEnvMap, err := BuildArgsEnvMap(exec.Args, inputArgs, envMap)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to build inputArgs env map: %w", err))
	}
	for key, val := range argEnvMap {
		val = os.Expand(val, func(key string) string {
			if v, ok := envMap[key]; ok {
				return v
			}
			return ""
		})
		if err := os.Setenv(key, val); err != nil {
			errs = append(errs, fmt.Errorf("failed to set env %s: %w", key, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to set values for parameters: %v", errs)
	}
	return nil
}

// CreateTempEnvFiles creates temporary files for parameters and arguments that have an OutputFile defined.
// It returns a cleanup function that should be called to remove these files after use.
func CreateTempEnvFiles(
	currentVault, flowfilePath, wsPath string,
	exec *executable.ExecutableEnvironment,
	args []string,
	promptedEnv map[string]string,
) (func(ctx *context.Context) error, error) {
	var errs []error
	var tempFiles []string
	for _, param := range exec.Params {
		if param.OutputFile == "" {
			continue
		}
		val, err := ResolveParameterValue(currentVault, param, promptedEnv)
		if err != nil {
			errs = append(errs, err)
		}

		dest, err := createEnvValueFile(param.OutputFile, val, wsPath, flowfilePath, promptedEnv)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		tempFiles = append(tempFiles, dest)
	}

	al, err := resolveArgValues(exec.Args, args, promptedEnv)
	if err != nil {
		errs = append(errs, err)
	} else {
		filtered := filterArgsWithOutputFile(al)
		for _, arg := range filtered {
			dest, err := createEnvValueFile(arg.OutputFile, arg.Value(), wsPath, flowfilePath, promptedEnv)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			tempFiles = append(tempFiles, dest)
		}
	}

	cb := func(ctx *context.Context) error {
		for _, tempFile := range tempFiles {
			if err := os.Remove(tempFile); err != nil {
				return fmt.Errorf("failed to remove temp file %s: %w", tempFile, err)
			}
		}
		return nil
	}

	if len(errs) > 0 {
		return cb, fmt.Errorf("failed to create temp files for parameters: %v", errs)
	}
	return cb, nil
}

// BuildEnvMap constructs a map of environment variables based on the executable parameters and arguments.
func BuildEnvMap(
	currentVault string,
	exec *executable.ExecutableEnvironment,
	inputArgs []string,
	inputEnv map[string]string,
	defaultEnv map[string]string,
) (map[string]string, error) {
	var errs []error

	envMap := make(map[string]string)
	maps.Copy(envMap, inputEnv)

	for k, v := range defaultEnv {
		if _, ok := envMap[k]; !ok {
			envMap[k] = v
		}
	}

	for _, param := range exec.Params {
		if param.OutputFile != "" && param.EnvKey == "" {
			continue
		}

		val, err := ResolveParameterValue(currentVault, param, envMap)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		envMap[param.EnvKey] = val
	}

	argEnvMap, err := BuildArgsEnvMap(exec.Args, inputArgs, envMap)
	if err != nil {
		return nil, fmt.Errorf("failed to build inputArgs env map: %w", err)
	}
	for key, val := range argEnvMap {
		envMap[key] = os.Expand(val, func(key string) string {
			if v, ok := envMap[key]; ok {
				return v
			}
			return ""
		})
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get values for parameters: %v", errs)
	}
	return envMap, nil
}

// EnvMapToEnvList converts a map of environment variables to a slice of strings in the format "KEY=VALUE".
func EnvMapToEnvList(envMap map[string]string) []string {
	envList := make([]string, 0, len(envMap))
	for k, v := range envMap {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}
	return envList
}

// EnvListToEnvMap converts a slice of strings in the format "KEY=VALUE" to a map of environment variables.
func EnvListToEnvMap(envList []string) map[string]string {
	envMap := make(map[string]string)
	for _, env := range envList {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	return envMap
}

func DefaultEnv(ctx *context.Context, executable *executable.Executable) map[string]string {
	envMap := make(map[string]string)
	envMap["FLOW_RUNNER"] = "true"
	envMap["FLOW_CURRENT_WORKSPACE"] = ctx.CurrentWorkspace.AssignedName()
	envMap["FLOW_CURRENT_NAMESPACE"] = ctx.Config.CurrentNamespace
	if ctx.ProcessTmpDir != "" {
		envMap["FLOW_TMP_DIRECTORY"] = ctx.ProcessTmpDir
	}
	envMap["FLOW_EXECUTABLE_NAME"] = executable.Name
	envMap["FLOW_DEFINITION_PATH"] = executable.FlowFilePath()
	envMap["FLOW_DEFINITION_DIR"] = filepath.Dir(executable.FlowFilePath())
	envMap["FLOW_WORKSPACE_PATH"] = executable.WorkspacePath()
	envMap["FLOW_CONFIG_PATH"] = filesystem.ConfigDirPath()
	envMap["FLOW_CACHE_PATH"] = filesystem.CachedDataDirPath()
	envMap["DISABLE_FLOW_INTERACTIVE"] = "true"
	if interactive := os.Getenv("DISABLE_FLOW_INTERACTIVE"); interactive != "" {
		envMap["DISABLE_FLOW_INTERACTIVE"] = interactive
	}
	return envMap
}

func createEnvValueFile(destination, content, wsPath, flowFileDir string, envMap map[string]string) (string, error) {
	destDir := utils.ExpandDirectory(destination, wsPath, flowFileDir, envMap)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for temp file: %w", err)
	}

	filename := filepath.Base(destination)
	dest := filepath.Clean(filepath.Join(destDir, filename))
	if err := os.WriteFile(dest, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return dest, nil
}
