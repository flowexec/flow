package executable

import (
	"fmt"
	"strconv"

	"github.com/flowexec/flow/internal/utils"
)

func (a *Argument) Set(value string) {
	a.value = value
}

func (a *Argument) Value() string {
	if a.value == "" {
		return a.Default
	}
	return a.value
}

func (a *Argument) Validate() error {
	if err := utils.ValidateOneOf("argument type", a.Flag, a.Pos); err != nil {
		return err
	}
	if err := utils.ValidateOneOf("argument destination", a.EnvKey, a.OutputFile); err != nil {
		return err
	}

	if err := validateArgType(a.Type); err != nil {
		return fmt.Errorf("%s - %w", a.EnvKey, err)
	}
	return nil
}

func (a *Argument) ValidateValue() error {
	if a.value == "" && a.Required {
		return fmt.Errorf("required argument not set")
	}

	switch a.Type {
	case ArgumentTypeInt:
		if _, err := strconv.Atoi(a.value); err != nil {
			return fmt.Errorf("value is not an integer")
		}
	case ArgumentTypeFloat:
		if _, err := strconv.ParseFloat(a.value, 64); err != nil {
			return fmt.Errorf("value is not a float")
		}
	case ArgumentTypeBool:
		if _, err := strconv.ParseBool(a.value); err != nil {
			return fmt.Errorf("value is not a boolean")
		}
	case ArgumentTypeString, "":
		// no-op
	default:
		// no-op, assume string
	}
	return nil
}

func validateArgType(t ArgumentType) error {
	switch t {
	case ArgumentTypeString, ArgumentTypeInt, ArgumentTypeBool, ArgumentTypeFloat:
		return nil
	case "":
		// type is assumed to be a string
		return nil
	default:
		return fmt.Errorf("unsupported argument type (%s)", t)
	}
}

func (al *ArgumentList) Flags() []string {
	var flags []string
	for _, arg := range *al {
		if arg.Flag != "" {
			flags = append(flags, arg.Flag)
		}
	}
	return flags
}

func (al *ArgumentList) Validate() error {
	var errs []error
	for _, arg := range *al {
		if err := arg.Validate(); err != nil {
			errs = append(
				errs,
				fmt.Errorf("argument (envKey=%s outputFile=%s) validation failed - %w", arg.EnvKey, arg.OutputFile, err),
			)
		}
	}
	collectedFlags := make(map[string]struct{})
	collectedPos := make(map[int]struct{})
	for _, arg := range *al {
		if arg.Flag != "" {
			if _, ok := collectedFlags[arg.Flag]; ok {
				errs = append(errs, fmt.Errorf("flag %s is assigned to more than one argument", arg.Flag))
			}
			collectedFlags[arg.Flag] = struct{}{}
		} else if arg.Pos != nil && *arg.Pos != 0 {
			if _, ok := collectedPos[*arg.Pos]; ok {
				errs = append(errs, fmt.Errorf("position %d is assigned to more than one argument", *arg.Pos))
			}
			collectedPos[*arg.Pos] = struct{}{}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d argument validation errors: %v", len(errs), errs)
	}
	return nil
}

func (al *ArgumentList) ValidateValues() error {
	var errs []error
	for _, arg := range *al {
		if err := arg.ValidateValue(); err != nil {
			errs = append(errs, fmt.Errorf("argument %s validation failed - %w", arg.EnvKey, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d argument validation errors: %v", len(errs), errs)
	}
	return nil
}
