package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sst/opencode/internal/config"
	"github.com/sst/opencode/internal/llm/tools"
	"github.com/sst/opencode/internal/permission"
	"github.com/sst/opencode/internal/version"
)

// Global variables to store MCP resources
var (
	globalMCPTools []tools.BaseTool
	mcpPrompts     []MCPPrompt
)

// GetMCPResources fetches both tools and prompts from all MCP servers
func GetMCPResources(ctx context.Context, permissions permission.Service) ([]tools.BaseTool, []MCPPrompt) {
	// If already loaded, return cached values
	if len(globalMCPTools) > 0 && len(mcpPrompts) > 0 {
		return globalMCPTools, mcpPrompts
	}

	// Clear existing resources
	globalMCPTools = []tools.BaseTool{}
	mcpPrompts = []MCPPrompt{}

	// Loop through all configured MCP servers
	for serverName, serverConfig := range config.Get().MCPServers {
		// Create a client for this server
		c, err := createMCPClient(ctx, serverConfig)
		if err != nil {
			slog.Error("error creating MCP client", 
				"server", serverName, 
				"error", err)
			continue
		}

		// Get tools from this server
		serverTools, err := fetchToolsFromClient(ctx, serverName, serverConfig, permissions, c)
		if err != nil {
			slog.Error("error fetching tools from MCP server",
				"server", serverName,
				"error", err)
		} else {
			globalMCPTools = append(globalMCPTools, serverTools...)
		}

		// Get prompts from this server
		serverPrompts, err := fetchPromptsFromClient(ctx, serverName, serverConfig, c)
		if err != nil {
			slog.Error("error fetching prompts from MCP server",
				"server", serverName,
				"error", err)
		} else {
			mcpPrompts = append(mcpPrompts, serverPrompts...)
		}

		// Close the client
		c.Close()
	}

	return globalMCPTools, mcpPrompts
}

// GetMcpTools returns all MCP tools
func GetMcpTools(ctx context.Context, permissions permission.Service) []tools.BaseTool {
	tools, _ := GetMCPResources(ctx, permissions)
	return tools
}

// GetMCPPrompts returns all MCP prompts
func GetMCPPrompts(ctx context.Context) []MCPPrompt {
	_, prompts := GetMCPResources(ctx, nil)
	return prompts
}

// createMCPClient creates and initializes an MCP client for a server
func createMCPClient(ctx context.Context, serverConfig config.MCPServer) (MCPClient, error) {
	var c MCPClient
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

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "OpenCode",
		Version: version.Version,
	}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("error initializing MCP client: %w", err)
	}

	return c, nil
}

// fetchToolsFromClient fetches tools using an existing MCP client
func fetchToolsFromClient(ctx context.Context, serverName string, serverConfig config.MCPServer, permissions permission.Service, c MCPClient) ([]tools.BaseTool, error) {
	var serverTools []tools.BaseTool

	// List tools
	toolsRequest := mcp.ListToolsRequest{}
	toolsResponse, err := c.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, fmt.Errorf("error listing tools: %w", err)
	}

	// Create tool wrappers
	for _, t := range toolsResponse.Tools {
		serverTools = append(serverTools, NewMcpTool(serverName, t, permissions, serverConfig))
	}

	return serverTools, nil
}

// fetchPromptsFromClient fetches prompts using an existing MCP client
func fetchPromptsFromClient(ctx context.Context, serverName string, serverConfig config.MCPServer, c MCPClient) ([]MCPPrompt, error) {
	var serverPrompts []MCPPrompt

	// List prompts
	promptsRequest := mcp.ListPromptsRequest{}
	promptsResponse, err := c.ListPrompts(ctx, promptsRequest)
	if err != nil {
		return nil, fmt.Errorf("error listing prompts: %w", err)
	}

	// Create prompt wrappers
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

		serverPrompts = append(serverPrompts, mcpPrompt)
	}

	return serverPrompts, nil
}

// ExecutePrompt executes a prompt on an MCP server
func ExecutePrompt(ctx context.Context, prompt MCPPrompt, args map[string]string) ([]mcp.PromptMessage, error) {
	// Create a client for this server
	c, err := createMCPClient(ctx, prompt.ServerConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating MCP client: %w", err)
	}
	defer c.Close()

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