package mcp_test

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/filesystem"
	flowMcp "github.com/flowexec/flow/internal/mcp"
	"github.com/flowexec/flow/internal/mcp/mocks"
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
				"get_execution_logs",
				"sync_executables",
			}

			for _, expectedTool := range expectedTools {
				Expect(toolNames).To(ContainElement(expectedTool))
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
				Expect(content).To(ContainSubstring("usageGuides"))
				Expect(content).To(ContainSubstring("schemas"))
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
					Execute("browse", "--output", "json", "--workspace", "*", "--namespace", "*").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("list_executables", nil))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
			})
		})

		Context("execute tool", func() {
			It("should call executor with provided arguments", func() {
				expectedOutput := "execution result"
				mockExecutor.EXPECT().
					Execute("test", "test:test-flow", "arg1", "arg2").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
					"args":            "arg1 arg2",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
			})

			It("should handle no args", func() {
				expectedOutput := "execution result with no args"
				mockExecutor.EXPECT().
					Execute("test", "test:test-flow").
					Return(expectedOutput, nil)

				result, err := mcpClient.CallTool(ctx, newCallToolRequest("execute", map[string]interface{}{
					"executable_verb": "test",
					"executable_id":   "test:test-flow",
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(getTextContent(result)).To(Equal(expectedOutput))
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
		})

		Context("sync_executables tool", func() {
			It("should call executor with correct arguments", func() {
				mockExecutor.EXPECT().
					Execute("sync").
					Return("Synced executables", nil)

				_, err := mcpClient.CallTool(ctx, newCallToolRequest("sync_executables", nil))

				Expect(err).ToNot(HaveOccurred())
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
