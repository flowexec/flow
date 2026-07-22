package executable_test

import (
	"testing"

	"github.com/flowexec/flow/v2/types/executable"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func TestContainerSetDefaults(t *testing.T) {
	c := &executable.ExecContainer{Image: "alpine:3"}
	c.SetDefaults()
	if c.Runtime != executable.ExecContainerRuntimeAuto {
		t.Errorf("runtime default = %q, want auto", c.Runtime)
	}
	if c.MountWorkspace != executable.DefaultContainerMount {
		t.Errorf("mountWorkspace default = %q, want %q", c.MountWorkspace, executable.DefaultContainerMount)
	}
}

func TestContainerSetDefaultsPreservesExplicit(t *testing.T) {
	c := &executable.ExecContainer{
		Image:          "alpine:3",
		Runtime:        executable.ExecContainerRuntimeDocker,
		MountWorkspace: "/src",
	}
	c.SetDefaults()
	if c.Runtime != executable.ExecContainerRuntimeDocker {
		t.Errorf("runtime = %q, want docker", c.Runtime)
	}
	if c.MountWorkspace != "/src" {
		t.Errorf("mountWorkspace = %q, want /src", c.MountWorkspace)
	}
}

func TestContainerEnvInherited(t *testing.T) {
	cases := []struct {
		name string
		c    *executable.ExecContainer
		want bool
	}{
		{"nil container", nil, true},
		{"unset", &executable.ExecContainer{}, true},
		{"explicit true", &executable.ExecContainer{InheritEnv: boolPtr(true)}, true},
		{"explicit false", &executable.ExecContainer{InheritEnv: boolPtr(false)}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.EnvInherited(); got != tc.want {
				t.Errorf("EnvInherited() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestContainerResolveEntrypoint(t *testing.T) {
	cases := []struct {
		name         string
		c            *executable.ExecContainer
		wantEntry    string
		wantOverride bool
	}{
		{"unset -> sh override", &executable.ExecContainer{}, "sh", true},
		{"empty -> image entrypoint", &executable.ExecContainer{Entrypoint: strPtr("")}, "", false},
		{"custom -> override", &executable.ExecContainer{Entrypoint: strPtr("/bin/bash")}, "/bin/bash", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry, override := tc.c.ResolveEntrypoint()
			if entry != tc.wantEntry || override != tc.wantOverride {
				t.Errorf("ResolveEntrypoint() = (%q, %v), want (%q, %v)", entry, override, tc.wantEntry, tc.wantOverride)
			}
		})
	}
}

func TestContainerValidate(t *testing.T) {
	type ec = executable.ExecContainer
	vols := func(v ...executable.ExecContainerVolume) []executable.ExecContainerVolume { return v }
	cases := []struct {
		name    string
		c       *ec
		wantErr bool
	}{
		{"valid minimal", &ec{Image: "alpine:3"}, false},
		{"missing image", &ec{}, true},
		{"invalid runtime", &ec{Image: "x", Runtime: executable.ExecContainerRuntime("containerd")}, true},
		{"relative workdir", &ec{Image: "x", Workdir: "sub/dir"}, true},
		{"relative mountWorkspace", &ec{Image: "x", MountWorkspace: "rel"}, true},
		{"valid volume", &ec{Image: "x", Volumes: vols("/host:/container")}, false},
		{"valid volume with options", &ec{Image: "x", Volumes: vols("/host:/container:ro")}, false},
		{"volume missing container", &ec{Image: "x", Volumes: vols("/host")}, true},
		{"volume relative container", &ec{Image: "x", Volumes: vols("/host:rel")}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.c.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
