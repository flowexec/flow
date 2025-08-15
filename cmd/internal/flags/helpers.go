package flags

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/flowexec/flow/internal/logger"
)

//nolint:errcheck
func ToPflag(cmd *cobra.Command, metadata Metadata, persistent bool) (*pflag.FlagSet, error) {
	flagSet := cmd.Flags()
	if persistent {
		flagSet = cmd.PersistentFlags()
	}
	if metadata.Default == nil {
		return nil, fmt.Errorf("metadata default must be defined using explicit type")
	}
	switch reflect.TypeOf(metadata.Default).Kind() { //nolint:exhaustive
	case reflect.String:
		if metadata.Shorthand != "" {
			flagSet.StringP(metadata.Name, metadata.Shorthand, metadata.Default.(string), metadata.Usage)
		} else {
			flagSet.String(metadata.Name, metadata.Default.(string), metadata.Usage)
		}
	case reflect.Bool:
		if metadata.Shorthand != "" {
			flagSet.BoolP(metadata.Name, metadata.Shorthand, metadata.Default.(bool), metadata.Usage)
		} else {
			flagSet.Bool(metadata.Name, metadata.Default.(bool), metadata.Usage)
		}
	case reflect.Slice:
		var def []string
		var ok bool
		if def, ok = metadata.Default.([]string); !ok {
			return nil, fmt.Errorf("unexpected type received for %s metadata", metadata.Name)
		}

		if len(def) == 0 {
			def = nil
		}

		if metadata.Shorthand != "" {
			flagSet.StringArrayP(metadata.Name, metadata.Shorthand, def, metadata.Usage)
		} else {
			flagSet.StringArray(metadata.Name, def, metadata.Usage)
		}
	case reflect.Int:
		if metadata.Shorthand != "" {
			flagSet.IntP(metadata.Name, metadata.Shorthand, metadata.Default.(int), metadata.Usage)
		} else {
			flagSet.Int(metadata.Name, metadata.Default.(int), metadata.Usage)
		}
	default:
		return nil, fmt.Errorf("unexpected metadata default type (%v)", reflect.TypeOf(metadata.Default).Kind())
	}

	if metadata.Required {
		if err := cmd.MarkFlagRequired(metadata.Name); err != nil {
			return nil, err
		}
	}

	return flagSet, nil
}

func HasFlag(cmd *cobra.Command, metadata Metadata) bool {
	flagName := metadata.Name
	flagSet := cmd.Flags()
	return flagSet.Lookup(flagName) != nil
}

func ValueFor[T any](cmd *cobra.Command, metadata Metadata, persistent bool) T {
	flagName := metadata.Name
	flagSet := cmd.Flags()
	if persistent {
		flagSet = cmd.PersistentFlags()
	}
	flag := cmd.Flag(flagName)
	if flag == nil {
		logger.Log().FatalErr(fmt.Errorf("flag %s not found", flagName))
	}

	var val interface{}
	switch reflect.TypeOf(metadata.Default).Kind() { //nolint:exhaustive
	case reflect.String:
		s, err := flagSet.GetString(flagName)
		if err != nil {
			logger.Log().FatalErr(err)
		}
		val = s
	case reflect.Bool:
		b, err := flagSet.GetBool(flagName)
		if err != nil {
			logger.Log().FatalErr(err)
		}
		val = b
	case reflect.Slice:
		s, err := flagSet.GetStringArray(flagName)
		if err != nil {
			logger.Log().FatalErr(err)
		}
		val = s
	case reflect.Int:
		i, err := flagSet.GetInt(flagName)
		if err != nil {
			logger.Log().FatalErr(err)
		}
		val = i
	default:
		logger.Log().Fatalf("unexpected flag default type (%v)", reflect.TypeOf(metadata.Default).Kind())
	}

	//nolint:errcheck
	return val.(T)
}
