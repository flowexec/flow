package fileparser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flowexec/tuikit/io"
	"github.com/pkg/errors"

	"github.com/flowexec/flow/types/executable"
)

const (
	TimeoutConfigurationKey     = "timeout"
	VerbConfigurationKey        = "verb"
	NameConfigurationKey        = "name"
	AliasConfigurationKey       = "alias"
	DescriptionConfigurationKey = "description"
	VisibilityConfigurationKey  = "visibility"
	TagConfigurationKey         = "tag"
	ParamConfigurationKey       = "param"
	ArgConfigurationKey         = "arg"
	DirConfigurationKey         = "dir"
	LogModeConfigurationKey     = "logmode"

	InternalListSeparator  = "|"
	InternalValueSeparator = ":"

	commentPrefix        = "# "
	multiLineKeyPrefix   = "f|"
	descriptionSeparator = "\n"
	descriptionAlias     = "desc"
)

var multiLineDescriptionTag = fmt.Sprintf("<%s%s>", multiLineKeyPrefix, DescriptionConfigurationKey)

// Regex to extract flow configuration fields from comment lines.
var flowConfigStartRegex = regexp.MustCompile(`f:(\w+)=`)

type ParseResult struct {
	SimpleFields map[string]string
	Params       executable.ParameterList
	Args         executable.ArgumentList
}

func ExtractExecConfig(data, prefix string) (*ParseResult, error) {
	result := &ParseResult{
		SimpleFields: make(map[string]string),
		Params:       make(executable.ParameterList, 0),
		Args:         make(executable.ArgumentList, 0),
	}
	processingMultiLineDescription := false
	for _, line := range strings.Split(data, "\n") {
		isComment := strings.HasPrefix(line, strings.TrimSpace(prefix))
		if trimmedLine := strings.TrimSpace(line); !isComment && trimmedLine != "" {
			// If the line is not a comment or empty, break out of the loop.
			// All flow executable configuration should be at the top of the file.
			break
		}

		line = strings.TrimPrefix(line, commentPrefix)
		if processingMultiLineDescription = parseMultiLineDescription(
			line, result.SimpleFields, processingMultiLineDescription,
		); processingMultiLineDescription {
			continue
		}

		lineResult, err := parseConfigurations(line)
		if err != nil {
			return nil, fmt.Errorf("unable to extract executable configurations: %w", err)
		}

		result.Params = append(result.Params, lineResult.Params...)
		result.Args = append(result.Args, lineResult.Args...)
		for key, value := range lineResult.SimpleFields {
			switch key {
			case DescriptionConfigurationKey:
				if existingValue, ok := result.SimpleFields[DescriptionConfigurationKey]; ok {
					result.SimpleFields[DescriptionConfigurationKey] =
						fmt.Sprintf("%s%s%s", existingValue, descriptionSeparator, value)
				} else {
					result.SimpleFields[DescriptionConfigurationKey] = value
				}
			case AliasConfigurationKey, TagConfigurationKey:
				value = strings.TrimSpace(value)
				if existingValue, ok := result.SimpleFields[key]; ok {
					result.SimpleFields[key] = fmt.Sprintf("%s%s%s", existingValue, InternalListSeparator, value)
				} else {
					result.SimpleFields[key] = value
				}
			default:
				result.SimpleFields[key] = strings.TrimSpace(value)
			}
		}
	}
	return result, nil
}

func ApplyExecConfig(exec *executable.Executable, result *ParseResult) error {
	if exec.Exec == nil {
		exec.Exec = &executable.ExecExecutableType{}
	}

	exec.Exec.Params = result.Params
	exec.Exec.Args = result.Args

	for key, value := range result.SimpleFields {
		switch key {
		case TimeoutConfigurationKey:
			dur, err := time.ParseDuration(value)
			if err != nil {
				return errors.Wrapf(err, "unable to parse timeout duration %s", value)
			}
			exec.Timeout = &dur
		case VerbConfigurationKey:
			exec.Verb = executable.Verb(value)
		case NameConfigurationKey:
			exec.Name = value
		case VisibilityConfigurationKey:
			v := executable.ExecutableVisibility(value)
			exec.Visibility = &v
		case DescriptionConfigurationKey:
			exec.Description = value
		case AliasConfigurationKey:
			if value != "" {
				exec.Aliases = strings.Split(value, InternalListSeparator)
			}
		case TagConfigurationKey:
			if value != "" {
				exec.Tags = strings.Split(value, InternalListSeparator)
			}
		case DirConfigurationKey:
			exec.Exec.Dir = executable.Directory(value)
		case LogModeConfigurationKey:
			exec.Exec.LogMode = io.LogMode(value)
		}
	}

	return nil
}

func parseConfigurations(line string) (*ParseResult, error) {
	result := &ParseResult{
		SimpleFields: make(map[string]string),
		Params:       make(executable.ParameterList, 0),
		Args:         make(executable.ArgumentList, 0),
	}

	matches := flowConfigStartRegex.FindAllStringSubmatchIndex(line, -1)
	if len(matches) == 0 {
		return result, nil
	}

	for i, match := range matches {
		key := strings.TrimSpace(strings.ToLower(line[match[2]:match[3]]))

		valueStart := match[1]
		var valueEnd int
		if i+1 < len(matches) {
			valueEnd = matches[i+1][0] // Start of next f:key=
		} else {
			valueEnd = len(line)
		}

		rawValue := strings.TrimSpace(line[valueStart:valueEnd])
		if rawValue == "" {
			continue
		}

		value := cleanValue(rawValue)
		normalizedKey := normalizeKey(key)

		if !validateKey(normalizedKey) {
			return nil, fmt.Errorf("invalid key (%s)", key)
		}

		switch normalizedKey {
		case ParamConfigurationKey:
			params, err := parseParams(value)
			if err != nil {
				return nil, fmt.Errorf("error parsing params in line '%s': %w", line, err)
			}
			result.Params = append(result.Params, params...)

		case ArgConfigurationKey:
			args, err := parseArgs(value)
			if err != nil {
				return nil, fmt.Errorf("error parsing args in line '%s': %w", line, err)
			}
			result.Args = append(result.Args, args...)

		default:
			processSimpleField(result.SimpleFields, normalizedKey, value)
		}
	}

	return result, nil
}

func parseMultiLineDescription(line string, configMap map[string]string, processing bool) bool {
	multiLineDescPrefixed := strings.HasPrefix(line, multiLineDescriptionTag)
	multiLineDescSuffixed := strings.HasSuffix(line, multiLineDescriptionTag)
	existingValue := configMap[DescriptionConfigurationKey]
	switch {
	case processing && multiLineDescPrefixed:
		processing = false
	case processing && multiLineDescSuffixed:
		processing = false
		line = strings.TrimSuffix(line, multiLineDescriptionTag)
		if existingValue != "" {
			configMap[DescriptionConfigurationKey] = fmt.Sprintf(
				"%s%s%s",
				existingValue,
				descriptionSeparator,
				line,
			)
		} else {
			configMap[DescriptionConfigurationKey] = line
		}
	case processing:
		if existingValue != "" {
			configMap[DescriptionConfigurationKey] = fmt.Sprintf(
				"%s%s%s",
				existingValue,
				descriptionSeparator,
				line,
			)
		} else {
			configMap[DescriptionConfigurationKey] = line
		}
	case multiLineDescPrefixed:
		processing = true
		line = strings.TrimPrefix(line, multiLineDescriptionTag)
		if trimmedLine := strings.TrimSpace(line); trimmedLine == "" {
			return processing
		} else if existingValue != "" {
			configMap[DescriptionConfigurationKey] = fmt.Sprintf(
				"%s%s%s",
				existingValue,
				descriptionSeparator,
				line,
			)
		} else {
			configMap[DescriptionConfigurationKey] = line
		}
	}
	return processing
}

func parseParams(value string) (executable.ParameterList, error) {
	if strings.TrimSpace(value) == "" {
		return executable.ParameterList{}, nil
	}

	var params executable.ParameterList
	items := splitValue(value, InternalListSeparator)

	for i, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		fields := splitValue(item, InternalValueSeparator)
		if len(fields) < 3 {
			return nil, fmt.Errorf("param %d requires at least 3 fields (type:value:envKey), got %d: %s",
				i+1, len(fields), item)
		}

		paramType := strings.TrimSpace(fields[0])
		paramValue := cleanValue(strings.TrimSpace(fields[1]))
		envKey := strings.TrimSpace(fields[2])

		param := executable.Parameter{
			EnvKey: envKey,
		}

		switch paramType {
		case "secretRef":
			param.SecretRef = paramValue
		case "prompt":
			param.Prompt = paramValue
		case "text":
			param.Text = paramValue
		default:
			return nil, fmt.Errorf("invalid parameter type: %s (expected secretRef, prompt, or text)", paramType)
		}

		if err := param.Validate(); err != nil {
			return nil, fmt.Errorf("error validating parameter %d: %w", i+1, err)
		}

		params = append(params, param)
	}

	return params, nil
}

func parseArgs(value string) (executable.ArgumentList, error) {
	if strings.TrimSpace(value) == "" {
		return executable.ArgumentList{}, nil
	}

	var args executable.ArgumentList
	items := splitValue(value, InternalListSeparator)

	for i, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		fields := splitValue(item, InternalValueSeparator)
		if len(fields) != 3 {
			return nil, fmt.Errorf("arg %d requires exactly 3 fields (type:name:envKey), got %d: %s",
				i+1, len(fields), item)
		}

		argType := strings.TrimSpace(fields[0])
		typeVal := cleanValue(strings.TrimSpace(fields[1]))
		envKey := strings.TrimSpace(fields[2])

		arg := executable.Argument{
			EnvKey: envKey,
			Type:   executable.ArgumentTypeString,
		}

		switch argType {
		case "flag":
			arg.Flag = typeVal
		case "pos":
			pos, err := strconv.Atoi(typeVal)
			if err != nil {
				return nil, fmt.Errorf("invalid position number: %s", typeVal)
			}
			arg.Pos = &pos
		default:
			return nil, fmt.Errorf("invalid argument type: %s (expected flag or pos)", argType)
		}

		if err := arg.Validate(); err != nil {
			return nil, fmt.Errorf("error validating argument %d: %w", i+1, err)
		}

		args = append(args, arg)
	}

	return args, nil
}

func splitValue(s, delimiter string) []string {
	if s == "" {
		return []string{}
	}

	var result []string
	var current strings.Builder
	escaped := false

	runes := []rune(s)
	delimiterRunes := []rune(delimiter)

	for i := 0; i < len(runes); i++ {
		if escaped {
			current.WriteRune(runes[i])
			escaped = false
			continue
		}

		if runes[i] == '\\' {
			escaped = true
			current.WriteRune(runes[i]) // Keep the backslash for later processing
			continue
		}

		// Check if we're at the start of a delimiter
		if i+len(delimiterRunes) <= len(runes) {
			match := true
			for j, delimRune := range delimiterRunes {
				if runes[i+j] != delimRune {
					match = false
					break
				}
			}

			if match {
				result = append(result, current.String())
				current.Reset()
				i += len(delimiterRunes) - 1 // Skip the delimiter
				continue
			}
		}

		current.WriteRune(runes[i])
	}

	// Add the last part
	if current.Len() > 0 || len(result) == 0 {
		result = append(result, current.String())
	}

	return result
}

func normalizeKey(key string) string {
	switch key {
	case "tags":
		return TagConfigurationKey
	case "aliases":
		return AliasConfigurationKey
	case "params":
		return ParamConfigurationKey
	case "args":
		return ArgConfigurationKey
	case descriptionAlias:
		return DescriptionConfigurationKey
	default:
		return key
	}
}

func cleanValue(value string) string {
	if len(value) >= 2 &&
		((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
		value = value[1 : len(value)-1]
	}

	// Unescape characters in the correct order
	value = strings.ReplaceAll(value, `\|`, `|`)
	value = strings.ReplaceAll(value, `\:`, `:`)
	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\'`, `'`)
	value = strings.ReplaceAll(value, `\\`, `\`)

	return strings.TrimSpace(value)
}

func processSimpleField(fields map[string]string, key, value string) {
	switch key {
	case AliasConfigurationKey, TagConfigurationKey:
		if existing, ok := fields[key]; ok {
			fields[key] = fmt.Sprintf("%s|%s", existing, value)
		} else {
			fields[key] = value
		}
	case DescriptionConfigurationKey, descriptionAlias:
		if existing, ok := fields[DescriptionConfigurationKey]; ok {
			fields[DescriptionConfigurationKey] = fmt.Sprintf("%s%s%s", existing, descriptionSeparator, value)
		} else {
			fields[DescriptionConfigurationKey] = value
		}
	default:
		fields[key] = value
	}
}

func validateKey(key string) bool {
	switch key {
	case VerbConfigurationKey, NameConfigurationKey, DescriptionConfigurationKey, descriptionAlias,
		AliasConfigurationKey, VisibilityConfigurationKey, TagConfigurationKey,
		TimeoutConfigurationKey, ParamConfigurationKey, ArgConfigurationKey,
		DirConfigurationKey, LogModeConfigurationKey:
		return true
	default:
		return false
	}
}
