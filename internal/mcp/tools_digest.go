package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/flowexec/flow/v2/internal/digest"
	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
)

// mcpDefaultMaxChars is lower than the CLI's default: an agent pays for every character in context,
// and can ask for more once it knows it needs it.
const mcpDefaultMaxChars = 12000

func addDigestTools(srv *server.MCPServer, executor CommandExecutor) {
	getDigest := mcp.NewTool("get_digest",
		mcp.WithDescription("Runs, executables and files in one budgeted document, e.g. a failed run's log "+
			"next to its task's definition."),
		mcp.WithArray("sources", mcp.Required(), mcp.WithStringItems(),
			mcp.Description("`run:<id>`, `runs:<filter>` (ref:, status:, ran:7d, limit:), `exec:<verb>/<id>`, "+
				"`file:<path>` (workspace-relative)")),
		mcp.WithNumber("max_chars", mcp.Description(fmt.Sprintf(
			"Content budget (default %d); logs shrink first", mcpDefaultMaxChars))),
		mcp.WithString("dir", mcp.Description("Your working directory")),
	)
	getDigest.Annotations = mcp.ToolAnnotation{
		Title:           "Get a digest of project state",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(getDigest, getDigestHandler(executor))
}

func getDigestHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sources := request.GetStringSlice("sources", nil)
		if len(sources) == 0 {
			return toolError(ErrCodeInvalidInput, "sources must be a non-empty array of strings"), nil
		}
		maxChars := request.GetInt("max_chars", mcpDefaultMaxChars)
		if maxChars <= 0 {
			return toolError(ErrCodeInvalidInput, "max_chars must be positive"), nil
		}
		cmdArgs := []string{"digest", outputFlag, outputJSON, "--max-chars", strconv.Itoa(maxChars)}
		// `--` keeps a source that starts with a dash from being read as a flag.
		cmdArgs = append(cmdArgs, "--")
		cmdArgs = append(cmdArgs, sources...)

		output, err := executor.ExecuteContext(withRunDir(ctx, request.GetString("dir", "")), cmdArgs...)
		if err != nil {
			if env, perr := flowerrors.ParseEnvelope([]byte(lastLine(output))); perr == nil {
				return toolErrorWithDetails(env.Error.Code, env.Error.Message, env.Error.Details), nil
			}
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to build digest: %s", output)), nil
		}

		// Output is combined stdout+stderr, so a stray log line can sit around the document.
		var doc digest.Document
		start := strings.Index(output, "{")
		if start < 0 || json.NewDecoder(strings.NewReader(output[start:])).Decode(&doc) != nil {
			return mcp.NewToolResultText(output), nil //nolint:nilerr // unparseable output is still the answer
		}
		jsonData, _ := json.Marshal(doc)
		return mcp.NewToolResultStructured(doc, string(jsonData)), nil
	}
}

func lastLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		return s[i+1:]
	}
	return s
}
