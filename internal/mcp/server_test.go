package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	flowMcp "github.com/flowexec/flow/v2/internal/mcp"
	"github.com/flowexec/flow/v2/internal/mcp/mocks"
	"github.com/flowexec/flow/v2/pkg/filesystem"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MCP Server Suite")
}

var _ = Describe("MCP Server", func() {
	var (
		flowServer   *flowMcp.Server
		mockExecutor *mocks.MockCommandExecutor
		mcpClient    *client.Client
		ctx          context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		ctrl := gomock.NewController(GinkgoT())
		mockExecutor = mocks.NewMockCommandExecutor(ctrl)
		flowServer = flowMcp.NewServer(mockExecutor)

		var err error
		mcpClient, err = client.NewInProcessClient(flowServer.GetMCPServer())
		Expect(err).ToNot(HaveOccurred())

		// Initialize the client
		initRequest := mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
				ClientInfo: mcp.Implementation{
					Name:    "flow-test-client",
					Version: "1.0.0",
				},
				Capabilities: mcp.ClientCapabilities{},
			},
		}

		_, err = mcpClient.Initialize(ctx, initRequest)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if mcpClient != nil {
			mcpClient.Close()
		}
	})

	Describe("Server Initialization", func() {
		It("should create server successfully", func() {
			Expect(flowServer).ToNot(BeNil())
			Expect(mcpClient).ToNot(BeNil())
		})
	})

	Describe("Tool Registration", func() {
		It("should register all expected tools", func() {
			toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
			Expect(err).ToNot(HaveOccurred())

			toolNames := make([]string, len(toolsResult.Tools))
			for i, tool := range toolsResult.Tools {
				toolNames[i] = tool.Name
			}

			expectedTools := []string{
				"get_info",
				"get_workspace",
				"list_workspaces",
				"switch_workspace",
				"get_executable",
				"list_executables",
				"execute",
				"run_command",
				"run_executable",
				"get_execution_logs",
				"sync_executables",
				"write_flowfile",
				"get_workspace_config",
			}

			for _, expectedTool := range expectedTools {
				Expect(toolNames).To(ContainElement(expectedTool))
			}
		})

		It("should include output schema on the execution tools", func() {
			toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
			Expect(err).ToNot(HaveOccurred())

			// Output schemas are advertised only on the execution tools (small, high-value). The
			// read/list tools omit them to keep the always-loaded tool surface small — they still
			// return structuredContent at call time.
			toolsWithSchema := map[string]bool{
				"execute":        true,
				"run_command":    true,
				"run_executable": true,
			}

			for _, tool := range toolsResult.Tools {
				if toolsWithSchema[tool.Name] {
					Expect(tool.OutputSchema.Type).ToNot(BeEmpty(),
						"tool %s should have an output schema", tool.Name)
				}
			}
		})
	})

	Describe("Resource Template Registration", func() {
		It("should register all expected resource templates", func() {
			result, err := mcpClient.ListResourceTemplates(ctx, mcp.ListResourceTemplatesRequest{})
			Expect(err).ToNot(HaveOccurred())

			templates := make([]string, len(result.ResourceTemplates))
			for i, t := range result.ResourceTemplates {
				templates[i] = t.URITemplate.Raw()
			}

			expectedTemplates := []string{
				"flow://workspace/{name}",
				"flow://executable/{workspace}/{namespace}/{name}",
				"flow://flowfile/{+path}",
				"flow://logs/{run_id}",
			}

			for _, expected := range expectedTemplates {
				Expect(templates).To(ContainElement(expected))
			}
		})
	})

	Describe("Prompt Registration", func() {
		It("should register all expected prompts", func() {
			promptsResult, err := mcpClient.ListPrompts(ctx, mcp.ListPromptsRequest{})
			Expect(err).ToNot(HaveOccurred())

			promptNames := make([]string, len(promptsResult.Prompts))
			for i, prompt := range promptsResult.Prompts {
				promptNames[i] = prompt.Name
			}

			expectedPrompts := []string{
				"generate_executable",
				"generate_project_executables",
				"debug_executable",
				"explain_flow",
				"migrate_automation",
			}

			for _, expectedPrompt := range expectedPrompts {
				Expect(promptNames).To(ContainElement(expectedPrompt))

				result, err := mcpClient.GetPrompt(ctx, mcp.GetPromptRequest{
					Params: mcp.GetPromptParams{
						Name: expectedPrompt,
					},
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Description).ToNot(BeEmpty())
				Expect(result.Messages).ToNot(BeEmpty())
				Expect(result.Messages[0].Role).To(Equal(mcp.RoleUser))
				Expect(result.Messages[0].Content).ToNot(BeNil())
			}
		})
	})

	Describe("Tool Execution", func() {
		Context("get_info tool", func() {
			It("should return flow information", func() {
				testDir := GinkgoTB().TempDir()
				GinkgoTB().Setenv(filesystem.FlowConfigDirEnvVar, testDir)
				err := filesystem.InitConfig()
				Expect(err).ToNot(HaveOccurred())
				_, err = filesystem.LoadConfig()
				Expect(err).ToNot(HaveOccurred())

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_info", nil))
				Expect(err).ToNot(HaveOccurred())
				content := getTextContent(result)
				Expect(content).To(ContainSubstring("currentContext"))
				Expect(content).To(ContainSubstring("docsUrl"))
				Expect(content).To(ContainSubstring("llmsTxtUrl"))
				Expect(content).To(ContainSubstring("schemaUrls"))
			})
		})

		Context("get_workspace tool", func() {
			It("should call executor with correct arguments", func() {
				expectedOutput := "get ws execution results"
				mockExecutor.EXPECT().
					Execute("workspace", "get", "test-workspace", "--output", "json").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_workspace", map[string]interface{}{
					"workspace_name": "test-workspace",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
			})
		})

		Context("list_workspaces tool", func() {
			It("should call executor with correct arguments", func() {
				expectedOutput := "list ws execution result"
				mockExecutor.EXPECT().
					Execute("workspace", "list", "--output", "json").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("list_workspaces", nil))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
			})
		})

		Context("switch_workspace tool", func() {
			It("should call executor with correct arguments", func() {
				mockExecutor.EXPECT().
					Execute("workspace", "switch", "test-workspace").
					Return("", nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("switch_workspace", map[string]interface{}{
					"workspace_name": "test-workspace",
				}))

				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("get_executable tool", func() {
			It("should call executor with correct arguments for full reference", func() {
				expectedOutput := "get exec execution results"
				mockExecutor.EXPECT().
					Execute("browse", "--output", "json", "test", "test:test-exec").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_executable", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-exec",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
			})

			It("should handle missing executable_id", func() {
				expectedOutput := "get exec execution results without id"
				mockExecutor.EXPECT().
					Execute("browse", "--output", "json", "test").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_executable", map[string]interface{}{
					"executable_verb": "test",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
			})
		})

		Context("list_executables tool", func() {
			It("should call executor with correct arguments", func() {
				expectedOutput := "list execs execution results"
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "browse", "--output", "json", "--workspace", "*", "--namespace", "*").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("list_executables", nil))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
			})
		})

		Context("execute tool", func() {
			It("should call executor with provided arguments", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "test", "test:test-flow", "arg1", "arg2").
					Return("execution result", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
					"args":            "arg1 arg2",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("execution result"))
			})

			It("should handle no args", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "test", "test:test-flow").
					Return("execution result with no args", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("execution result with no args"))
			})

			It("should cap oversized output and report truncation", func() {
				huge := strings.Repeat("x", 300_000)
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "test", "test:test-flow").
					Return(huge, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
				}))
				Expect(err).ToNot(HaveOccurred())

				var out flowMcp.ExecutionOutput
				Expect(json.Unmarshal([]byte(getTextContent(result)), &out)).To(Succeed())
				Expect(out.Truncated).To(BeTrue())
				Expect(len(out.Output)).To(BeNumerically("<=", 200_000))
				// The tail is kept, since errors typically surface at the end of a run's output.
				Expect(out.Output).To(HaveSuffix(strings.Repeat("x", 100)))
			})
		})

		Context("run_command tool", func() {
			It("should run an ad-hoc command via flow exec --cmd with label and dir", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--cmd", "echo hi", "--label", "greet", "--dir", "/tmp").
					Return("command result", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_command", map[string]interface{}{
					"command": "echo hi",
					"label":   "greet",
					"dir":     "/tmp",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("command result"))
			})

			It("should run multiple commands in one call with a mode", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--cmd", "echo a", "--cmd", "echo b", "--mode", "parallel").
					Return("batch result", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_command", map[string]interface{}{
					"commands": []interface{}{"echo a", "echo b"},
					"mode":     "parallel",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("batch result"))
			})

			It("should require a command or commands", func() {
				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_command", map[string]interface{}{}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
			})

			It("should cap oversized output and report truncation", func() {
				huge := strings.Repeat("y", 300_000)
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--cmd", "echo hi").
					Return(huge, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_command", map[string]interface{}{
					"command": "echo hi",
				}))
				Expect(err).ToNot(HaveOccurred())

				var out flowMcp.ExecutionOutput
				Expect(json.Unmarshal([]byte(getTextContent(result)), &out)).To(Succeed())
				Expect(out.Truncated).To(BeTrue())
				Expect(len(out.Output)).To(BeNumerically("<=", 200_000))
			})
		})

		Context("run_executable tool", func() {
			It("should run a transient executable from an inline spec", func() {
				spec := `{"verb":"run","serial":{"execs":[{"cmd":"echo one"}]}}`
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--spec", spec, "--label", "batch").
					Return("spec result", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_executable", map[string]interface{}{
					"spec":  spec,
					"label": "batch",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("spec result"))
			})

			It("should require a spec", func() {
				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_executable", map[string]interface{}{}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
			})
		})

		Context("get_execution_logs tool", func() {
			It("should call executor with correct arguments", func() {
				expectedOutput := "execution logs result"
				mockExecutor.EXPECT().
					Execute("logs", "--output", "json", "--last").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_execution_logs", map[string]interface{}{
					"last": true,
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
			})

			It("should forward source/session/status filters to the CLI", func() {
				mockExecutor.EXPECT().
					Execute("logs", "--output", "json", "--source", "mcp", "--session", "sess-1", "--status", "running").
					Return(`{"history":[]}`, nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("get_execution_logs", map[string]interface{}{
					"source":  "mcp",
					"session": "sess-1",
					"status":  "running",
				}))

				Expect(err).ToNot(HaveOccurred())
			})

			It("should request content with the default byte cap when tail is set", func() {
				mockExecutor.EXPECT().
					Execute("logs", "--output", "json", "--content", "--max-bytes", "50000", "--tail", "20").
					Return(`{"history":[]}`, nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("get_execution_logs", map[string]interface{}{
					"tail": 20,
				}))

				Expect(err).ToNot(HaveOccurred())
			})

			It("should forward grep and clamp an oversized max_bytes", func() {
				mockExecutor.EXPECT().
					Execute("logs", "--output", "json", "--content", "--max-bytes", "200000", "--grep", "error").
					Return(`{"history":[]}`, nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("get_execution_logs", map[string]interface{}{
					"grep":      "error",
					"max_bytes": 999999,
				}))

				Expect(err).ToNot(HaveOccurred())
			})

			It("should not request content when no content option is set", func() {
				mockExecutor.EXPECT().
					Execute("logs", "--output", "json").
					Return(`{"history":[]}`, nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("get_execution_logs", map[string]interface{}{}))

				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("sync_executables tool", func() {
			It("should call executor with correct arguments", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "sync").
					Return("Synced executables", nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("sync_executables", nil))

				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("structured error responses", func() {
			It("should return a structured error JSON when executor fails", func() {
				mockExecutor.EXPECT().
					Execute("workspace", "get", "missing-ws", "--output", "json").
					Return("workspace not found", errors.New("exit status 1"))

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_workspace", map[string]interface{}{
					"workspace_name": "missing-ws",
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())

				text := getTextContent(result)
				var payload struct {
					Error struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				Expect(json.Unmarshal([]byte(text), &payload)).To(Succeed())
				Expect(payload.Error.Code).To(Equal("NOT_FOUND"))
				Expect(payload.Error.Message).To(ContainSubstring("missing-ws"))
			})

			It("should return INVALID_INPUT error for missing required parameter", func() {
				result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_workspace", map[string]interface{}{}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())

				text := getTextContent(result)
				Expect(text).To(ContainSubstring("INVALID_INPUT"))
				Expect(text).To(ContainSubstring("workspace_name"))
			})
		})

		Context("write_flowfile tool", func() {
			It("should write a valid flowfile and return summary", func() {
				tmpDir := GinkgoTB().TempDir()
				flowPath := filepath.Join(tmpDir, "test.flow")

				validYAML := `
executables:
  - verb: run
    name: hello
    description: say hello
    exec:
      cmd: echo hello
  - verb: test
    name: greet
    exec:
      cmd: echo greet
`
				mockExecutor.EXPECT().ExecuteContext(gomock.Any(), "sync").Return("synced", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("write_flowfile", map[string]interface{}{
					"path":    flowPath,
					"content": validYAML,
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeFalse())

				text := getTextContent(result)
				var out flowMcp.WriteFlowFileOutput
				Expect(json.Unmarshal([]byte(text), &out)).To(Succeed())
				Expect(out.Path).To(Equal(flowPath))
				Expect(out.Executables).To(ContainElements("hello", "greet"))
				Expect(out.SyncFailed).To(BeFalse())

				// Verify file was actually written
				_, statErr := os.Stat(flowPath)
				Expect(statErr).ToNot(HaveOccurred())
			})

			It("should report SyncFailed when the post-write cache refresh fails, without failing the write", func() {
				tmpDir := GinkgoTB().TempDir()
				flowPath := filepath.Join(tmpDir, "sync-fail.flow")

				mockExecutor.EXPECT().ExecuteContext(gomock.Any(), "sync").Return("", errors.New("sync exploded"))

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("write_flowfile", map[string]interface{}{
					"path":    flowPath,
					"content": "executables: []",
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeFalse())

				var out flowMcp.WriteFlowFileOutput
				Expect(json.Unmarshal([]byte(getTextContent(result)), &out)).To(Succeed())
				Expect(out.SyncFailed).To(BeTrue())

				// The write itself still succeeded despite the sync failure.
				_, statErr := os.Stat(flowPath)
				Expect(statErr).ToNot(HaveOccurred())
			})

			It("should reject invalid file extension", func() {
				result, err := mcpClient.CallTool(ctx, newCallToolRequest("write_flowfile", map[string]interface{}{
					"path":    "/tmp/bad.txt",
					"content": "executables: []",
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
				Expect(getTextContent(result)).To(ContainSubstring("VALIDATION_FAILED"))
			})

			It("should reject invalid YAML content", func() {
				tmpDir := GinkgoTB().TempDir()
				flowPath := filepath.Join(tmpDir, "bad.flow")

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("write_flowfile", map[string]interface{}{
					"path":    flowPath,
					"content": "not: valid: yaml: [[[",
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
				Expect(getTextContent(result)).To(ContainSubstring("VALIDATION_FAILED"))
			})

			It("should reject existing file without overwrite flag", func() {
				tmpDir := GinkgoTB().TempDir()
				flowPath := filepath.Join(tmpDir, "existing.flow")
				Expect(os.WriteFile(flowPath, []byte("executables: []\n"), 0600)).To(Succeed())

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("write_flowfile", map[string]interface{}{
					"path":    flowPath,
					"content": "executables: []",
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
				Expect(getTextContent(result)).To(ContainSubstring("already exists"))
			})

			It("should overwrite existing file when overwrite=true", func() {
				tmpDir := GinkgoTB().TempDir()
				flowPath := filepath.Join(tmpDir, "existing.flow")
				Expect(os.WriteFile(flowPath, []byte("executables: []\n"), 0600)).To(Succeed())

				mockExecutor.EXPECT().ExecuteContext(gomock.Any(), "sync").Return("synced", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("write_flowfile", map[string]interface{}{
					"path":      flowPath,
					"content":   "executables: []",
					"overwrite": true,
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeFalse())
			})
		})

		Context("list_executables pagination", func() {
			It("should paginate results and include next cursor", func() {
				// Build a CLI JSON output with 30 executables.
				var cliResp struct {
					Executables []flowMcp.ExecutableOutput `json:"executables"`
				}
				for i := 0; i < 30; i++ {
					cliResp.Executables = append(cliResp.Executables, flowMcp.ExecutableOutput{
						ID:   "ws/ns:exec",
						Ref:  "run ws/ns:exec",
						Name: "exec",
						Verb: "run",
					})
				}
				cliJSON, _ := json.Marshal(cliResp)

				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "browse", "--output", "json", "--workspace", "*", "--namespace", "*").
					Return(string(cliJSON), nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("list_executables", nil))
				Expect(err).ToNot(HaveOccurred())

				var out flowMcp.ExecutableListOutput
				Expect(json.Unmarshal([]byte(getTextContent(result)), &out)).To(Succeed())
				Expect(out.TotalCount).To(Equal(30))
				Expect(out.Executables).To(HaveLen(25)) // defaultPageSize
				Expect(out.NextCursor).ToNot(BeEmpty())
			})
		})
	})
})

// Helper function to create a CallToolRequest
func newCallToolRequest(name string, args map[string]interface{}) mcp.CallToolRequest {
	if args == nil {
		args = make(map[string]interface{})
	}
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
}

// Helper function to extract text content from mcp.CallToolResult
func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}
	GinkgoTB().Fatalf("Expected text content, got %T", result.Content[0])
	return ""
}
