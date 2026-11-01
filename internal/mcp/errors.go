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
	// ErrCodeNotStarted reports that flow failed before running anything.
	ErrCodeNotStarted = "NOT_STARTED"
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

// maxFailureMessageBytes limits the output repeated in a failure's message.
const maxFailureMessageBytes = 2048

// executionFailure returns the error result for a run that failed. The structured result holds the
// output; the message keeps only the end of it, where errors surface, so a large failure isn't sent twice.
func executionFailure(message, output string, err error) *mcp.CallToolResult {
	exitCode := -1
	started := false
	code := ErrCodeExecutionFailed
	var exitErr *CLIExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.Code
		started = exitErr.Started
	}
	if !started {
		code = ErrCodeNotStarted
	}
	capped, truncated := capOutput(output)
	tail := strings.TrimSpace(capped)
	if len(tail) > maxFailureMessageBytes {
		tail = tail[len(tail)-maxFailureMessageBytes:]
	}
	result := toolErrorWithDetails(
		code,
		fmt.Sprintf("%s: %s", message, tail),
		map[string]any{"exitCode": exitCode, "started": started, "truncated": truncated},
	)
	result.StructuredContent = ExecutionOutput{
		Output: capped, ExitCode: exitCode, Started: started, Truncated: truncated,
	}
	return result
}
