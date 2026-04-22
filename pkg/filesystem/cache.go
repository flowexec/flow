package filesystem

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const FlowCacheDirEnvVar = "FLOW_CACHE_DIR"

func CachedDataDirPath() string {
	if dir := os.Getenv(FlowCacheDirEnvVar); dir != "" {
		return dir
	}

	dirname, err := os.UserCacheDir()
	if err != nil {
		panic(errors.Wrap(err, "unable to get cache directory"))
	}
	return filepath.Join(dirname, dataDirName)
}

func EnsureCachedDataDir() error {
	dir := CachedDataDirPath()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0750)
		if err != nil {
			return errors.Wrap(err, "unable to create cache directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "unable to check for cache directory")
	}
	return nil
}
