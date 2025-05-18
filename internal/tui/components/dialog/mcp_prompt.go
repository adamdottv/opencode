package dialog

import (
	"github.com/sst/opencode/internal/llm/agent"
)

// MCPPromptRunMsg is sent when an MCP prompt is executed
type MCPPromptRunMsg struct {
	Prompt agent.MCPPrompt
	Args   map[string]string
}