package errors

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ExitFunc is invoked by the CLI error handler to terminate the process after
// a structured envelope has been written. It is overridable so tests can
// observe the exit code (and message) without actually exiting.
var ExitFunc = os.Exit

// Machine-readable error codes for structured error responses across the CLI and MCP surfaces.
const (
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeExecutionFailed  = "EXECUTION_FAILED"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeCancelled        = "CANCELLED"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
)

// Envelope is the top-level JSON/YAML payload for structured error responses.
type Envelope struct {
	Error Detail `json:"error" yaml:"error"`
}

// Detail is the structured error body inside an Envelope.
type Detail struct {
	Code    string         `json:"code"              yaml:"code"`
	Message string         `json:"message"           yaml:"message"`
	Details map[string]any `json:"details,omitempty" yaml:"details,omitempty"`
}

// NewEnvelope builds an Envelope from the given code, message, and optional details.
func NewEnvelope(code, message string, details map[string]any) Envelope {
	return Envelope{Error: Detail{Code: code, Message: message, Details: details}}
}

// Coder is implemented by errors that carry a machine-readable error code.
type Coder interface {
	error
	Code() string
}

// ParseEnvelope attempts to decode data as a JSON or YAML error envelope.
// It returns the parsed Envelope and a nil error on success. If data does not
// contain a valid envelope (or has an empty error code), it returns a non-nil
// error — callers can use this to distinguish structured errors from plain text.
func ParseEnvelope(data []byte) (Envelope, error) {
	var env Envelope

	// Try JSON first (most common for --output json).
	if err := json.Unmarshal(data, &env); err == nil && env.Error.Code != "" {
		return env, nil
	}

	// Fall back to YAML (--output yaml).
	if err := yaml.Unmarshal(data, &env); err == nil && env.Error.Code != "" {
		return env, nil
	}

	return Envelope{}, fmt.Errorf("data is not a valid error envelope")
}
