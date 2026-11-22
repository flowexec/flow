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

func addDigestTools(srv *server.MCPServer, executor CommandExecutor) {
	getDigest := mcp.NewTool("get_digest",
		mcp.WithDescription("Gather project state into one document: run logs, run history, executable "+
			"definitions and file contents. Each source is a type:value pointer: `run:<id>` (metadata and log "+
			"tail), `runs:<filter>` (table of runs; terms `ref:`, `status:completed|failed|cancelled|running`, "+
			"`ran:7d`, `limit:`, bare words match the ref), `exec:<verb>/<id>` (definition, script and last 5 "+
			"runs) or `file:<path>` (relative to the workspace root, which it can't leave)."),
		mcp.WithArray("sources", mcp.Required(), mcp.WithStringItems(),
			mcp.Description(`Sources to resolve, e.g. ["run:3f2a…", "runs:status:failed ran:1d", "exec:test/unit", `+
				`"file:flow.yaml"]`)),
		mcp.WithNumber("max_chars", mcp.Description(fmt.Sprintf(
			"Character budget shared by all sections (default %d). Logs shrink first, then files.",
			digest.DefaultMaxChars))),
		mcp.WithString("dir", mcp.Description(
			"Resolve the workspace from this directory. Set it to your working directory (a worktree or "+
				"fresh clone) so file and exec sources read that tree.")),
		mcp.WithOutputSchema[digest.Document](),
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
		cmdArgs := []string{"digest", outputFlag, outputJSON}
		if maxChars := request.GetInt("max_chars", 0); maxChars != 0 {
			if maxChars < 0 {
				return toolError(ErrCodeInvalidInput, "max_chars must be positive"), nil
			}
			cmdArgs = append(cmdArgs, "--max-chars", strconv.Itoa(maxChars))
		}
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
