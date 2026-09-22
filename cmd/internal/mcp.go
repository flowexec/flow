package internal

import (
	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/internal/mcp"
	"github.com/flowexec/flow/v2/pkg/context"
)

func RegisterMCPCmd(ctx *context.Context, rootCmd *cobra.Command, opts ...mcp.Option) {
	subCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start Model Context Provider (MCP) server for AI assistant integration",
		Long: "Start a Model Context Protocol server that enables AI assistants to interact with your flow executables, " +
			"workspaces, and configurations through natural language. AI assistants can discover, validate, and execute " +
			"flow workflows, making your automation platform accessible through conversational interfaces/clients.\n\n" +
			"This server used stdio for transport. For more information on MCP, see https://modelcontextprotocol.io",
		Args: cobra.NoArgs,
		Run:  func(cmd *cobra.Command, args []string) { mcpFunc(ctx, cmd, opts) },
	}
	rootCmd.AddCommand(subCmd)
}

func mcpFunc(ctx *context.Context, cmd *cobra.Command, opts []mcp.Option) {
	server, err := mcp.NewServer(&mcp.FlowCLIExecutor{}, opts...)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}
	if err := server.Run(); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
}
