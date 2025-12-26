package internal

import (
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/internal/mcp"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
)

func RegisterMCPCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start Model Context Provider (MCP) server for AI assistant integration",
		Long: "Start a Model Context Protocol server that enables AI assistants to interact with your flow executables, " +
			"workspaces, and configurations through natural language. AI assistants can discover, validate, and execute " +
			"flow workflows, making your automation platform accessible through conversational interfaces/clients.\n\n" +
			"This server used stdio for transport. For more information on MCP, see https://modelcontextprotocol.io",
		Args: cobra.NoArgs,
		Run:  func(cmd *cobra.Command, args []string) { mcpFunc(ctx, cmd, args) },
	}
	rootCmd.AddCommand(subCmd)
}

func mcpFunc(_ *context.Context, _ *cobra.Command, _ []string) {
	server := mcp.NewServer(&mcp.FlowCLIExecutor{})
	if err := server.Run(); err != nil {
		logger.Log().FatalErr(err)
	}
}
