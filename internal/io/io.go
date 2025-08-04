package io

import "os"

const DisableInteractiveEnvKey = "DISABLE_FLOW_INTERACTIVE"

var (
	Stdout = os.Stdout
	Stdin  = os.Stdin
)
