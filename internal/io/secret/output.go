package secret

import (
	"fmt"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/vault"
)

func PrintSecrets(ctx *context.Context, vaultName string, vlt vault.Vault, format string, plaintext bool) {
	secrets, err := vault.NewSecretList(vaultName, vlt)
	if err != nil {
		logger.Log().FatalErr(err)
	}

	if plaintext {
		secrets = secrets.AsPlaintext()
	} else {
		secrets = secrets.AsObfuscatedText()
	}

	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := secrets.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal secrets - %v", err)
		}
		_, _ = fmt.Fprint(ctx.StdOut(), str)
	case common.JSONFormat:
		str, err := secrets.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal secrets - %v", err)
		}
		_, _ = fmt.Fprint(ctx.StdOut(), str)
	}
}
