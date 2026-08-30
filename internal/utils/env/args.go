package env

import (
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/flowexec/flow/v2/types/executable"
)

func BuildArgsEnvMap(
	args executable.ArgumentList,
	execArgs []string,
	env map[string]string,
) (map[string]string, error) {
	al, err := resolveArgValues(args, execArgs, env, false)
	if err != nil {
		return nil, err
	}
	return argsToEnvMap(al, env), nil
}

// BuildChildArgsEnvMap is BuildArgsEnvMap with the precedence a child step needs: a value
// passed explicitly on the step wins over one inherited from the parent environment. At
// the top level the opposite holds, so that a --param override beats a positional arg.
func BuildChildArgsEnvMap(
	args executable.ArgumentList,
	execArgs []string,
	env map[string]string,
) (map[string]string, error) {
	al, err := resolveArgValues(args, execArgs, env, true)
	if err != nil {
		return nil, err
	}
	return argsToEnvMap(al, env), nil
}

func parseArgs(args executable.ArgumentList, execArgs []string) (flagArgs map[string]string, posArgs []string) {
	flagArgs = make(map[string]string)
	posArgs = make([]string, 0)
	knownFlags := args.Flags()
	for i := 0; i < len(execArgs); i++ {
		arg := execArgs[i]
		if !strings.HasPrefix(arg, "--") {
			posArgs = append(posArgs, arg)
			continue
		}

		// Strip the -- prefix
		flagStr := strings.TrimPrefix(arg, "--")

		// Handle --flag=value
		if name, value, ok := strings.Cut(flagStr, "="); ok {
			if slices.Contains(knownFlags, name) {
				flagArgs[name] = value
			}
			continue
		}

		// Handle --flag (no value)
		if !slices.Contains(knownFlags, flagStr) {
			continue
		}
		if args.FlagType(flagStr) == executable.ArgumentTypeBool {
			flagArgs[flagStr] = strconv.FormatBool(true)
		} else if i+1 < len(execArgs) && !strings.HasPrefix(execArgs[i+1], "--") {
			i++
			flagArgs[flagStr] = execArgs[i]
		}
	}
	return flagArgs, posArgs
}

func resolveArgValues(
	args executable.ArgumentList,
	execArgs []string,
	env map[string]string,
	preferInput bool,
) (executable.ArgumentList, error) {
	if len(args) == 0 {
		return nil, nil
	}
	flagArgs, posArgs := parseArgs(args, execArgs)
	if err := setArgValues(args, flagArgs, posArgs, env, preferInput); err != nil {
		return nil, err
	}
	return args, nil
}

func setArgValues(
	args executable.ArgumentList,
	flagArgs map[string]string,
	posArgs []string,
	env map[string]string,
	preferInput bool,
) error {
	fromEnv := func(arg executable.Argument) (string, bool) {
		if arg.EnvKey == "" {
			return "", false
		}
		val, found := env[arg.EnvKey]
		return val, found
	}
	fromInput := func(arg executable.Argument) (string, bool) {
		if arg.Flag != "" {
			val, ok := flagArgs[arg.Flag]
			return val, ok
		}
		if arg.Pos != nil && *arg.Pos != 0 && *arg.Pos <= len(posArgs) {
			return posArgs[*arg.Pos-1], true
		}
		return "", false
	}

	sources := []func(executable.Argument) (string, bool){fromEnv, fromInput}
	if preferInput {
		sources = []func(executable.Argument) (string, bool){fromInput, fromEnv}
	}
	for i, arg := range args {
		for _, source := range sources {
			if val, ok := source(arg); ok {
				arg.Set(val)
				args[i] = arg
				break
			}
		}
	}
	return args.ValidateValues()
}

func argsToEnvMap(args executable.ArgumentList, env map[string]string) map[string]string {
	envMap := make(map[string]string)
	for _, arg := range args {
		if arg.OutputFile != "" && arg.EnvKey == "" {
			continue
		}
		envMap[arg.EnvKey] = argValue(arg, env)
	}
	return envMap
}

// argValue returns the value to use for a resolved argument. A value that was actually
// supplied - on the command line, or inherited from the parent environment - is a
// literal. Only the declared default is authored in the flow file, so only it is
// expanded.
func argValue(arg executable.Argument, env map[string]string) string {
	if arg.IsSet() {
		return arg.Value()
	}
	return ExpandAuthored(arg.Default, env)
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
			if childArg.Pos != nil {
				pos := *childArg.Pos
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

	result := make([]string, len(argsWithPositions)+len(flagArgs))
	for i, arg := range argsWithPositions {
		result[i] = arg.value
	}
	pos := len(argsWithPositions)
	for flag, value := range flagArgs {
		result[pos] = "--" + flag + "=" + value
		pos++
	}

	return result
}
