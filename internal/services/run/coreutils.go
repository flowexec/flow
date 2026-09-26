package run

import (
	"os"
	"runtime"
	"strconv"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/x/coreutils"
)

// CoreUtilsEnv overrides whether the shell interpreter provides built-in core utilities
// (cp, mkdir, mv, rm, cat, ls, find, …). By default they are on for Windows only, where
// those commands are otherwise missing unless Git Bash's usr\bin is on PATH.
const CoreUtilsEnv = "FLOW_CORE_UTILS"

const goosWindows = "windows"

func useCoreUtils() bool {
	if v, err := strconv.ParseBool(os.Getenv(CoreUtilsEnv)); err == nil {
		return v
	}
	return runtime.GOOS == goosWindows
}

func execHandlers() []func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	if useCoreUtils() {
		return []func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc{coreutils.ExecHandler}
	}
	return nil
}
