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
		var err error

		flowServer, err = flowMcp.NewServer(mockExecutor)
		Expect(err).ToNot(HaveOccurred())

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
				"run_python",
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
				"run_python":     true,
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
				"flow://executable/{workspace}/{+namespace}/{name}",
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

			It("should leave full descriptions to get_workspace", func() {
				mockExecutor.EXPECT().
					Execute("workspace", "list", "--output", "json").
					Return(`{"workspaces":[{"name":"a","path":"/a","description":"short",`+
						`"fullDescription":"very long readme"}]}`, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("list_workspaces", nil))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring(`"description":"short"`))
				Expect(getTextContent(result)).ToNot(ContainSubstring("fullDescription"))
				Expect(getTextContent(result)).ToNot(ContainSubstring("very long readme"))
			})

			It("should keep the full description on get_workspace", func() {
				mockExecutor.EXPECT().
					Execute("workspace", "get", "a", "--output", "json").
					Return(`{"name":"a","path":"/a","fullDescription":"very long readme"}`, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_workspace", map[string]interface{}{
					"workspace_name": "a",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("very long readme"))
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

			It("should leave full descriptions to get_executable", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "browse", "--output", "json", "--workspace", "*", "--namespace", "*").
					Return(`{"executables":[{"id":"a","ref":"run a","name":"a","verb":"run","description":"short",`+
						`"fullDescription":"very long markdown"}]}`, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("list_executables", nil))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring(`"description":"short"`))
				Expect(getTextContent(result)).ToNot(ContainSubstring("fullDescription"))
				Expect(getTextContent(result)).ToNot(ContainSubstring("very long markdown"))
			})
		})

		Context("execute tool", func() {
			It("should call executor with provided arguments", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "test", "test:test-flow", "--log-mode", "text", "--", "arg1", "arg2").
					Return("execution result", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
					"args":            "arg1 arg2",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("execution result"))
			})

			DescribeTable("should keep args away from flow's own flags",
				func(args string, cliArgs ...any) {
					want := append([]any{"test", "test:test-flow", "--workspace", "ws", "--sync", "--log-mode", "text"}, cliArgs...)
					mockExecutor.EXPECT().ExecuteContext(gomock.Any(), want...).Return("ok", nil)

					result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
						"executable_verb": "test",
						"executable_id":   "test:test-flow",
						"workspace":       "ws",
						"sync":            true,
						"args":            args,
					}))
					Expect(err).ToNot(HaveOccurred())
					Expect(result.IsError).To(BeFalse())
				},
				Entry("flag without a separator", "--node workbox", "--", "--node", "workbox"),
				Entry("flag with a separator", "-- --node workbox", "--", "--node", "workbox"),
				Entry("quoted value", "devbox 'uname -a'", "--", "devbox", "uname -a"),
				Entry("flow's --param", "--param K=v", "--", "--param", "K=v"),
				Entry("flow's --background", "--background", "--", "--background"),
			)

			It("should forward params as sorted --param flags ahead of the args", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "test", "test:test-flow",
						"--param", "A=1=2", "--param", "B=two words", "--log-mode", "text", "--", "pos").
					Return("ok", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
					"args":            "pos",
					"params":          map[string]interface{}{"B": "two words", "A": "1=2"},
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeFalse())
			})

			DescribeTable("should reject invalid params without running anything",
				func(params interface{}) {
					result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
						"executable_verb": "test",
						"executable_id":   "test:test-flow",
						"params":          params,
					}))
					Expect(err).ToNot(HaveOccurred())
					Expect(result.IsError).To(BeTrue())
					Expect(getTextContent(result)).To(ContainSubstring(flowMcp.ErrCodeInvalidInput))
				},
				Entry("not a map", "K=v"),
				Entry("non-string value", map[string]interface{}{"K": 1}),
				Entry("key with an equals sign", map[string]interface{}{"K=X": "v"}),
				Entry("empty key", map[string]interface{}{"": "v"}),
			)

			It("should target the nameless executable when args come without an ID", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "test", ":", "--log-mode", "text", "--", "--node", "workbox").
					Return("ok", nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"args":            "--node workbox",
				}))
				Expect(err).ToNot(HaveOccurred())
			})

			It("should reject args with an unterminated quote without running anything", func() {
				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
					"args":            "don't",
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
				Expect(getTextContent(result)).To(ContainSubstring(flowMcp.ErrCodeInvalidInput))
			})

			It("should handle no args", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "test", "test:test-flow", "--log-mode", "text").
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
					ExecuteContext(gomock.Any(), "test", "test:test-flow", "--log-mode", "text").
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

			It("should report exit code 0 on success", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "test", "test:test-flow", "--log-mode", "text").
					Return("ok", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeFalse())
				Expect(getTextContent(result)).To(ContainSubstring(`"exitCode":0`))
			})
		})

		DescribeTable("failed runs",
			func(tool string, args map[string]interface{}, cliArgs []any) {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), cliArgs...).
					Return("step output\nboom\n", &flowMcp.CLIExitError{Code: 2})

				result, err := mcpClient.CallTool(ctx, newCallToolRequest(tool, args))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
				Expect(getTextContent(result)).To(ContainSubstring("boom"))
				Expect(getTextContent(result)).To(ContainSubstring(flowMcp.ErrCodeExecutionFailed))

				structured, marshalErr := json.Marshal(result.StructuredContent)
				Expect(marshalErr).ToNot(HaveOccurred())
				var out flowMcp.ExecutionOutput
				Expect(json.Unmarshal(structured, &out)).To(Succeed())
				Expect(out.ExitCode).To(Equal(2))
				Expect(out.Output).To(Equal("step output\nboom\n"))
			},
			Entry("execute", "execute",
				map[string]interface{}{"executable_verb": "test", "executable_id": "test:flow"},
				[]any{"test", "test:flow", "--log-mode", "text"}),
			Entry("run_command", "run_command",
				map[string]interface{}{"command": "false"},
				[]any{"exec", "--cmd", "false", "--log-mode", "text"}),
			Entry("run_python", "run_python",
				map[string]interface{}{"code": "raise SystemExit(1)"},
				[]any{"exec", "--interpreter", "python", "--cmd", "raise SystemExit(1)", "--log-mode", "text"}),
			Entry("run_executable", "run_executable",
				map[string]interface{}{"spec": "verb: run"},
				[]any{"exec", "--spec", "verb: run", "--log-mode", "text"}),
		)

		It("should report -1 when flow could not be started", func() {
			mockExecutor.EXPECT().
				ExecuteContext(gomock.Any(), "exec", "--cmd", "true", "--log-mode", "text").
				Return("", errors.New("flow not found"))

			result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_command", map[string]interface{}{
				"command": "true",
			}))
			Expect(err).ToNot(HaveOccurred())
			Expect(result.IsError).To(BeTrue())
			Expect(getTextContent(result)).To(ContainSubstring(`"exitCode":-1`))
		})

		It("should report a missing workspace as an error", func() {
			mockExecutor.EXPECT().
				Execute("workspace", "get", "missing", "--output", "json").
				Return(`{"error":{"code":"NOT_FOUND"}}`, &flowMcp.CLIExitError{Code: 1})

			result, err := mcpClient.CallTool(ctx, newCallToolRequest("get_workspace", map[string]interface{}{
				"workspace_name": "missing",
			}))
			Expect(err).ToNot(HaveOccurred())
			Expect(result.IsError).To(BeTrue())
		})

		Context("run_command tool", func() {
			It("should run an ad-hoc command via flow exec --cmd with label and dir", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--cmd", "echo hi", "--label", "greet", "--dir", "/tmp", "--log-mode", "text").
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
					ExecuteContext(gomock.Any(), "exec", "--cmd", "echo a", "--cmd", "echo b", "--mode", "parallel", "--log-mode", "text").
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
					ExecuteContext(gomock.Any(), "exec", "--cmd", "echo hi", "--log-mode", "text").
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

		Context("run_python tool", func() {
			It("should run python via flow exec --interpreter python --cmd", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--interpreter", "python", "--cmd", "print(1)",
						"--label", "compute", "--dir", "/tmp", "--log-mode", "text").
					Return("1", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_python", map[string]interface{}{
					"code":  "print(1)",
					"label": "compute",
					"dir":   "/tmp",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("1"))
			})

			It("should forward workspace and sync", func() {
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--interpreter", "python", "--cmd", "print(2)",
						"--workspace", "other", "--sync", "--log-mode", "text").
					Return("2", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_python", map[string]interface{}{
					"code":      "print(2)",
					"workspace": "other",
					"sync":      true,
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("2"))
			})

			It("should preserve multi-line code as a single argument", func() {
				// Line structure has to survive intact or tracebacks point at the
				// wrong line.
				code := "import sys\nif True:\n    print(sys.version)\n"
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--interpreter", "python", "--cmd", code, "--log-mode", "text").
					Return("ok", nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_python", map[string]interface{}{
					"code": code,
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(ContainSubstring("ok"))
			})

			It("should require code", func() {
				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_python", map[string]interface{}{}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
			})

			It("should reject empty code rather than running an empty script", func() {
				result, err := mcpClient.CallTool(ctx, newCallToolRequest("run_python", map[string]interface{}{
					"code": "",
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(result.IsError).To(BeTrue())
			})
		})

		Context("run_executable tool", func() {
			It("should run a transient executable from an inline spec", func() {
				spec := `{"verb":"run","serial":{"execs":[{"cmd":"echo one"}]}}`
				mockExecutor.EXPECT().
					ExecuteContext(gomock.Any(), "exec", "--spec", spec, "--label", "batch", "--log-mode", "text").
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

			It("should look up a single run by its id, with its output", func() {
				mockExecutor.EXPECT().
					Execute("logs", "--output", "json", "--id", "run-1", "--content", "--max-bytes", "50000", "--tail", "40").
					Return(`{"history":[]}`, nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("get_execution_logs", map[string]interface{}{
					"id":   "run-1",
					"tail": 40,
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
