---
title: Working with Secrets
---

# Working with Secrets

flow's built-in vault keeps your sensitive data secure while making it easy to use in your workflows.
Whether you're managing API keys, database passwords, or deployment tokens, the vault has you covered.

## Quick Start

Create your first vault and add a secret:

```shell
# Create a vault and set it as current
flow vault create my-vault --set

# Extract and store the generated key using JSON output
export FLOW_VAULT_KEY=$(flow vault create my-vault --set --output json | jq -r '.result.data.generatedKey')

# Add a secret (you'll be prompted for the value)
flow secret set database-password
```

```yaml
# Use it in an executable
executables:
  - verb: backup
    name: database
    exec:
      params:
        - secretRef: database-password
          envKey: DB_PASSWORD
      cmd: pg_dump -h localhost -U admin mydb
```

## Vault Types

flow supports multiple vault backends for different security needs:

:::tabs

== AES256 (Default)

Symmetric encryption using a generated key. This is the simplest vault type - flow generates a random encryption key for you.

```shell
# Create an AES256 vault (default type)
flow vault create myapp
# or explicitly specify the type
flow vault create myapp --type aes256
```

This creates an AES256-encrypted vault with a randomly generated key that will be displayed in the output.
Store this key securely - if you lose it, you won't be able to access your secrets.

**Key Management Options:**
```shell
# Store key in environment variable
flow vault create myapp --key-env MY_VAULT_KEY

# Store key in file
flow vault create myapp --key-file ~/mykeys/myapp.key
```

**Key Sharing:**
If you specify a `--key-env` and that environment variable already contains a valid encryption key, the vault will use that existing key instead of generating a new one:

```shell
# Create first vault and extract the generated key
export SHARED_VAULT_KEY=$(flow vault create dev --key-env SHARED_VAULT_KEY --output json \
  | jq -r '.result.data.generatedKey')

# Create additional vaults using the same key
flow vault create staging --key-env SHARED_VAULT_KEY
```

> [!NOTE]
> **Valid Key Format**: The existing key must be a base64-encoded 32-byte (256-bit) encryption key. You can generate a compatible key using `flow vault create` and copying the output, or by generating 32 random bytes and base64-encoding them. If the environment variable contains an invalid key format, vault creation will fail.

== Age

Asymmetric encryption using recipient keys. This is ideal for team vaults where multiple people need access.

**Prerequisites:**
Install and use the [age-keygen](https://github.com/FiloSottile/age) tool to generate keys:

```shell
# Generate an age identity (private key)
age-keygen -o ~/.age/identity.txt

# Extract the public key (recipient) from the identity
age-keygen -y ~/.age/identity.txt
```

The public key output is what you use as recipients, and the identity file contains your private key for decryption.

**Creating Age Vaults:**
```shell
# Create vault with recipient keys
flow vault create team --type age --recipients key1,key2,key3 --identity-file ~/.age/identity.txt

# With identity environment variable
flow vault create team --type age --recipients key1,key2,key3 --identity-env MY_IDENTITY
```

== Unencrypted
A simple vault that stores secrets in plain text JSON files.
This is not recommended for production use but can be useful for development or testing.

```shell
# Create an unencrypted vault
flow vault create dev --type unencrypted
```


== Keyring

A vault that uses your operating system's keyring for managing secrets.
This is a good option for personal use where you want seamless integration with your OS security.

```shell
# Create a keyring vault
flow vault create dev --type keyring
```

== External (other CLI tools)

An external vault reads secrets that already live in another tool — 1Password, `pass`,
AWS SSM — through that tool's CLI.

It holds **links**, not secrets. Each link pairs a name you choose with a *reference*
the provider understands. Reading the name resolves the reference and reads through.
Nothing is copied into flow, and nothing is ever written back, so pointing a vault at a
store you already use cannot damage it.

```shell
flow secret link aws-access-key 'op://Team/AWS/access_key_id'
flow secret link aws-secret-key 'op://Team/AWS/secret_access_key'

flow secret get aws-access-key --plaintext
```

References use each provider's own syntax:

| Provider | Reference looks like |
|----------|----------------------|
| 1Password | `op://Team/AWS/access_key_id` |
| pass | `team/db/password` |
| AWS SSM | `/prod/db/password` |

The configuration only needs to say how to read one:

```json
{
  "id": "work",
  "type": "external",
  "external": {
    "get": {
      "cmd": "pass show '{{ref}}'"
    },
    "metadata": {
      "cmd": "cat \"$PASSWORD_STORE_DIR/.gpg-id\""
    },
    "reference_pattern": "^[A-Za-z0-9][A-Za-z0-9 ._/-]{0,255}$",
    "not_found_pattern": "is not in the password store",
    "environment": {
      "PASSWORD_STORE_DIR": "$PASSWORD_STORE_DIR"
    },
    "timeout": "30s"
  }
}
```

`reference_pattern` describes what a reference for this provider looks like, so a typo is
caught when you link it rather than weeks later when you read it. `not_found_pattern`
separates "this link is broken" from "the provider is unreachable" — without it, an
expired session is indistinguishable from a deleted secret.

> [!INFO]
> See the [flowexec/vault examples](https://github.com/flowexec/vault/tree/main/examples)
> for ready-to-use configurations for 1Password, pass, AWS SSM and Bitwarden.

```shell
# Create an external vault
flow vault create work --type external --config /path/to/config.json

# Point a name at a secret that already exists
flow secret link db-password 'team/db/password'

# Remove the link. The secret in the provider is untouched.
flow secret unlink db-password
```

**Template Variables**

Available in `cmd` and `output` fields:

- <span v-pre>`{{ref}}`</span> - The reference this name is linked to. **This is what a provider command should use.**
- <span v-pre>`{{key}}`</span> - The local alias (also available as `id`, `name`)
- <span v-pre>`{{env["VariableName"]}}`</span>- Environment variable value
- <span v-pre>`{{output}}`</span> - Raw command output (for output templates)

All [Expr language](https://expr-lang.org/docs/language-definition) operators and functions can be used in the command templates, allowing for powerful dynamic secret management.

> [!WARNING]
> **External vaults are read-only.** `flow secret set` fails on one, and `flow secret remove`
> removes the *link* rather than the secret. Create secrets in the tool that owns them, then
> link them.

:::

### Authentication

The environment variable or file that you provide at setup will be used to resolve the encryption key when accessing the vault.
If you did not provide a key or file, these default environment variables will be used:

- For AES256 vaults: `FLOW_VAULT_KEY` environment variable
- For Age vaults: `FLOW_VAULT_IDENTITY` environment variable
- For Unencrypted vaults: no key is needed, it stores secrets in plain text
- For Keyring vaults: no key is needed, it uses the OS keyring directly
- For External vaults: no key is needed, it uses the external CLI tool directly. Auth may be required by the tool itself

At least one of the key or file will be used. You can configure key storage during vault creation:

```shell
# Expect to store the key in a specific environment variable
flow vault create myapp --key-env MY_VAULT_KEY

# Store key in file (file is created with the key if it doesn't exist)
flow vault create myapp --key-file ~/mykeys/myapp.key

# Age vault with identity file
flow vault create team --type age --identity-file ~/identities/identity.txt --identity-env MY_IDENTITY
```

## Using Secrets in Workflows

### Basic Usage

```yaml
executables:
  - verb: deploy
    name: app
    exec:
      params:
        - secretRef: api-key
          envKey: API_KEY
        - secretRef: database-url
          envKey: DATABASE_URL
      cmd: ./deploy.sh
```

### Cross-Vault References

Reference secrets from different vaults:

```yaml
executables:
  - verb: sync
    name: environments
    exec:
      params:
        - secretRef: production/api-key
          envKey: PROD_API_KEY
        - secretRef: staging/api-key
          envKey: STAGING_API_KEY
      cmd: ./sync-environments.sh
```

## Secret Management

### Adding Secrets

```shell
# Interactive prompt (recommended)
flow secret set my-secret

# From command line (less secure)
flow secret set my-secret "secret-value"

# From file
cat secret.txt | flow secret set my-secret
# OR
flow secret set my-secret --file secret.txt
```

### Viewing Secrets

```shell
# List all secrets (values hidden)
flow secret list

# Get specific secret (obfuscated)
flow secret get my-secret

# Get plaintext value
flow secret get my-secret --plaintext

# Copy to clipboard
flow secret get my-secret --copy

# Extract value as clean JSON (no ANSI colors, ideal for scripts and tools)
flow secret get my-secret --plaintext --output json | jq -r '.result.data.value'
```

### Updating and Removing

```shell
# Update a secret (prompts for new value)
flow secret set existing-secret

# Remove a secret
flow secret remove old-secret
```

### Working with Multiple Vaults

When working with multiple vaults, secrets are isolated per vault but the vault's name can be used to reference secrets across vaults.
You can retrieve secrets from a specific vault without switching to it by using the vault name as a prefix:

```shell
# Retrieve secrets from different vaults without switching
flow secret get production/db-password
flow secret get development/api-key
```

## Vault Management

See the [vault command reference](../cli/flow_vault.md) for detailed commands and options.

### Vault Configuration

```shell
# View the current vault
flow vault get

# View specific vault details
flow vault get my-vault

# Edit vault settings
flow vault edit my-vault --key-env NEW_KEY_VAR

# Remove a vault's configuration. The encrypted secret data remains on disk at its
# storage path; only the vault config flow tracks is deleted.
flow vault remove old-vault
```

#### Custom Vault Storage Location

You can specify a custom storage location for the encrypted data when creating a vault:

```shell
flow vault create myapp --path /storage/myapp
```

This data is encrypted, so you can safely store it as-is without worrying about plaintext secrets being exposed.

### Managing Multiple Vaults

Switch between vaults for different projects or environments:

```shell
# List all vaults
# Authentication for the created vaults must be resolvable by the environment variable or file you
# specified during vault creation in order to list them.
flow vault list

# Switch to a different vault
flow vault switch production

# Work with secrets in current vault
flow secret set api-key
flow secret list
```

### Backup and Recovery

Vault data is stored in your flow config directory:

```shell
# Find your vaults
ls ~/.config/flow/vaults/  # Linux
ls ~/Library/Caches/flow/vaults/  # macOS

# Backup (encrypted data is safe to copy)
cp -r ~/.config/flow/vaults/ ~/backups/
```

Each vault you create gets its own configuration file and data file.
You can back up these directories to ensure you have a copy of your vaults.
Note that if you are using a custom storage path, you should include that in your backup strategy.
