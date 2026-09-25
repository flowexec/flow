//nolint:testpackage // tests unexported URI parsing helpers
package mcp

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestExtractExecutableURIParts(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		wantWS   string
		wantNS   string
		wantName string
	}{
		{
			name:     "fully qualified",
			uri:      "flow://executable/myws/myns/myexec",
			wantWS:   "myws",
			wantNS:   "myns",
			wantName: "myexec",
		},
		{
			name:     "empty namespace",
			uri:      "flow://executable/myws//myexec",
			wantWS:   "myws",
			wantNS:   "",
			wantName: "myexec",
		},
		{
			name:     "empty workspace",
			uri:      "flow://executable//myns/myexec",
			wantWS:   "",
			wantNS:   "myns",
			wantName: "myexec",
		},
		{
			name:     "empty workspace and namespace",
			uri:      "flow://executable///myexec",
			wantWS:   "",
			wantNS:   "",
			wantName: "myexec",
		},
		{
			name:     "nested namespace",
			uri:      "flow://executable/myws/parent/child/myexec",
			wantWS:   "myws",
			wantNS:   "parent/child",
			wantName: "myexec",
		},
		{
			name:     "malformed missing segments",
			uri:      "flow://executable/onlyname",
			wantWS:   "",
			wantNS:   "",
			wantName: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractExecutableURIParts(tc.uri)
			if got.workspace != tc.wantWS {
				t.Errorf("workspace: got %q, want %q", got.workspace, tc.wantWS)
			}
			if got.namespace != tc.wantNS {
				t.Errorf("namespace: got %q, want %q", got.namespace, tc.wantNS)
			}
			if got.name != tc.wantName {
				t.Errorf("name: got %q, want %q", got.name, tc.wantName)
			}
		})
	}
}

func TestExecutableURITemplateMatchesNestedNamespace(t *testing.T) {
	tmpl := mcp.NewResourceTemplate(executableURITemplate, "test")
	for _, uri := range []string{
		"flow://executable/myws/myns/myexec",
		"flow://executable/myws/parent/child/myexec",
		"flow://executable/myws//myexec",
	} {
		if !tmpl.URITemplate.Regexp().MatchString(uri) {
			t.Errorf("template did not match %q", uri)
		}
	}
}
