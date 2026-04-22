package filesystem

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func LogsDir() string {
	return filepath.Join(StateDirPath(), "logs")
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
