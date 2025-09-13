package main

import (
	"context"
	"log"

	"github.com/Namita-Singroha/mcp-gitreview/internal/server"
	"github.com/Namita-Singroha/mcp-gitreview/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	serverCtx := server.NewContext()

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "gitreview"}, nil)

	// Register all tools
	tools.RegisterAll(mcpServer, serverCtx)

	log.Printf("DEBUG: Starting MCP server...")
	if err := mcpServer.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server failed: %v", err)
	}
}
