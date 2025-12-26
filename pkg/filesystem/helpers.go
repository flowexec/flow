package filesystem

import (
	"os"
	"strings"

	cp "github.com/otiai10/copy"
)

const dataDirName = "flow"

func CopyFile(src, dst string) error {
	opts := cp.Options{
		PreserveTimes: true,
		PreserveOwner: true,
		OnError: func(src, dest string, err error) error {
			switch {
			case err == nil:
				return nil
			case strings.Contains(err.Error(), src):
				return err
			case os.IsExist(err):
				return nil
			case os.IsNotExist(err):
				if _, err := os.Create(dest); err != nil {
					return err
				}
				return nil
			}
			return err
		},
	}
	return cp.Copy(src, dst, opts)
}
