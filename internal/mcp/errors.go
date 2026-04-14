package mcp

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

// Machine-readable error codes for structured error responses.
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

type errorPayload struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// toolError returns a CallToolResult with IsError set and a structured JSON error payload.
func toolError(code, message string) *mcp.CallToolResult {
	return toolErrorWithDetails(code, message, nil)
}

// toolErrorWithDetails is like toolError but includes a details object in the error payload.
func toolErrorWithDetails(code, message string, details map[string]any) *mcp.CallToolResult {
	payload := errorPayload{Error: errorDetail{
		Code:    code,
		Message: message,
		Details: details,
	}}
	data, err := json.Marshal(payload)
	if err != nil {
		return mcp.NewToolResultError(message)
	}
	return mcp.NewToolResultError(string(data))
}
