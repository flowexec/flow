package vault

import (
	"strings"
	"time"

	"github.com/flowexec/vault"
)

const DemoVaultReservedName = "demo"

type demoVaultProvider struct {
	data map[string]string
}

func (d demoVaultProvider) GetSecret(key string) (vault.Secret, error) {
	if value, exists := d.data[key]; exists {
		return vault.NewSecretValue([]byte(value)), nil
	}
	return nil, vault.ErrSecretNotFound
}

func (d demoVaultProvider) SetSecret(_ string, _ vault.Secret) error {
	// do nothing - read only
	return nil
}

func (d demoVaultProvider) DeleteSecret(_ string) error {
	// do nothing - read only
	return nil
}

func (d demoVaultProvider) ListSecrets() ([]string, error) {
	keys := make([]string, 0, len(d.data))
	for key := range d.data {
		keys = append(keys, key)
	}
	return keys, nil
}

func (d demoVaultProvider) HasSecret(key string) (bool, error) {
	if _, exists := d.data[key]; exists {
		return true, nil
	}
	return false, nil
}

func (d demoVaultProvider) ID() string {
	return "demo"
}

func (d demoVaultProvider) Metadata() vault.Metadata {
	return vault.Metadata{
		Created:      time.Now().Local().Add(-24 * time.Hour),
		LastModified: time.Now().Local(),
	}
}

func (d demoVaultProvider) Close() error {
	return nil
}

func newDemoVault() Vault {
	return &demoVaultProvider{data: demoData()}
}

func newDemoVaultConfig() *VaultConfig {
	return &VaultConfig{ID: DemoVaultReservedName, Type: "demo"}
}

//nolint:lll
func demoData() map[string]string {
	return map[string]string{
		// Basic secrets (common use cases)
		"api-key":        "demo-api-key-12345-dont-use-in-production",
		"database-url":   "postgres://demo:password@localhost:5432/flowdb",
		"admin-password": "admin123-change-me-immediately",
		"webhook-secret": "webhook-secret-abcdef123456",
		"jwt-secret":     "super-secret-jwt-key-for-demo-only",
		"slack-webhook":  "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
		"message":        "Thanks for trying flow! Report bugs at github.com/flowexec/flow",

		// Long secrets (testing edge cases)
		"rsa-private-key": `-----BEGIN RSA PRIVATE KEY-----
DEMOdATAAAKCAQEA1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMN
OPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQR
STUVWXYZabcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuv
wxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890
-----END RSA PRIVATE KEY-----`,
		"long-token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkRlbW8gVXNlciIsImlhdCI6MTUxNjIzOTAyMiwiZXhwIjo5OTk5OTk5OTk5LCJyb2xlcyI6WyJhZG1pbiIsInVzZXIiLCJkZW1vIl0sInBlcm1pc3Npb25zIjpbInJlYWQiLCJ3cml0ZSIsImRlbGV0ZSJdLCJjdXN0b21fZGF0YSI6eyJkZW1vIjp0cnVlLCJlbnZpcm9ubWVudCI6InRlc3RpbmciLCJ2ZXJzaW9uIjoiMS4wLjAifX0.demo-signature-not-real",

		// JSON/structured data
		"config-json": `{
  "environment": "demo",
  "debug": true,
  "features": {
    "new_ui": true,
    "beta_api": false
  },
  "limits": {
    "requests_per_minute": 1000,
    "max_file_size": "10MB"
  }
}`,
		"env-vars": `NODE_ENV=development
DEBUG=flow:*
DATABASE_URL=postgres://demo:password@localhost/flowdb
REDIS_URL=redis://localhost:6379
API_BASE_URL=https://api.demo.com`,

		// Special characters and edge cases
		"special-chars":   `!@#$%^&*()_+-=[]{}|;:'",./<>?~\`,
		"unicode-secret":  "🔐🔑✨🚀🎉🌟💫🔥⭐🎯🎪🎨🎭🎪",
		"whitespace-only": "   \t\n   ",

		// Testing various lengths
		"empty-secret": "",
		"tiny":         "x",
		"short":        "hello",
		"medium":       "this-is-a-medium-length-secret-for-testing-purposes",
		"huge":         strings.Repeat("demo-data-", 100) + "end", // 900+ characters

		// Base64 encoded data
		"base64-cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURYVENDQWtXZ0F3SUJBZ0lKQUt4bVFuSjVkZW1kTUEwR0NTcUdTSWIzRFFFQkN3VUFNRVV4Q3pBSkJnTlYKQkFZVEFsVlRNUk13RVFZRFZRUUlEQXBUYjIxbExWTjBZWFJsTVNFd0h3WURWUVFLREJoSmJuUmxjbTVsZENCWAphV1JuYVhSeklGQjBlU0JNZEdRd0hoY05NVGt3TkRJeU1Ea3dOekE1V2hjTk1qa3dOREU1TURrd056QTVXakJGCk1Rc3dDUVlEVlFRR0V3SlZVekVUTUJFR0ExVUVDQXdLVTI5dFpTMVRkR0YwWlRFaE1COEdBMVVFQ2d3WVNXNTAKWlhKdVpYUWdWMmxrWjJsMGN5QlFkSGtnVEhSa01JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQgpDZ0tDQVFFQXRWcDNlcVJrVjkrSE9uVUJRcVg0c1MxTXVJZjVzZGM4ZGVNTzF3PT0=",
	}
}
