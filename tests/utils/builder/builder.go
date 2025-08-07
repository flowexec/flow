package builder

import (
	"github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
)

func sharedExecTags() executable.ExecutableTags {
	return executable.ExecutableTags{"generated"}
}

func privateFlowFileVisibility() *executable.FlowFileVisibility {
	v := executable.FlowFileVisibility(common.VisibilityPrivate)
	return &v
}

func privateExecVisibility() *executable.ExecutableVisibility {
	v := executable.ExecutableVisibility(common.VisibilityPrivate)
	return &v
}
