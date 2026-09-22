package mcp_test

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	flowMcp "github.com/flowexec/flow/v2/internal/mcp"
	"github.com/flowexec/flow/v2/internal/mcp/mocks"
)

var _ = Describe("MCP Server Extensions", func() {
	var mockExecutor *mocks.MockCommandExecutor

	BeforeEach(func() {
		mockExecutor = mocks.NewMockCommandExecutor(gomock.NewController(GinkgoT()))
	})

	echoTool := func(name string) server.ServerTool {
		return server.ServerTool{
			Tool: mcp.NewTool(name, mcp.WithDescription("echo")),
			Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return mcp.NewToolResultText("pong"), nil
			},
		}
	}

	It("registers extension tools alongside the built-ins and appends instructions", func() {
		flowServer, err := flowMcp.NewServer(mockExecutor, flowMcp.WithExtensions(flowMcp.Extension{
			Name:         "ext",
			Tools:        []server.ServerTool{echoTool("ext_ping")},
			Instructions: "Use ext_ping to check liveness.",
		}))
		Expect(err).ToNot(HaveOccurred())

		c, err := client.NewInProcessClient(flowServer.GetMCPServer())
		Expect(err).ToNot(HaveOccurred())
		initRes, err := c.Initialize(context.Background(), mcp.InitializeRequest{
			Params: mcp.InitializeParams{ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(initRes.Instructions).To(ContainSubstring("Flow MCP Server Instructions"))
		Expect(initRes.Instructions).To(HaveSuffix("Use ext_ping to check liveness."))

		tools := flowServer.GetMCPServer().ListTools()
		Expect(tools).To(HaveKey("ext_ping"))
		Expect(tools).To(HaveKey("execute"))

		req := mcp.CallToolRequest{}
		req.Params.Name = "ext_ping"
		res, err := c.CallTool(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())
		text, ok := res.Content[0].(mcp.TextContent)
		Expect(ok).To(BeTrue())
		Expect(text.Text).To(Equal("pong"))
	})

	It("applies extension server options to built-in tools", func() {
		var called []string
		mw := func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
			return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				called = append(called, req.Params.Name)
				return next(ctx, req)
			}
		}
		flowServer, err := flowMcp.NewServer(mockExecutor, flowMcp.WithExtensions(flowMcp.Extension{
			ServerOptions: []server.ServerOption{server.WithToolHandlerMiddleware(mw)},
		}))
		Expect(err).ToNot(HaveOccurred())

		c, err := client.NewInProcessClient(flowServer.GetMCPServer())
		Expect(err).ToNot(HaveOccurred())
		_, err = c.Initialize(context.Background(), mcp.InitializeRequest{
			Params: mcp.InitializeParams{ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION},
		})
		Expect(err).ToNot(HaveOccurred())

		req := mcp.CallToolRequest{}
		req.Params.Name = "get_info"
		_, _ = c.CallTool(context.Background(), req)
		Expect(called).To(ConsistOf("get_info"))
	})

	It("reports flow's server info by default and an override when given", func() {
		serverName := func(opts ...flowMcp.Option) (string, string) {
			flowServer, err := flowMcp.NewServer(mockExecutor, opts...)
			Expect(err).ToNot(HaveOccurred())
			c, err := client.NewInProcessClient(flowServer.GetMCPServer())
			Expect(err).ToNot(HaveOccurred())
			res, err := c.Initialize(context.Background(), mcp.InitializeRequest{
				Params: mcp.InitializeParams{ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION},
			})
			Expect(err).ToNot(HaveOccurred())
			return res.ServerInfo.Name, res.ServerInfo.Version
		}

		name, version := serverName()
		Expect(name).To(Equal("Flow"))
		Expect(version).To(Equal("1.0.0"))

		name, version = serverName(flowMcp.WithServerInfo("Mine", "2.3.4"))
		Expect(name).To(Equal("Mine"))
		Expect(version).To(Equal("2.3.4"))

		name, version = serverName(flowMcp.WithServerInfo("Mine", ""))
		Expect(name).To(Equal("Mine"))
		Expect(version).To(Equal("1.0.0"))
	})

	DescribeTable("rejects invalid or colliding registrations",
		func(exts []flowMcp.Extension, msg string) {
			_, err := flowMcp.NewServer(mockExecutor, flowMcp.WithExtensions(exts...))
			Expect(err).To(MatchError(ContainSubstring(msg)))
		},
		Entry("built-in tool name",
			[]flowMcp.Extension{{Name: "ext", Tools: []server.ServerTool{echoTool("execute")}}},
			`mcp extension ext: tool "execute" is already registered`),
		Entry("tool name across extensions",
			[]flowMcp.Extension{
				{Name: "a", Tools: []server.ServerTool{echoTool("dup")}},
				{Name: "b", Tools: []server.ServerTool{echoTool("dup")}},
			},
			`mcp extension b: tool "dup" is already registered`),
		Entry("tool without handler",
			[]flowMcp.Extension{{Tools: []server.ServerTool{{Tool: mcp.NewTool("nohandler")}}}},
			`mcp extension #0: tool "nohandler" has no handler`),
		Entry("built-in prompt name",
			[]flowMcp.Extension{{Name: "ext", Prompts: []server.ServerPrompt{{
				Prompt: mcp.NewPrompt("explain_flow"),
				Handler: func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
					return &mcp.GetPromptResult{}, nil
				},
			}}}},
			`prompt "explain_flow" is already registered`),
		Entry("reserved resource template scheme",
			[]flowMcp.Extension{{Name: "ext", ResourceTemplates: []server.ServerResourceTemplate{{
				Template: mcp.NewResourceTemplate("flow://workspace/{name}", "ws"),
				Handler: func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
					return nil, nil
				},
			}}}},
			`uses the reserved flow:// scheme`),
	)
})
