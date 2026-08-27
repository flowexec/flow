package executable

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	InterpreterSh     = ExecInterpreterSh
	InterpreterPython = ExecInterpreterPython

	// DefaultInterpreter is used when interpreter is unset.
	DefaultInterpreter = InterpreterSh
)

// ResolveInterpreter returns the interpreter that should run cmd.
//
// Interpreter is a pointer so that unset (nil) stays distinguishable from an
// explicit `sh`: an unset interpreter lets a file extension decide, while an
// explicit one overrides it. It is deliberately left nil rather than filled in by
// SetDefaults so `flow browse` and `flow get` do not print "sh" on every shell
// executable.
func (e *ExecExecutableType) ResolveInterpreter() ExecInterpreter {
	if e == nil || e.Interpreter == nil || *e.Interpreter == "" {
		return DefaultInterpreter
	}
	return *e.Interpreter
}

// InterpreterForFile returns the interpreter that should run the configured file.
// An explicit interpreter wins; otherwise it is inferred from the extension, and
// an unrecognised extension falls back to sh (flow's built-in POSIX interpreter),
// matching how RunFile has always treated unknown script types.
func (e *ExecExecutableType) InterpreterForFile() ExecInterpreter {
	if e == nil {
		return DefaultInterpreter
	}
	if e.Interpreter != nil && *e.Interpreter != "" {
		return *e.Interpreter
	}
	if strings.EqualFold(filepath.Ext(e.File), ".py") {
		return InterpreterPython
	}
	return DefaultInterpreter
}

// InterpreterIsSet reports whether an interpreter was explicitly configured.
func (e *ExecExecutableType) InterpreterIsSet() bool {
	return e != nil && e.Interpreter != nil && *e.Interpreter != ""
}

// Validate performs semantic validation that the generated models do not enforce.
// go-jsonschema runs with --only-models, so the schema's enum is applied to
// flowfiles by the JSON schema validator but never by Go's unmarshalling; an
// invalid value would otherwise reach the runner and fail late.
func (e *ExecExecutableType) Validate() error {
	if e == nil || e.Interpreter == nil {
		return nil
	}
	switch *e.Interpreter {
	case "", InterpreterSh, InterpreterPython:
	default:
		return fmt.Errorf("invalid interpreter %q (must be sh or python)", *e.Interpreter)
	}
	return nil
}
