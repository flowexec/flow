package cache

import (
	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/logger"
)

func PrintCache(cache map[string]string, format string) {
	if cache == nil {
		return
	}
	output := cacheData{Cache: cache}
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := output.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal cache - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := output.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal cache - %v", err)
		}
		logger.Log().Println(str)
	}
}
