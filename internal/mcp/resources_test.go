//nolint:testpackage // tests unexported URI parsing helpers
package mcp

import (
	"testing"
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
