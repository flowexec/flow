package filesystem

import (
	"bufio"
	"bytes"
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

func LatestCachedDataDir() string {
	return CachedDataDirPath() + "/latestcache"
}

func LatestCachedDataFilePath(cacheKey string) string {
	return filepath.Join(LatestCachedDataDir(), cacheKey)
}

func EnsureCachedDataDir() error {
	if _, err := os.Stat(LatestCachedDataDir()); os.IsNotExist(err) {
		err = os.MkdirAll(LatestCachedDataDir(), 0750)
		if err != nil {
			return errors.Wrap(err, "unable to create cache directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "unable to check for cache directory")
	}

	return nil
}

func WriteLatestCachedData(cacheKey string, data []byte) error {
	if err := EnsureCachedDataDir(); err != nil {
		return errors.Wrap(err, "unable to ensure existence of cache directory")
	}

	file, err := os.OpenFile(filepath.Clean(LatestCachedDataFilePath(cacheKey)), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "unable to open cache data file")
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return errors.Wrap(err, "unable to truncate cache data file")
	}

	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, []byte("\n")...)
	}
	if _, err := file.Write(data); err != nil {
		return errors.Wrap(err, "unable to write cache data file")
	}

	return nil
}

func LoadLatestCachedData(cacheKey string) ([]byte, error) {
	if err := EnsureCachedDataDir(); err != nil {
		return nil, errors.Wrap(err, "unable to ensure existence of cache directory")
	}

	if _, err := os.Stat(LatestCachedDataFilePath(cacheKey)); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to stat cache data file")
	}

	file, err := os.Open(LatestCachedDataFilePath(cacheKey))
	if err != nil {
		return nil, errors.Wrap(err, "unable to open cache data file")
	}
	defer file.Close()

	data := make([]byte, 0)
	buf := bufio.NewReader(file)
	for {
		var line []byte
		line, err = buf.ReadBytes('\n')
		if err != nil {
			break
		}
		data = append(data, line...)
	}
	if err.Error() != "EOF" {
		return nil, errors.Wrap(err, "unable to read cache data file")
	}

	data = bytes.TrimSuffix(data, []byte("\n"))
	return data, nil
}
