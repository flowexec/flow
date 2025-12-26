package filesystem

import (
	"os"

	"github.com/pkg/errors"
)

func LogsDir() string {
	return CachedDataDirPath() + "/logs"
}

func EnsureLogsDir() error {
	if _, err := os.Stat(LogsDir()); os.IsNotExist(err) {
		err = os.MkdirAll(LogsDir(), 0750)
		if err != nil {
			return errors.Wrap(err, "unable to create logs directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "unable to check for logs directory")
	}
	return nil
}
