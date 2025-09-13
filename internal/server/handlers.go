package server

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ErrorResponse creates a consistent error response
func ErrorResponse(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}

// SuccessResponse creates a consistent success response
func SuccessResponse(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}