package env

import (
	"errors"

	"github.com/flowexec/flow/internal/vault"
	"github.com/flowexec/flow/types/executable"
)

func ResolveParameterValue(
	currentVault string,
	param executable.Parameter,
	promptedEnv map[string]string,
) (string, error) {
	if val, found := promptedEnv[param.EnvKey]; found {
		// existing values win - these could come in as a param override from the CLI
		return val, nil
	}

	switch {
	case param.Text == "" && param.SecretRef == "" && param.Prompt == "":
		return "", nil
	case param.Text != "":
		return param.Text, nil
	case param.Prompt != "":
		val, ok := promptedEnv[param.EnvKey]
		if !ok {
			return "", errors.New("failed to get value for parameter")
		}
		return val, nil
	case param.SecretRef != "":
		return resolveSecretValue(currentVault, param.SecretRef)
	case param.OutputFile != "":
		return "", errors.New("outputFile parameter value should be resolved using ResolveParameterFileValue")
	default:
		return "", errors.New("failed to get value for parameter")
	}
}

func resolveSecretValue(
	currentVault string,
	secretRef string,
) (string, error) {
	rVault, key, err := vault.RefToParts(vault.SecretRef(secretRef))
	if err != nil {
		return "", err
	}
	if rVault == "" {
		rVault = currentVault
	}
	_, v, err := vault.VaultFromName(rVault)
	if err != nil {
		return "", err
	}
	defer v.Close()
	secret, err := v.GetSecret(key)
	if err != nil {
		return "", err
	}
	return secret.PlainTextString(), nil
}
