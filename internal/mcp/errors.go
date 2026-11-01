package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
)

// Machine-readable error codes for structured error responses.
// Aliased from pkg/errors so the CLI and MCP surfaces share a single source of truth.
const (
	ErrCodeInvalidInput     = flowerrors.ErrCodeInvalidInput
	ErrCodeNotFound         = flowerrors.ErrCodeNotFound
	ErrCodeExecutionFailed  = flowerrors.ErrCodeExecutionFailed
	ErrCodeTimeout          = flowerrors.ErrCodeTimeout
	ErrCodeCancelled        = flowerrors.ErrCodeCancelled
	ErrCodeValidationFailed = flowerrors.ErrCodeValidationFailed
	ErrCodeInternal         = flowerrors.ErrCodeInternal
	ErrCodePermissionDenied = flowerrors.ErrCodePermissionDenied
)

// toolError returns a CallToolResult with IsError set and a structured JSON error payload.
func toolError(code, message string) *mcp.CallToolResult {
	return toolErrorWithDetails(code, message, nil)
}

// toolErrorWithDetails is like toolError but includes a details object in the error payload.
func toolErrorWithDetails(code, message string, details map[string]any) *mcp.CallToolResult {
	payload := flowerrors.NewEnvelope(code, message, details)
	data, err := json.Marshal(payload)
	if err != nil {
		return mcp.NewToolResultError(message)
	}
	return mcp.NewToolResultError(string(data))
}

// executionFailure returns the error result for a run that failed, keeping its output and exit
// code in the structured result so a client doesn't have to parse them out of the error text.
func executionFailure(message, output string, err error) *mcp.CallToolResult {
	exitCode := -1
	var exitErr *CLIExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.Code
	}
	capped, truncated := capOutput(output)
	result := toolErrorWithDetails(
		ErrCodeExecutionFailed,
		fmt.Sprintf("%s: %s", message, strings.TrimSpace(capped)),
		map[string]any{"exitCode": exitCode},
	)
	result.StructuredContent = ExecutionOutput{Output: capped, ExitCode: exitCode, Truncated: truncated}
	return result
}
