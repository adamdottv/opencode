package permission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"log/slog"

	"github.com/google/uuid"
	"github.com/sst/opencode/internal/config"
	"github.com/sst/opencode/internal/pubsub"
)

var ErrorPermissionDenied = errors.New("permission denied")

// Tool represents a tool that can be executed (avoiding import cycle with tools package)
type Tool interface {
	Info() ToolInfo
	Run(ctx context.Context, params ToolCall) (ToolResponse, error)
}

// ToolInfo represents tool information
type ToolInfo struct {
	Name        string
	Description string
	Parameters  map[string]any
	Required    []string
}

// ToolCall represents a tool call
type ToolCall struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Input string `json:"input"`
}

// ToolResponse represents a tool response
type ToolResponse struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	Metadata string `json:"metadata,omitempty"`
	IsError  bool   `json:"is_error"`
}

type CreatePermissionRequest struct {
	SessionID   string `json:"session_id"`
	ToolName    string `json:"tool_name"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Params      any    `json:"params"`
	Path        string `json:"path"`
}

type PermissionRequest struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	ToolName    string `json:"tool_name"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Params      any    `json:"params"`
	Path        string `json:"path"`
}

type PermissionResponse struct {
	Request PermissionRequest
	Granted bool
}

const (
	EventPermissionRequested pubsub.EventType = "permission_requested"
	EventPermissionGranted   pubsub.EventType = "permission_granted"
	EventPermissionDenied    pubsub.EventType = "permission_denied"
	EventPermissionPersisted pubsub.EventType = "permission_persisted"
)

type Service interface {
	pubsub.Subscriber[PermissionRequest]
	SubscribeToResponseEvents(ctx context.Context) <-chan pubsub.Event[PermissionResponse]

	GrantPersistant(ctx context.Context, permission PermissionRequest)
	Grant(ctx context.Context, permission PermissionRequest)
	Deny(ctx context.Context, permission PermissionRequest)
	Request(ctx context.Context, opts CreatePermissionRequest) bool
	AutoApproveSession(ctx context.Context, sessionID string)
	IsAutoApproved(ctx context.Context, sessionID string) bool
	SetPermissionPromptTool(ctx context.Context, sessionID string, tool Tool)
	GetPermissionPromptTool(ctx context.Context, sessionID string) (Tool, bool)
	ParseMCPResponse(ctx context.Context, response string) bool
}

type permissionService struct {
	broker         *pubsub.Broker[PermissionRequest]
	responseBroker *pubsub.Broker[PermissionResponse]

	sessionPermissions   map[string][]PermissionRequest
	pendingRequests      sync.Map
	autoApproveSessions  map[string]bool
	permissionPromptTool map[string]Tool // sessionID -> actual tool instance
	mu                   sync.RWMutex
}

var globalPermissionService *permissionService

func InitService() error {
	if globalPermissionService != nil {
		return fmt.Errorf("permission service already initialized")
	}
	globalPermissionService = &permissionService{
		broker:               pubsub.NewBroker[PermissionRequest](),
		responseBroker:       pubsub.NewBroker[PermissionResponse](),
		sessionPermissions:   make(map[string][]PermissionRequest),
		autoApproveSessions:  make(map[string]bool),
		permissionPromptTool: make(map[string]Tool),
	}
	return nil
}

func GetService() *permissionService {
	if globalPermissionService == nil {
		panic("permission service not initialized. Call permission.InitService() first.")
	}
	return globalPermissionService
}

func (s *permissionService) GrantPersistant(ctx context.Context, permission PermissionRequest) {
	s.mu.Lock()
	s.sessionPermissions[permission.SessionID] = append(s.sessionPermissions[permission.SessionID], permission)
	s.mu.Unlock()

	respCh, ok := s.pendingRequests.Load(permission.ID)
	if ok {
		select {
		case respCh.(chan bool) <- true:
		case <-ctx.Done():
			slog.Warn("Context cancelled while sending grant persistent response", "request_id", permission.ID)
		}
	}
	s.responseBroker.Publish(EventPermissionPersisted, PermissionResponse{Request: permission, Granted: true})
}

func (s *permissionService) Grant(ctx context.Context, permission PermissionRequest) {
	respCh, ok := s.pendingRequests.Load(permission.ID)
	if ok {
		select {
		case respCh.(chan bool) <- true:
		case <-ctx.Done():
			slog.Warn("Context cancelled while sending grant response", "request_id", permission.ID)
		}
	}
	s.responseBroker.Publish(EventPermissionGranted, PermissionResponse{Request: permission, Granted: true})
}

func (s *permissionService) Deny(ctx context.Context, permission PermissionRequest) {
	respCh, ok := s.pendingRequests.Load(permission.ID)
	if ok {
		select {
		case respCh.(chan bool) <- false:
		case <-ctx.Done():
			slog.Warn("Context cancelled while sending deny response", "request_id", permission.ID)
		}
	}
	s.responseBroker.Publish(EventPermissionDenied, PermissionResponse{Request: permission, Granted: false})
}

func (s *permissionService) Request(ctx context.Context, opts CreatePermissionRequest) bool {
	s.mu.RLock()
	if s.autoApproveSessions[opts.SessionID] {
		s.mu.RUnlock()
		return true
	}

	// Check if we have a permission prompt tool configured for this session
	if tool, hasTool := s.permissionPromptTool[opts.SessionID]; hasTool {
		s.mu.RUnlock()
		return s.callPermissionPromptTool(ctx, opts, tool)
	}
	s.mu.RUnlock()

	requestPath := opts.Path
	if !filepath.IsAbs(requestPath) {
		requestPath = filepath.Join(config.WorkingDirectory(), requestPath)
	}
	requestPath = filepath.Clean(requestPath)

	if permissions, ok := s.sessionPermissions[opts.SessionID]; ok {
		for _, p := range permissions {
			storedPath := p.Path
			if !filepath.IsAbs(storedPath) {
				storedPath = filepath.Join(config.WorkingDirectory(), storedPath)
			}
			storedPath = filepath.Clean(storedPath)

			if p.ToolName == opts.ToolName && p.Action == opts.Action &&
				(requestPath == storedPath || strings.HasPrefix(requestPath, storedPath+string(filepath.Separator))) {
				s.mu.RUnlock()
				return true
			}
		}
	}
	s.mu.RUnlock()

	normalizedPath := opts.Path
	if !filepath.IsAbs(normalizedPath) {
		normalizedPath = filepath.Join(config.WorkingDirectory(), normalizedPath)
	}
	normalizedPath = filepath.Clean(normalizedPath)

	permissionReq := PermissionRequest{
		ID:          uuid.New().String(),
		Path:        normalizedPath,
		SessionID:   opts.SessionID,
		ToolName:    opts.ToolName,
		Description: opts.Description,
		Action:      opts.Action,
		Params:      opts.Params,
	}

	respCh := make(chan bool, 1)
	s.pendingRequests.Store(permissionReq.ID, respCh)
	defer s.pendingRequests.Delete(permissionReq.ID)

	s.broker.Publish(EventPermissionRequested, permissionReq)

	select {
	case resp := <-respCh:
		return resp
	case <-ctx.Done():
		slog.Warn("Permission request timed out or context cancelled", "request_id", permissionReq.ID, "tool", opts.ToolName)
		return false
	}
}

func (s *permissionService) AutoApproveSession(ctx context.Context, sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoApproveSessions[sessionID] = true
}

func (s *permissionService) IsAutoApproved(ctx context.Context, sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.autoApproveSessions[sessionID]
}

func (s *permissionService) SetPermissionPromptTool(ctx context.Context, sessionID string, tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.permissionPromptTool[sessionID] = tool
	
	slog.Info("Set permission prompt tool for session", "sessionID", sessionID, "tool", tool.Info().Name)
}

func (s *permissionService) GetPermissionPromptTool(ctx context.Context, sessionID string) (Tool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tool, exists := s.permissionPromptTool[sessionID]
	return tool, exists
}

func (s *permissionService) ParseMCPResponse(ctx context.Context, response string) bool {
	// Parse MCP tool response according to claude-code spec
	var responseObj map[string]interface{}
	if err := json.Unmarshal([]byte(response), &responseObj); err != nil {
		slog.Error("Failed to parse MCP permission response as JSON", "response", response, "error", err)
		return false
	}
	
	// Check the behavior field according to claude-code spec
	behavior, ok := responseObj["behavior"].(string)
	if !ok {
		slog.Error("Missing or invalid 'behavior' field in MCP permission response", "response", response)
		return false
	}
	
	switch behavior {
	case "allow":
		slog.Debug("Permission granted by MCP tool", "response", response)
		return true
	case "deny":
		if message, ok := responseObj["message"].(string); ok {
			slog.Info("Permission denied by MCP tool", "reason", message)
		} else {
			slog.Info("Permission denied by MCP tool")
		}
		return false
	default:
		slog.Error("Invalid behavior in MCP permission response", "behavior", behavior)
		return false
	}
}

func (s *permissionService) callPermissionPromptTool(ctx context.Context, opts CreatePermissionRequest, tool Tool) bool {
	// Create the permission request payload matching claude-code format
	requestPayload := map[string]interface{}{
		"id":          opts.SessionID + "_" + opts.ToolName + "_" + opts.Action,
		"session_id":  opts.SessionID,
		"tool_name":   opts.ToolName,
		"description": opts.Description,
		"action":      opts.Action,
		"params":      opts.Params,
		"path":        opts.Path,
	}
	
	// Convert payload to JSON string (same format as existing MCP tools expect)
	payloadJSON, err := json.Marshal(requestPayload)
	if err != nil {
		slog.Error("Failed to marshal permission request payload", "error", err)
		return false
	}
	
	// Call the tool using the existing tool interface
	toolCall := ToolCall{
		ID:    opts.SessionID + "_permission_check",
		Name:  tool.Info().Name,
		Input: string(payloadJSON),
	}
	
	result, err := tool.Run(ctx, toolCall)
	if err != nil {
		slog.Error("MCP permission tool execution failed", "error", err, "tool", tool.Info().Name)
		return false
	}
	
	if result.IsError {
		slog.Error("MCP permission tool returned error", "error", result.Content, "tool", tool.Info().Name)
		return false
	}
	
	// Parse the response using our existing parser
	return s.ParseMCPResponse(ctx, result.Content)
}

func (s *permissionService) Subscribe(ctx context.Context) <-chan pubsub.Event[PermissionRequest] {
	return s.broker.Subscribe(ctx)
}

func (s *permissionService) SubscribeToResponseEvents(ctx context.Context) <-chan pubsub.Event[PermissionResponse] {
	return s.responseBroker.Subscribe(ctx)
}

func GrantPersistant(ctx context.Context, permission PermissionRequest) {
	GetService().GrantPersistant(ctx, permission)
}

func Grant(ctx context.Context, permission PermissionRequest) {
	GetService().Grant(ctx, permission)
}

func Deny(ctx context.Context, permission PermissionRequest) {
	GetService().Deny(ctx, permission)
}

func Request(ctx context.Context, opts CreatePermissionRequest) bool {
	return GetService().Request(ctx, opts)
}

func AutoApproveSession(ctx context.Context, sessionID string) {
	GetService().AutoApproveSession(ctx, sessionID)
}

func IsAutoApproved(ctx context.Context, sessionID string) bool {
	return GetService().IsAutoApproved(ctx, sessionID)
}

func SubscribeToRequests(ctx context.Context) <-chan pubsub.Event[PermissionRequest] {
	return GetService().Subscribe(ctx)
}

func SubscribeToResponses(ctx context.Context) <-chan pubsub.Event[PermissionResponse] {
	return GetService().SubscribeToResponseEvents(ctx)
}

func SetPermissionPromptTool(ctx context.Context, sessionID string, tool Tool) {
	GetService().SetPermissionPromptTool(ctx, sessionID, tool)
}

func GetPermissionPromptTool(ctx context.Context, sessionID string) (Tool, bool) {
	return GetService().GetPermissionPromptTool(ctx, sessionID)
}

func ParseMCPResponse(ctx context.Context, response string) bool {
	return GetService().ParseMCPResponse(ctx, response)
}
