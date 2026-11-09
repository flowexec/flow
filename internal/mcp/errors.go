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

// executionResult turns a finished subprocess into a tool result. A run that started and ran to
// completion is a result whatever its exit code, since for a check a non-zero exit is the answer;
// only a run that never started (rejected input, trust refusal, a missing executable) is an error.
func executionResult(message, output string, err error) *mcp.CallToolResult {
	output = normalizeOutput(output)
	var exitErr *CLIExitError
	if err != nil && (!errors.As(err, &exitErr) || !exitErr.Started) {
		return executionFailure(message, output, err)
	}
	result := ExecutionOutput{Started: true}
	if exitErr != nil {
		result.ExitCode = exitErr.Code
	}
	result.Output, result.Truncated = capOutput(output)
	jsonData, _ := json.Marshal(result)
	return mcp.NewToolResultStructured(result, string(jsonData))
}

// normalizeOutput makes captured terminal output readable as text: CRLF becomes LF, and a line a
// progress indicator redrew with bare carriage returns keeps only its last frame.
func normalizeOutput(output string) string {
	if !strings.Contains(output, "\r") {
		return output
	}
	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	for i, line := range lines {
		frames := strings.Split(line, "\r")
		lines[i] = frames[len(frames)-1]
		for j := len(frames) - 1; j >= 0 && lines[i] == ""; j-- {
			lines[i] = frames[j]
		}
	}
	return strings.Join(lines, "\n")
}

// maxFailureMessageBytes limits the output repeated in a failure's message.
const maxFailureMessageBytes = 2048

// executionFailure returns the error result for a run that failed before it started, or could not be
// launched at all. The structured result holds the
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
	if tail == "" && exitErr == nil {
		tail = err.Error()
	}
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
