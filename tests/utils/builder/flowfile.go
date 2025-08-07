package builder

import (
	"github.com/flowexec/flow/types/executable"
)

func RootExecFlowFile(opts ...Option) *executable.FlowFile {
	d := &executable.FlowFile{
		Visibility: privateFlowFileVisibility(),
		Tags:       sharedExecTags(),
		Executables: []*executable.Executable{
			NamelessExec(opts...),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		d.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.FlowFilePath)
	}
	return d
}

func ExamplesExecFlowFile(opts ...Option) *executable.FlowFile {
	d := &executable.FlowFile{
		Namespace:  "examples",
		Visibility: privateFlowFileVisibility(),
		Tags:       sharedExecTags(),
		Executables: []*executable.Executable{
			SimpleExec(opts...),
			ExecWithPauses(opts...),
			ExecWithExit(opts...),
			ExecWithTmpDir(opts...),
			ExecWithArgs(opts...),
			ExecWithParams(opts...),
			ExecWithLogMode(opts...),
			ExecWithTimeout(opts...),
			ExecWithEnvOutputFiles(opts...),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		d.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.FlowFilePath)
	}
	return d
}

func ExamplesMultiExecFlowFile(opts ...Option) *executable.FlowFile {
	d := &executable.FlowFile{
		Namespace:  "examples",
		Visibility: privateFlowFileVisibility(),
		Tags:       sharedExecTags(),
		Executables: []*executable.Executable{
			SerialExecWithExit(opts...),
			ParallelExecWithExit(opts...),
			ParallelExecWithMaxThreads(opts...),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		d.SetDefaults()
		d.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.FlowFilePath)
	}
	return d
}

func ExamplesRequestExecFlowFile(opts ...Option) *executable.FlowFile {
	d := &executable.FlowFile{
		Namespace:  "examples",
		Visibility: privateFlowFileVisibility(),
		Tags:       sharedExecTags(),
		Executables: []*executable.Executable{
			RequestExec(opts...),
			RequestExecWithBody(opts...),
			RequestExecWithTimeout(opts...),
			RequestExecWithTransform(opts...),
			RequestExecWithValidatedStatus(opts...),
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		d.SetDefaults()
		d.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.FlowFilePath)
	}
	return d
}
