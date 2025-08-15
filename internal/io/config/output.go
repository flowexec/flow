package config

import (
	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/config"
)

func PrintUserConfig(format string, userConfig *config.Config) {
	if userConfig == nil {
		logger.Log().Fatalf("Config type is nil")
	}

	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := userConfig.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal user config - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := userConfig.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal user config - %v", err)
		}
		logger.Log().Println(str)
	default:
		logger.Log().Fatalf("Unsupported output format %s", format)
	}
}
