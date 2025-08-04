package env

import (
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/flowexec/flow/types/executable"
)

func BuildArgsEnvMap(
	args executable.ArgumentList,
	execArgs []string,
	env map[string]string,
) (map[string]string, error) {
	al, err := resolveArgValues(args, execArgs, env)
	if err != nil {
		return nil, err
	}
	return argsToEnvMap(al), nil
}

func parseArgs(args executable.ArgumentList, execArgs []string) (flagArgs map[string]string, posArgs []string) {
	flagArgs = make(map[string]string)
	posArgs = make([]string, 0)
	for i := 0; i < len(execArgs); i++ {
		split := strings.SplitN(execArgs[i], "=", 2)
		if len(split) == 2 && slices.Contains(args.Flags(), split[0]) {
			flagArgs[split[0]] = split[1]
			continue
		}
		posArgs = append(posArgs, execArgs[i])
	}
	return
}

func resolveArgValues(
	args executable.ArgumentList,
	execArgs []string,
	env map[string]string,
) (executable.ArgumentList, error) {
	if len(args) == 0 {
		return nil, nil
	}
	if env != nil {
		// Expand environment variables in arguments
		for i, a := range execArgs {
			execArgs[i] = os.Expand(a, func(key string) string {
				return env[key]
			})
		}
	}
	flagArgs, posArgs := parseArgs(args, execArgs)
	if err := setArgValues(args, flagArgs, posArgs, env); err != nil {
		return nil, err
	}
	return args, nil
}

func setArgValues(
	args executable.ArgumentList,
	flagArgs map[string]string,
	posArgs []string,
	env map[string]string,
) error {
	for i, arg := range args {
		if arg.EnvKey != "" {
			if val, found := env[arg.EnvKey]; found {
				// Use the input value if provided
				arg.Set(val)
				args[i] = arg
				continue
			}
		}

		if arg.Flag != "" {
			if val, ok := flagArgs[arg.Flag]; ok {
				arg.Set(val)
				args[i] = arg
			}
		} else if arg.Pos != nil && *arg.Pos != 0 {
			if *arg.Pos <= len(posArgs) {
				arg.Set(posArgs[*arg.Pos-1])
				args[i] = arg
			}
		}
	}
	return args.ValidateValues()
}

func argsToEnvMap(args executable.ArgumentList) map[string]string {
	envMap := make(map[string]string)
	for _, arg := range args {
		if arg.OutputFile != "" && arg.EnvKey == "" {
			continue
		}
		envMap[arg.EnvKey] = arg.Value()
	}
	return envMap
}

func filterArgsWithOutputFile(args executable.ArgumentList) executable.ArgumentList {
	var outputArgs executable.ArgumentList
	for _, arg := range args {
		if arg.OutputFile != "" {
			outputArgs = append(outputArgs, arg)
		}
	}

	return outputArgs
}

// BuildArgsFromEnv builds a list of arguments from the provided environment and expected args list. It will
// return the positional arguments in the order they are expected and then append any flag arguments at the end.
//
// TODO: Add support for overriding flag values.
func BuildArgsFromEnv(
	argsList executable.ArgumentList,
	inputEnv map[string]string,
) []string {
	if len(argsList) == 0 {
		return nil
	}

	type argWithPos struct {
		value string
		pos   int
	}
	var argsWithPositions []argWithPos
	flagArgs := make(map[string]string)

	for _, childArg := range argsList {
		if childArg.EnvKey == "" {
			continue
		}

		if value, found := inputEnv[childArg.EnvKey]; found {
			pos := 0
			if childArg.Pos != nil {
				pos = *childArg.Pos
				argsWithPositions = append(argsWithPositions, argWithPos{value: value, pos: pos})
			}
			if childArg.Flag != "" {
				flagArgs[childArg.Flag] = value
			}
		}
	}

	sort.Slice(argsWithPositions, func(i, j int) bool {
		return argsWithPositions[i].pos < argsWithPositions[j].pos
	})

	result := make([]string, len(argsWithPositions))
	for i, arg := range argsWithPositions {
		result[i] = arg.value
	}
	for flag, value := range flagArgs {
		result = append(result, flag+"="+value)
	}

	return result
}
