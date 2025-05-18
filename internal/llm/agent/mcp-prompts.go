package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sst/opencode/internal/config"
	"github.com/sst/opencode/internal/version"
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

// GetMCPPrompts fetches all prompts from all registered MCP servers
func GetMCPPrompts(ctx context.Context) []MCPPrompt {
	var prompts []MCPPrompt

	for serverName, serverConfig := range config.Get().MCPServers {
		serverPrompts, err := getPromptsFromServer(ctx, serverName, serverConfig)
		if err != nil {
			slog.Error("error fetching prompts from MCP server",
				"server", serverName,
				"error", err)
			continue
		}
		prompts = append(prompts, serverPrompts...)
	}

	return prompts
}

// getPromptsFromServer fetches prompts from a specific MCP server
func getPromptsFromServer(ctx context.Context, serverName string, serverConfig config.MCPServer) ([]MCPPrompt, error) {
	var c client.MCPClient
	var err error

	switch serverConfig.Type {
	case config.MCPStdio:
		c, err = client.NewStdioMCPClient(
			serverConfig.Command,
			serverConfig.Env,
			serverConfig.Args...,
		)
	case config.MCPSse:
		c, err = client.NewSSEMCPClient(
			serverConfig.URL,
			client.WithHeaders(serverConfig.Headers),
		)
	default:
		return nil, fmt.Errorf("unsupported MCP server type: %s", serverConfig.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating MCP client: %w", err)
	}
	defer c.Close()

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "OpenCode",
		Version: version.Version,
	}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("error initializing MCP client: %w", err)
	}

	// List prompts
	promptsRequest := mcp.ListPromptsRequest{}
	promptsResponse, err := c.ListPrompts(ctx, promptsRequest)
	if err != nil {
		return nil, fmt.Errorf("error listing prompts: %w", err)
	}

	var result []MCPPrompt
	for _, prompt := range promptsResponse.Prompts {
		mcpPrompt := MCPPrompt{
			Name:         prompt.Name,
			Description:  prompt.Description,
			ServerName:   serverName,
			ServerConfig: serverConfig,
		}

		for _, arg := range prompt.Arguments {
			mcpPrompt.Arguments = append(mcpPrompt.Arguments, MCPPromptArgument{
				Name:        arg.Name,
				Description: arg.Description,
				Required:    arg.Required,
			})
		}

		result = append(result, mcpPrompt)
	}

	return result, nil
}

// ExecutePrompt executes a prompt on an MCP server
func ExecutePrompt(ctx context.Context, prompt MCPPrompt, args map[string]string) ([]mcp.PromptMessage, error) {
	var c client.MCPClient
	var err error

	switch prompt.ServerConfig.Type {
	case config.MCPStdio:
		c, err = client.NewStdioMCPClient(
			prompt.ServerConfig.Command,
			prompt.ServerConfig.Env,
			prompt.ServerConfig.Args...,
		)
	case config.MCPSse:
		c, err = client.NewSSEMCPClient(
			prompt.ServerConfig.URL,
			client.WithHeaders(prompt.ServerConfig.Headers),
		)
	default:
		return nil, fmt.Errorf("unsupported MCP server type: %s", prompt.ServerConfig.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating MCP client: %w", err)
	}
	defer c.Close()

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "OpenCode",
		Version: version.Version,
	}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("error initializing MCP client: %w", err)
	}

	// Convert string args to any map
	promptArgs := make(map[string]any)
	for k, v := range args {
		promptArgs[k] = v
	}

	// Get prompt
	promptRequest := mcp.GetPromptRequest{}
	promptRequest.Params.Name = prompt.Name
	promptRequest.Params.Arguments = args

	promptResponse, err := c.GetPrompt(ctx, promptRequest)
	if err != nil {
		return nil, fmt.Errorf("error getting prompt: %w", err)
	}

	// Return the full array of messages
	return promptResponse.Messages, nil
}
