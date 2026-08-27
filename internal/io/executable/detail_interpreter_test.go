package executable_test

import (
	"strings"
	"testing"

	io "github.com/flowexec/flow/v2/internal/io/executable"
	"github.com/flowexec/flow/v2/types/executable"
)

func interp(v executable.ExecInterpreter) *executable.ExecInterpreter { return &v }

// The browse views are how an interpreter is discovered without opening the
// flowfile, so each surface is pinned: the library subtitle, the detail heading,
// the interpreter line, and the fence language on the command block.
func TestBrowseSurfacesInterpreter(t *testing.T) {
	tests := []struct {
		name         string
		exec         *executable.ExecExecutableType
		wantSubtitle string
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "shell command",
			exec:         &executable.ExecExecutableType{Cmd: "echo hi"},
			wantSubtitle: "Shell Executable",
			wantContains: []string{"## Shell Configuration", "```sh"},
			wantMissing:  []string{"**Interpreter:**", "```python"},
		},
		{
			name:         "python command",
			exec:         &executable.ExecExecutableType{Cmd: "print('hi')", Interpreter: interp(executable.InterpreterPython)},
			wantSubtitle: "Python Executable",
			wantContains: []string{"## Python Configuration", "**Interpreter:** python", "```python"},
			wantMissing:  []string{"```sh"},
		},
		{
			// The extension already says python, so the heading follows it while
			// the interpreter line stays out - nothing was configured to report.
			name:         "py file without an explicit interpreter",
			exec:         &executable.ExecExecutableType{File: "report.py"},
			wantSubtitle: "Python Executable",
			wantContains: []string{"## Python Configuration", "report.py"},
			wantMissing:  []string{"**Interpreter:**"},
		},
		{
			name:         "explicit sh overrides a py extension",
			exec:         &executable.ExecExecutableType{File: "report.py", Interpreter: interp(executable.InterpreterSh)},
			wantSubtitle: "Shell Executable",
			wantContains: []string{"## Shell Configuration"},
			wantMissing:  []string{"## Python Configuration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executable.Executable{Verb: "run", Name: "probe", Exec: tt.exec}

			if got := io.ExecTypeNameForTest(e); got != tt.wantSubtitle {
				t.Errorf("execTypeName() = %q, want %q", got, tt.wantSubtitle)
			}

			body := io.ExecBodyMarkdownForTest(e)
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("detail body missing %q:\n%s", want, body)
				}
			}
			for _, missing := range tt.wantMissing {
				if strings.Contains(body, missing) {
					t.Errorf("detail body unexpectedly contains %q:\n%s", missing, body)
				}
			}
		})
	}
}
