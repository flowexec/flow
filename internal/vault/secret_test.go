package vault_test

import (
	"strings"
	"testing"

	extVault "github.com/flowexec/vault"

	"github.com/flowexec/flow/v2/internal/vault"
)

const testSecretValue = "hunter2-super-secret"

func newTestSecret(t *testing.T, key, value string) vault.Secret {
	t.Helper()
	s, err := vault.NewSecret("testvault", key, vault.NewSecretValue([]byte(value)))
	if err != nil {
		t.Fatalf("NewSecret(%q) returned error: %v", key, err)
	}
	return s
}

// A secret list serialized in obfuscated mode must never leak the plaintext value.
// This guards against a regression where serialization always emitted plaintext.
func TestSecretList_YAML_ObfuscatedByDefault(t *testing.T) {
	list := vault.SecretList{newTestSecret(t, "token", testSecretValue)}.AsObfuscatedText()

	out, err := list.YAML()
	if err != nil {
		t.Fatalf("YAML() error: %v", err)
	}
	if strings.Contains(out, testSecretValue) {
		t.Fatalf("obfuscated YAML leaked the plaintext secret value:\n%s", out)
	}
	if !strings.Contains(out, "********") {
		t.Fatalf("obfuscated YAML did not contain the masked value:\n%s", out)
	}
}

func TestSecretList_JSON_ObfuscatedByDefault(t *testing.T) {
	list := vault.SecretList{newTestSecret(t, "token", testSecretValue)}.AsObfuscatedText()

	out, err := list.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}
	if strings.Contains(out, testSecretValue) {
		t.Fatalf("obfuscated JSON leaked the plaintext secret value:\n%s", out)
	}
}

func TestSecretList_YAML_PlaintextWhenRequested(t *testing.T) {
	list := vault.SecretList{newTestSecret(t, "token", testSecretValue)}.AsPlaintext()

	out, err := list.YAML()
	if err != nil {
		t.Fatalf("YAML() error: %v", err)
	}
	if !strings.Contains(out, testSecretValue) {
		t.Fatalf("plaintext YAML did not contain the secret value:\n%s", out)
	}
}

// AsPlaintext previously allocated a zero-length slice then indexed into it, panicking on
// any non-empty list. Verify it converts every element without panicking.
func TestSecretList_AsPlaintext_NoPanic(t *testing.T) {
	list := vault.SecretList{
		newTestSecret(t, "k1", "a"),
		newTestSecret(t, "k2", "b"),
		newTestSecret(t, "k3", "c"),
	}

	result := list.AsPlaintext()
	if len(result) != 3 {
		t.Fatalf("AsPlaintext() returned %d secrets, want 3", len(result))
	}
	for i, s := range result {
		if !s.IsPlaintext() {
			t.Errorf("result[%d] is not in plaintext mode", i)
		}
	}
}

func TestSecretList_AsObfuscatedText_NoPanic(t *testing.T) {
	list := vault.SecretList{
		newTestSecret(t, "k1", "a"),
		newTestSecret(t, "k2", "b"),
	}
	if got := len(list.AsObfuscatedText()); got != 2 {
		t.Fatalf("AsObfuscatedText() returned %d secrets, want 2", got)
	}
}

func TestValidateIdentifier(t *testing.T) {
	valid := []string{"myvault", "my_vault", "my-vault", "abc123", "A1_b-2"}
	for _, name := range valid {
		if err := vault.ValidateIdentifier(name); err != nil {
			t.Errorf("ValidateIdentifier(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{
		"", "../etc", "a/b", "a.b", "a b", "..", "vault/../x", "name.json",
		"-myvault", "_myvault", "-", "_",
	}
	for _, name := range invalid {
		if err := vault.ValidateIdentifier(name); err == nil {
			t.Errorf("ValidateIdentifier(%q) = nil, want error", name)
		}
	}
}

// Any vault name flow accepts must also be acceptable to the vault library,
// which derives filesystem paths and keyring entry names from it. If flow were
// the laxer of the two, a name would pass flow's check and then be refused
// downstream -- and any vault already created under such a name would become
// unreachable. Asserted rather than assumed, because the two rules live in
// different repositories and will drift.
func TestValidateIdentifierIsStricterThanTheVaultLibrary(t *testing.T) {
	candidates := []string{
		"myvault", "my_vault", "my-vault", "abc123", "A1_b-2", "v1",
		"-myvault", "_myvault", "-", "_", "a.b", "..", "a/b", "a b", "",
		"name.json", "../etc", "vault/../x",
	}

	for _, name := range candidates {
		if vault.ValidateIdentifier(name) != nil {
			continue // flow rejects it; the library's opinion does not matter
		}
		if err := extVault.ValidateVaultID(name); err != nil {
			t.Errorf("flow accepts vault name %q but the vault library rejects it: %v", name, err)
		}
	}
}
