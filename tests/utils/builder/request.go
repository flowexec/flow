package builder

import (
	"github.com/flowexec/flow/types/executable"
)

func RequestExec(opts ...Option) *executable.Executable {
	name := "request"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Request: &executable.RequestExecutableType{
			URL:    "https://httpbin.org/get",
			Method: "GET",
			Headers: map[string]string{
				"Authorization": "Bearer token",
				"User-Agent":    "flow",
			},
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func RequestExecWithBody(opts ...Option) *executable.Executable {
	name := "request-with-body"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Request: &executable.RequestExecutableType{
			URL:    "https://httpbin.org/post",
			Method: "POST",
			Body:   `{"key": "value"}`,
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func RequestExecWithTransform(opts ...Option) *executable.Executable {
	name := "request-with-transform"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Request: &executable.RequestExecutableType{
			URL:               "https://httpbin.org/get",
			TransformResponse: "status",
			ValidStatusCodes:  []int{200, 503}, // TODO: use a mock server to avoid relying on external services
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func RequestExecWithTimeout(opts ...Option) *executable.Executable {
	name := "request-with-timeout"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Request: &executable.RequestExecutableType{
			URL:     "https://httpbin.org/delay/3",
			Timeout: 1,
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}

func RequestExecWithValidatedStatus(opts ...Option) *executable.Executable {
	name := "request-with-validated-status"
	e := &executable.Executable{
		Verb:       "run",
		Name:       name,
		Visibility: privateExecVisibility(),
		Request: &executable.RequestExecutableType{
			URL:              "https://httpbin.org/status/400",
			ValidStatusCodes: []int{200},
		},
	}
	if len(opts) > 0 {
		vals := NewOptionValues(opts...)
		e.SetContext(vals.WorkspaceName, vals.WorkspacePath, vals.NamespaceName, vals.FlowFilePath)
	}
	return e
}
