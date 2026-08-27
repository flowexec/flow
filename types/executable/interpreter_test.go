package executable_test

import (
	"strings"
	"testing"

	"github.com/flowexec/flow/v2/types/executable"
)

func interpPtr(i executable.ExecInterpreter) *executable.ExecInterpreter { return &i }

func TestResolveInterpreterDefaultsToSh(t *testing.T) {
	// Interpreter is a pointer precisely so that unset stays distinguishable from
	// an explicit value; the generated models apply no schema default in Go.
	cases := map[string]*executable.ExecExecutableType{
		"nil spec":         nil,
		"unset":            {Cmd: "echo hi"},
		"explicitly empty": {Cmd: "echo hi", Interpreter: interpPtr("")},
	}
	for name, spec := range cases {
		if got := spec.ResolveInterpreter(); got != executable.InterpreterSh {
			t.Errorf("%s: ResolveInterpreter() = %q, want sh", name, got)
		}
	}
}

func TestResolveInterpreterHonoursExplicitPython(t *testing.T) {
	spec := &executable.ExecExecutableType{
		Cmd:         "print(1)",
		Interpreter: interpPtr(executable.InterpreterPython),
	}
	if got := spec.ResolveInterpreter(); got != executable.InterpreterPython {
		t.Errorf("ResolveInterpreter() = %q, want python", got)
	}
}

func TestInterpreterForFile(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		interpreter *executable.ExecInterpreter
		want        executable.ExecInterpreter
	}{
		{"py extension infers python", "script.py", nil, executable.InterpreterPython},
		{"uppercase extension still infers", "SCRIPT.PY", nil, executable.InterpreterPython},
		{"sh extension stays sh", "script.sh", nil, executable.InterpreterSh},
		{"unknown extension falls back to sh", "script.unknown", nil, executable.InterpreterSh},
		{
			"explicit interpreter overrides extension",
			"script.sh",
			interpPtr(executable.InterpreterPython),
			executable.InterpreterPython,
		},
		{
			"explicit sh overrides a py extension",
			"script.py",
			interpPtr(executable.InterpreterSh),
			executable.InterpreterSh,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &executable.ExecExecutableType{File: tt.file, Interpreter: tt.interpreter}
			if got := spec.InterpreterForFile(); got != tt.want {
				t.Errorf("InterpreterForFile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInterpreterIsSet(t *testing.T) {
	if (&executable.ExecExecutableType{}).InterpreterIsSet() {
		t.Error("unset interpreter reported as set")
	}
	if (&executable.ExecExecutableType{Interpreter: interpPtr("")}).InterpreterIsSet() {
		t.Error("empty interpreter reported as set")
	}
	spec := &executable.ExecExecutableType{Interpreter: interpPtr(executable.InterpreterPython)}
	if !spec.InterpreterIsSet() {
		t.Error("explicit interpreter reported as unset")
	}
}

func TestExecValidateRejectsUnknownInterpreter(t *testing.T) {
	// The generated models enforce no enum, so Go-side validation is the only
	// thing standing between a bad value and a late runtime failure.
	spec := &executable.ExecExecutableType{Cmd: "x", Interpreter: interpPtr("ruby")}
	err := spec.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want an error for an unknown interpreter")
	}
	if !strings.Contains(err.Error(), "ruby") {
		t.Errorf("Validate() error = %q, want it to name the invalid value", err)
	}
}

func TestExecValidateAcceptsKnownInterpreters(t *testing.T) {
	for _, i := range []*executable.ExecInterpreter{
		nil,
		interpPtr(""),
		interpPtr(executable.InterpreterSh),
		interpPtr(executable.InterpreterPython),
	} {
		spec := &executable.ExecExecutableType{Cmd: "x", Interpreter: i}
		if err := spec.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	}
	var nilSpec *executable.ExecExecutableType
	if err := nilSpec.Validate(); err != nil {
		t.Errorf("nil spec Validate() = %v, want nil", err)
	}
}
