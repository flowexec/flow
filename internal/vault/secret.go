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
}

type SecretValue = vault.SecretValue

type secret struct {
	vault     string
	key       string
	plaintext bool
	value     vault.Secret
}

// enrichedSecret is used for JSON/YAML marshaling to control how the value is serialized
type enrichedSecret struct {
	Vault string `json:"vault" yaml:"vault"`
	Key   string `json:"key"   yaml:"key"`
	Value string `json:"value" yaml:"value"`
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
	valueStr := s.String() // Default to obfuscated
	if s.AsPlaintext() != nil {
		valueStr = s.PlainTextString()
	}

	return enrichedSecret{
		Vault: s.Ref().Vault(),
		Key:   s.Ref().Key(),
		Value: valueStr,
	}
}

// toEnrichedSecretWithMode allows explicit control over plaintext vs obfuscated
func toEnrichedSecretWithMode(s Secret, plaintext bool) enrichedSecret {
	valueStr := s.String()
	if plaintext {
		valueStr = s.PlainTextString()
	}

	return enrichedSecret{
		Vault: s.Ref().Vault(),
		Key:   s.Ref().Key(),
		Value: valueStr,
	}
}

type SecretList []Secret

func NewSecretList(vaultName string, v Vault) (SecretList, error) {
	secrets, err := v.ListSecrets()
	if err != nil {
		return nil, err
	}

	result := make(SecretList, 0, len(secrets))
	for _, key := range secrets {
		s, _ := v.GetSecret(key)
		if s == nil {
			continue
		}
		scrt, err := NewSecret(vaultName, key, s)
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
	for i, s := range l {
		if s == nil {
			continue
		}
		result[i] = s.AsPlaintext()
	}
	return result
}

func (l SecretList) AsObfuscatedText() SecretList {
	result := make(SecretList, len(l))
	for i, s := range l {
		if s == nil {
			continue
		}
		result[i] = s.AsObfuscatedText()
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
	re := regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
	if !re.MatchString(reference) {
		return fmt.Errorf(
			"reference (%s) must only contain alphanumeric characters, dashes and/or underscores",
			reference,
		)
	}
	return nil
}
