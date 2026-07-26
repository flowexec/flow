package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/vault"
	"gopkg.in/yaml.v3"
)

type SecretRef string

func (r SecretRef) Key() string {
	parts := strings.Split(string(r), "/")
	if len(parts) < 2 {
		return string(r)
	}
	return parts[1]
}

func (r SecretRef) Vault() string {
	parts := strings.Split(string(r), "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

type Secret interface {
	vault.Secret
	types.Entity

	Ref() SecretRef
	AsPlaintext() Secret
	AsObfuscatedText() Secret
	// IsPlaintext reports whether the secret is currently in plaintext (unobfuscated) mode.
	IsPlaintext() bool
	// Reference returns where this secret lives when it is a link into another
	// system, and the empty string when the vault stores the value itself.
	Reference() string
}

type SecretValue = vault.SecretValue

type secret struct {
	vault     string
	key       string
	reference string
	plaintext bool
	value     vault.Secret
}

// enrichedSecret is used for JSON/YAML marshaling to control how the value is serialized
type enrichedSecret struct {
	Vault string `json:"vault" yaml:"vault"`
	Key   string `json:"key"   yaml:"key"`
	Value string `json:"value" yaml:"value"`
	// Reference is present only for secrets linked from another system. It is
	// the path, not the secret, so it is safe to print unmasked -- and it is the
	// only useful thing to show when the value has not been resolved.
	Reference string `json:"reference,omitempty" yaml:"reference,omitempty"`
}

func NewSecret(vaultName, key string, value vault.Secret) (Secret, error) {
	if err := ValidateIdentifier(vaultName); err != nil {
		return nil, err
	}
	if key == "" {
		return nil, errors.New("key cannot be empty")
	} else if vaultName == "" {
		return nil, errors.New("vault name cannot be empty")
	}

	return &secret{
		vault: vaultName,
		key:   key,
		value: value,
	}, nil
}

// NewLinkedSecret builds a secret that points at another system.
//
// value may be nil: listing a read-through vault deliberately does not resolve
// every link, because that would run one provider command per entry.
func NewLinkedSecret(vaultName, key, reference string, value vault.Secret) (Secret, error) {
	s, err := NewSecret(vaultName, key, value)
	if err != nil {
		return nil, err
	}
	s.(*secret).reference = reference
	return s, nil
}

func NewSecretValue(value []byte) *SecretValue {
	return vault.NewSecretValue(value)
}

func (s *secret) Ref() SecretRef {
	return SecretRef(fmt.Sprintf("%s/%s", s.vault, s.key))
}

func (s *secret) AsPlaintext() Secret {
	s.plaintext = true
	return s
}

func (s *secret) AsObfuscatedText() Secret {
	s.plaintext = false
	return s
}

func (s *secret) IsPlaintext() bool {
	return s.plaintext
}

func (s *secret) YAML() (string, error) {
	yamlBytes, err := yaml.Marshal(toEnrichedSecret(s))
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret - %w", err)
	}
	return string(yamlBytes), nil
}

func (s *secret) JSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(toEnrichedSecret(s), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret - %w", err)
	}
	return string(jsonBytes), nil
}

func (s *secret) Markdown() string {
	var mkdwn string

	mkdwn = fmt.Sprintf("# [Secret] %s\n", s.Ref())

	valueStr := s.value.String()
	if s.plaintext {
		valueStr = s.value.PlainTextString()
	}

	mkdwn += fmt.Sprintf("**Value**\n```\n%s\n```", valueStr)
	return mkdwn
}

func (s *secret) String() string {
	return s.value.String()
}

func (s *secret) PlainTextString() string {
	return s.value.PlainTextString()
}

func (s *secret) Bytes() []byte {
	return s.value.Bytes()
}

func (s *secret) Zero() {
	s.value.Zero()
}

func RefToParts(ref SecretRef) (vaultName, key string, err error) {
	parts := strings.Split(string(ref), "/")
	if len(parts) == 1 {
		return "", parts[0], nil
	} else if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid secret reference format: %s", ref)
	}
	vaultName = parts[0]
	key = parts[1]
	if key == "" || vaultName == "" {
		return "", "", fmt.Errorf("vault name and key cannot be empty: %s", ref)
	}
	return vaultName, key, nil
}

func toEnrichedSecret(s Secret) enrichedSecret {
	return toEnrichedSecretWithMode(s, s.IsPlaintext())
}

// toEnrichedSecretWithMode allows explicit control over plaintext vs obfuscated
func toEnrichedSecretWithMode(s Secret, plaintext bool) enrichedSecret {
	valueStr := s.String()
	if plaintext {
		valueStr = s.PlainTextString()
	}

	return enrichedSecret{
		Vault:     s.Ref().Vault(),
		Key:       s.Ref().Key(),
		Value:     valueStr,
		Reference: s.Reference(),
	}
}

type SecretList []Secret

// NewSecretList builds the list of secrets in a vault.
//
// resolve controls whether each secret's value is actually read. For a vault
// that stores its own secrets that is nearly free, but a read-through vault runs
// one provider command per key -- and for 1Password, potentially one biometric
// prompt per key -- so callers that are only going to print masks must pass
// false. The reference is still shown, which is the part worth seeing anyway.
func NewSecretList(vaultName string, v Vault, resolve bool) (SecretList, error) {
	keys, err := v.ListSecrets()
	if err != nil {
		return nil, err
	}

	links, linked := AsReferenceVault(v)

	result := make(SecretList, 0, len(keys))
	for _, key := range keys {
		var reference string
		if linked {
			// An unreadable reference is reported in place rather than dropping
			// the entry: a key that exists is a key the user should see.
			if reference, err = links.Reference(key); err != nil {
				reference = fmt.Sprintf("<unresolved: %v>", err)
			}
		}

		var value vault.Secret
		if resolve || !linked {
			// Errors are deliberately not fatal here. A broken link -- the
			// secret was removed in the provider -- used to make the whole
			// entry vanish from the listing, which reads as "it was never
			// there" rather than "this needs re-linking".
			value, _ = v.GetSecret(key)
		}
		if value == nil {
			if !linked {
				continue
			}
			value = NewSecretValue(nil)
		}

		scrt, err := NewLinkedSecret(vaultName, key, reference, value)
		if err != nil {
			return nil, err
		}
		result = append(result, scrt)
	}

	return result, nil
}

type enrichedSecretList struct {
	Secrets []enrichedSecret `json:"secrets" yaml:"secrets"`
}

func (l SecretList) AsPlaintext() SecretList {
	result := make(SecretList, 0, len(l))
	for _, s := range l {
		if s == nil {
			continue
		}
		result = append(result, s.AsPlaintext())
	}
	return result
}

func (l SecretList) AsObfuscatedText() SecretList {
	result := make(SecretList, 0, len(l))
	for _, s := range l {
		if s == nil {
			continue
		}
		result = append(result, s.AsObfuscatedText())
	}
	return result
}

func (l SecretList) YAML() (string, error) {
	scrts := make([]enrichedSecret, 0, len(l))
	for _, s := range l {
		if s == nil {
			continue
		}
		scrts = append(scrts, toEnrichedSecret(s))
	}
	enriched := enrichedSecretList{Secrets: scrts}
	yamlBytes, err := yaml.Marshal(enriched)
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret list - %w", err)
	}
	return string(yamlBytes), nil
}

func (l SecretList) JSON() (string, error) {
	scrts := make([]enrichedSecret, 0, len(l))
	for _, s := range l {
		if s == nil {
			continue
		}
		scrts = append(scrts, toEnrichedSecret(s))
	}
	enriched := enrichedSecretList{Secrets: scrts}
	jsonBytes, err := json.MarshalIndent(enriched, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret list - %w", err)
	}
	return string(jsonBytes), nil
}

// YAMLWithMode allows explicit control over plaintext vs obfuscated serialization
func (l SecretList) YAMLWithMode(plaintext bool) (string, error) {
	scrts := make([]enrichedSecret, 0, len(l))
	for _, s := range l {
		if s == nil {
			continue
		}
		scrts = append(scrts, toEnrichedSecretWithMode(s, plaintext))
	}
	enriched := enrichedSecretList{Secrets: scrts}
	yamlBytes, err := yaml.Marshal(enriched)
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret list - %w", err)
	}
	return string(yamlBytes), nil
}

// JSONWithMode allows explicit control over plaintext vs obfuscated serialization
func (l SecretList) JSONWithMode(plaintext bool) (string, error) {
	scrts := make([]enrichedSecret, 0, len(l))
	for _, s := range l {
		if s == nil {
			continue
		}
		scrts = append(scrts, toEnrichedSecretWithMode(s, plaintext))
	}
	enriched := enrichedSecretList{Secrets: scrts}
	jsonBytes, err := json.MarshalIndent(enriched, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret list - %w", err)
	}
	return string(jsonBytes), nil
}

func (l SecretList) FindByName(name string) Secret {
	for _, scrt := range l {
		if scrt.Ref().Key() == name {
			return scrt
		}
	}
	return nil
}

func (l SecretList) Items() []*types.EntityInfo {
	items := make([]*types.EntityInfo, 0)
	for _, s := range l {
		item := types.EntityInfo{
			Header: s.Ref().Key(),
			ID:     string(s.Ref()),
		}
		items = append(items, &item)
	}
	return items
}

func (l SecretList) Singular() string {
	return "secret"
}

func (l SecretList) Plural() string {
	return "secrets"
}

func ValidateIdentifier(reference string) error {
	if reference == "" {
		return errors.New("reference cannot be empty")
	}
	// Must stay a strict subset of the vault library's own ValidateVaultID, which
	// requires an alphanumeric first character.
	re := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-_]*$`)
	if !re.MatchString(reference) {
		return fmt.Errorf(
			"reference (%s) must start with a letter or digit and contain only "+
				"alphanumeric characters, dashes and/or underscores",
			reference,
		)
	}
	return nil
}

// Reference returns where a linked secret lives, or "" when the vault stores
// the value itself.
func (s *secret) Reference() string {
	return s.reference
}
