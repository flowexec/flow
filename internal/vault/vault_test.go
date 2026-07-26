package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flowexec/flow/v2/internal/vault"
	"github.com/flowexec/flow/v2/pkg/filesystem"
)

// A crafted vault name must be rejected before it is turned into a filesystem path,
// so it cannot be used to read arbitrary files as vault configs.
func TestVaultFromName_RejectsTraversal(t *testing.T) {
	for _, name := range []string{"../../etc/passwd", "a/b", "..", "with space"} {
		_, _, err := vault.VaultFromName(name)
		if err == nil {
			t.Errorf("VaultFromName(%q) = nil error, want rejection", name)
			continue
		}
		if !strings.Contains(err.Error(), "invalid vault name") {
			t.Errorf("VaultFromName(%q) error = %v, want invalid-name rejection", name, err)
		}
	}
}

func TestVaultFromName_EmptyAndDemo(t *testing.T) {
	if _, _, err := vault.VaultFromName(""); err == nil {
		t.Error("VaultFromName(\"\") = nil error, want error")
	}

	cfg, v, err := vault.VaultFromName(vault.DemoVaultReservedName)
	if err != nil {
		t.Fatalf("VaultFromName(demo) error: %v", err)
	}
	if cfg == nil || v == nil {
		t.Fatal("VaultFromName(demo) returned nil config or provider")
	}
}

func TestListVaultNames_ReadsConfigsDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(filesystem.FlowCacheDirEnvVar, tmp)

	configs := vault.CacheDirectory("configs")
	if err := os.MkdirAll(configs, 0750); err != nil {
		t.Fatalf("mkdir configs: %v", err)
	}
	write := func(name string) {
		if err := os.WriteFile(filepath.Join(configs, name), []byte("{}"), 0600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write("alpha.json")
	write("beta.json")
	write("notes.txt") // non-config file, must be ignored

	names, err := vault.ListVaultNames()
	if err != nil {
		t.Fatalf("ListVaultNames() error: %v", err)
	}
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Fatalf("ListVaultNames() = %v, want [alpha beta] sorted", names)
	}
}

func TestListVaultNames_MissingDir(t *testing.T) {
	t.Setenv(filesystem.FlowCacheDirEnvVar, t.TempDir())
	names, err := vault.ListVaultNames()
	if err != nil {
		t.Fatalf("ListVaultNames() on missing dir error: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("ListVaultNames() on missing dir = %v, want empty", names)
	}
}

func TestVaultExists(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(filesystem.FlowCacheDirEnvVar, tmp)
	configs := vault.CacheDirectory("configs")
	if err := os.MkdirAll(configs, 0750); err != nil {
		t.Fatalf("mkdir configs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configs, "alpha.json"), []byte("{}"), 0600); err != nil {
		t.Fatalf("write alpha: %v", err)
	}

	cases := map[string]bool{
		"alpha":                     true,
		vault.DemoVaultReservedName: true,
		"missing":                   false,
		"../escape":                 false,
		"a/b":                       false,
	}
	for name, want := range cases {
		if got := vault.VaultExists(name); got != want {
			t.Errorf("VaultExists(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestRemoveVaultConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(filesystem.FlowCacheDirEnvVar, tmp)
	configs := vault.CacheDirectory("configs")
	if err := os.MkdirAll(configs, 0750); err != nil {
		t.Fatalf("mkdir configs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configs, "alpha.json"), []byte("{}"), 0600); err != nil {
		t.Fatalf("write alpha: %v", err)
	}

	if err := vault.RemoveVaultConfig("alpha"); err != nil {
		t.Fatalf("RemoveVaultConfig(alpha) error: %v", err)
	}
	if vault.VaultExists("alpha") {
		t.Error("alpha still exists after RemoveVaultConfig")
	}
	// Removing a non-existent vault is not an error.
	if err := vault.RemoveVaultConfig("alpha"); err != nil {
		t.Errorf("RemoveVaultConfig on missing vault error: %v", err)
	}
	// Invalid names are rejected rather than acted on.
	if err := vault.RemoveVaultConfig("../escape"); err == nil {
		t.Error("RemoveVaultConfig(../escape) = nil, want error")
	}
}
