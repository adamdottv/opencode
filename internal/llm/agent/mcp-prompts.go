package agent

import (
	"github.com/sst/opencode/internal/config"
)

// MCPPrompt represents a prompt from an MCP server
type MCPPrompt struct {
	Name         string
	Description  string
	Arguments    []MCPPromptArgument
	ServerName   string
	ServerConfig config.MCPServer
}

// MCPPromptArgument represents an argument for an MCP prompt
type MCPPromptArgument struct {
	Name        string
	Description string
	Required    bool
}
