package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"log/slog"

	charmlog "github.com/charmbracelet/log"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/config"
	"github.com/sst/opencode/internal/db"
	"github.com/sst/opencode/internal/format"
	"github.com/sst/opencode/internal/llm/agent"
	"github.com/sst/opencode/internal/llm/tools"
	"github.com/sst/opencode/internal/message"
	"github.com/sst/opencode/internal/permission"
	"github.com/sst/opencode/internal/tui/components/spinner"
	"github.com/sst/opencode/internal/tui/theme"
)

// syncWriter is a thread-safe writer that prevents interleaved output
type syncWriter struct {
	w  io.Writer
	mu sync.Mutex
}

// Write implements io.Writer
func (sw *syncWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}

// newSyncWriter creates a new synchronized writer
func newSyncWriter(w io.Writer) io.Writer {
	return &syncWriter{w: w}
}

// filterTools filters the provided tools based on allowed or excluded tool names
func filterTools(allTools []tools.BaseTool, allowedTools, excludedTools []string) []tools.BaseTool {
	// If neither allowed nor excluded tools are specified, return all tools
	if len(allowedTools) == 0 && len(excludedTools) == 0 {
		return allTools
	}

	// Create a map for faster lookups
	allowedMap := make(map[string]bool)
	for _, name := range allowedTools {
		allowedMap[name] = true
	}

	excludedMap := make(map[string]bool)
	for _, name := range excludedTools {
		excludedMap[name] = true
	}

	var filteredTools []tools.BaseTool

	for _, tool := range allTools {
		toolName := tool.Info().Name

		// If we have an allowed list, only include tools in that list
		if len(allowedTools) > 0 {
			if allowedMap[toolName] {
				filteredTools = append(filteredTools, tool)
			}
		} else if len(excludedTools) > 0 {
			// If we have an excluded list, include all tools except those in the list
			if !excludedMap[toolName] {
				filteredTools = append(filteredTools, tool)
			}
		}
	}

	return filteredTools
}

// toolWrapper wraps tools.BaseTool to implement permission.Tool interface
type toolWrapper struct {
	tool tools.BaseTool
}

func (tw *toolWrapper) Info() permission.ToolInfo {
	info := tw.tool.Info()
	return permission.ToolInfo{
		Name:        info.Name,
		Description: info.Description,
		Parameters:  info.Parameters,
		Required:    info.Required,
	}
}

func (tw *toolWrapper) Run(ctx context.Context, params permission.ToolCall) (permission.ToolResponse, error) {
	toolsParams := tools.ToolCall{
		ID:    params.ID,
		Name:  params.Name,
		Input: params.Input,
	}
	
	result, err := tw.tool.Run(ctx, toolsParams)
	if err != nil {
		return permission.ToolResponse{}, err
	}
	
	return permission.ToolResponse{
		Type:     string(result.Type),
		Content:  result.Content,
		Metadata: result.Metadata,
		IsError:  result.IsError,
	}, nil
}

// findPermissionTool finds the specified permission prompt tool in the tools list
func findPermissionTool(allTools []tools.BaseTool, permissionToolName string) (tools.BaseTool, string, error) {
	// Parse the claude-code format mcp__{server}__{tool} to OpenCode format {server}_{tool}
	if !strings.HasPrefix(permissionToolName, "mcp__") {
		return nil, "", fmt.Errorf("invalid permission prompt tool format: %s (expected: mcp__{server}__{tool})", permissionToolName)
	}
	
	// Remove "mcp__" prefix and convert "__" to "_"
	parsed := strings.TrimPrefix(permissionToolName, "mcp__")
	openCodeToolName := strings.Replace(parsed, "__", "_", 1)
	
	// Find the permission tool
	var permissionTool tools.BaseTool
	var availableMCPTools []string
	
	for _, tool := range allTools {
		toolInfo := tool.Info()
		// Check if this is an MCP tool (contains underscore indicating server_tool format)
		if strings.Contains(toolInfo.Name, "_") {
			availableMCPTools = append(availableMCPTools, "mcp__"+strings.Replace(toolInfo.Name, "_", "__", 1))
			if toolInfo.Name == openCodeToolName {
				permissionTool = tool
			}
		}
	}
	
	if permissionTool == nil {
		if len(availableMCPTools) == 0 {
			return nil, "", fmt.Errorf("MCP tool %s (passed via --permission-prompt-tool) not found. Available MCP tools: none", permissionToolName)
		}
		return nil, "", fmt.Errorf("MCP tool %s (passed via --permission-prompt-tool) not found. Available MCP tools: %s", 
			permissionToolName, strings.Join(availableMCPTools, ", "))
	}
	
	slog.Info("Found permission prompt tool", "tool", permissionTool.Info().Name)
	return permissionTool, openCodeToolName, nil
}

// handleNonInteractiveMode processes a single prompt in non-interactive mode
func handleNonInteractiveMode(ctx context.Context, prompt string, outputFormat format.OutputFormat, quiet bool, verbose bool, allowedTools, excludedTools []string, permissionPromptTool string) error {
	// Initial log message using standard slog
	slog.Info("Running in non-interactive mode", "prompt", prompt, "format", outputFormat, "quiet", quiet, "verbose", verbose,
		"allowedTools", allowedTools, "excludedTools", excludedTools, "permissionPromptTool", permissionPromptTool)

	// Sanity check for mutually exclusive flags
	if quiet && verbose {
		return fmt.Errorf("--quiet and --verbose flags cannot be used together")
	}

	// Set up logging to stderr if verbose mode is enabled
	if verbose {
		// Create a synchronized writer to prevent interleaved output
		syncWriter := newSyncWriter(os.Stderr)

		// Create a charmbracelet/log logger that writes to the synchronized writer
		charmLogger := charmlog.NewWithOptions(syncWriter, charmlog.Options{
			Level:           charmlog.DebugLevel,
			ReportCaller:    true,
			ReportTimestamp: true,
			TimeFormat:      time.RFC3339,
			Prefix:          "OpenCode",
		})

		// Set the global logger for charmbracelet/log
		charmlog.SetDefault(charmLogger)

		// Create a slog handler that uses charmbracelet/log
		// This will forward all slog logs to charmbracelet/log
		slog.SetDefault(slog.New(charmLogger))

		// Log a message to confirm verbose logging is enabled
		charmLogger.Info("Verbose logging enabled")
	}

	// Start spinner if not in quiet mode
	var s *spinner.Spinner
	if !quiet {
		// Get the current theme to style the spinner
		currentTheme := theme.CurrentTheme()

		// Create a themed spinner
		if currentTheme != nil {
			// Use the primary color from the theme
			s = spinner.NewThemedSpinner("Thinking...", currentTheme.Primary())
		} else {
			// Fallback to default spinner if no theme is available
			s = spinner.NewSpinner("Thinking...")
		}

		s.Start()
		defer s.Stop()
	}

	// Connect DB, this will also run migrations
	conn, err := db.Connect()
	if err != nil {
		return err
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create the app
	app, err := app.New(ctx, conn)
	if err != nil {
		slog.Error("Failed to create app", "error", err)
		return err
	}

	// Create a new session for this prompt
	session, err := app.Sessions.Create(ctx, "Non-interactive prompt")
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Set the session as current
	app.CurrentSession = &session

	// Initialize MCP tools synchronously in non-interactive mode (if any are configured)
	mcpServers := config.Get().MCPServers
	if len(mcpServers) > 0 {
		mcpCtx, mcpCancel := context.WithTimeout(ctx, 10*time.Second)
		agent.GetMcpTools(mcpCtx, app.Permissions)
		mcpCancel()
	}

	// Get all tools including MCP tools
	allTools := agent.PrimaryAgentTools(
		app.Permissions,
		app.Sessions,
		app.Messages,
		app.History,
		app.LSPClients,
	)

	// Handle permission prompt tool setup
	if permissionPromptTool != "" {
		// Find the permission tool
		permissionTool, openCodeToolName, err := findPermissionTool(allTools, permissionPromptTool)
		if err != nil {
			return err
		}
		
		// Store the permission tool for this session (wrapped to match interface)
		permission.SetPermissionPromptTool(ctx, session.ID, &toolWrapper{tool: permissionTool})
		
		// Add permission tool to excluded tools so it gets filtered out of LLM tools
		excludedTools = append(excludedTools, openCodeToolName)
	} else {
		// Auto-approve all permissions for this session (current behavior)
		permission.AutoApproveSession(ctx, session.ID)
	}

	// Create the user message
	_, err = app.Messages.Create(ctx, session.ID, message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: prompt}},
	})
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// If tool restrictions are specified, create a new agent with filtered tools
	if len(allowedTools) > 0 || len(excludedTools) > 0 {
		// Filter tools based on allowed/excluded lists (permission tool automatically excluded if present)
		filteredTools := filterTools(allTools, allowedTools, excludedTools)

		// Log the filtered tools for debugging
		var toolNames []string
		for _, tool := range filteredTools {
			toolNames = append(toolNames, tool.Info().Name)
		}
		slog.Debug("Using filtered tools", "count", len(filteredTools), "tools", toolNames)

		// Create a new agent with the filtered tools
		restrictedAgent, err := agent.NewAgent(
			config.AgentPrimary,
			app.Sessions,
			app.Messages,
			filteredTools,
		)
		if err != nil {
			return fmt.Errorf("failed to create restricted agent: %w", err)
		}

		// Use the restricted agent for this request
		eventCh, err := restrictedAgent.Run(ctx, session.ID, prompt)
		if err != nil {
			return fmt.Errorf("failed to run restricted agent: %w", err)
		}

		// Wait for the response
		var response message.Message
		for event := range eventCh {
			if event.Err() != nil {
				return fmt.Errorf("agent error: %w", event.Err())
			}
			response = event.Response()
		}

		// Format and print the output
		content := ""
		if textContent := response.Content(); textContent != nil {
			content = textContent.Text
		}

		formattedOutput, err := format.FormatOutput(content, outputFormat)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		// Stop spinner before printing output
		if !quiet && s != nil {
			s.Stop()
		}

		// Print the formatted output to stdout
		fmt.Println(formattedOutput)

		// Shutdown the app
		app.Shutdown()

		return nil
	}

	// Run the default agent if no tool restrictions
	eventCh, err := app.PrimaryAgent.Run(ctx, session.ID, prompt)
	if err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	// Wait for the response
	var response message.Message
	for event := range eventCh {
		if event.Err() != nil {
			return fmt.Errorf("agent error: %w", event.Err())
		}
		response = event.Response()
	}

	// Get the text content from the response
	content := ""
	if textContent := response.Content(); textContent != nil {
		content = textContent.Text
	}

	// Format the output according to the specified format
	formattedOutput, err := format.FormatOutput(content, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Stop spinner before printing output
	if !quiet && s != nil {
		s.Stop()
	}

	// Print the formatted output to stdout
	fmt.Println(formattedOutput)

	// Shutdown the app
	app.Shutdown()

	return nil
}
