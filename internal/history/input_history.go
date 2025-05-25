// Package history provides persistent user input history management.
//
// The service tracks user input messages per session and provides navigation
// through history. Use CreateInputHistoryNavigator(sessionID) to get a
// navigator for UI components that need history functionality.
//
// Example usage:
//
//	nav := history.CreateInputHistoryNavigator("session-123")
//	nav.AddMessage(ctx, "hello world")
//	if content, ok := nav.NavigateUp(); ok {
//	    // use content
//	}
package history

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sst/opencode/internal/db"
	"github.com/sst/opencode/internal/pubsub"
)

type InputHistoryEntry struct {
	ID        string
	SessionID string
	InputText string
	CreatedAt time.Time
}

const (
	EventInputHistoryCreated        pubsub.EventType = "input_history_created"
	EventInputHistorySessionDeleted pubsub.EventType = "input_history_session_deleted"
)

type InputHistoryConfig struct {
	EnablePersistence bool
}

func DefaultInputHistoryConfig() InputHistoryConfig {
	return InputHistoryConfig{
		EnablePersistence: true,
	}
}

type InputHistoryNavigator interface {
	GetCurrentMessage() string
	SetCurrentMessage(message string)
	NavigateUp() (string, bool)
	NavigateDown() (string, bool)
	AddMessage(ctx context.Context, message string) error
	HasHistory() bool
	Reset()
}

type InputHistoryService interface {
	pubsub.Subscriber[InputHistoryEntry]
	AddInput(ctx context.Context, sessionID, inputText string) error
	GetHistory(ctx context.Context, sessionID string, limit int) ([]string, error)
	GetHistoryEntries(ctx context.Context, sessionID string, limit int) ([]InputHistoryEntry, error)
	DeleteSessionHistory(ctx context.Context, sessionID string) error
	CountHistory(ctx context.Context, sessionID string) (int64, error)
	CreateNavigator(sessionID string) InputHistoryNavigator
}

type inputHistoryService struct {
	db     *db.Queries
	sqlDB  *sql.DB
	broker *pubsub.Broker[InputHistoryEntry]
	config InputHistoryConfig
	mu     sync.RWMutex
}

var globalInputHistoryService *inputHistoryService

func InitInputHistoryService(sqlDatabase *sql.DB, config InputHistoryConfig) error {
	if globalInputHistoryService != nil {
		return fmt.Errorf("input history service already initialized")
	}

	queries := db.New(sqlDatabase)
	broker := pubsub.NewBroker[InputHistoryEntry]()

	globalInputHistoryService = &inputHistoryService{
		db:     queries,
		sqlDB:  sqlDatabase,
		broker: broker,
		config: config,
	}

	return nil
}

func GetInputHistoryService() InputHistoryService {
	if globalInputHistoryService == nil {
		panic("input history service not initialized. Call history.InitInputHistoryService() first.")
	}
	return globalInputHistoryService
}

func (s *inputHistoryService) AddInput(ctx context.Context, sessionID, inputText string) error {
	if !s.config.EnablePersistence {
		return nil 
	}

	
	inputText = strings.TrimSpace(inputText)
	if inputText == "" {
		return nil 
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	
	recent, err := s.db.GetUserInputHistoryBySession(ctx, db.GetUserInputHistoryBySessionParams{
		SessionID: sessionID,
		Limit:     1,
	})
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check for duplicate input: %w", err)
	}

	
	if len(recent) > 0 && recent[0].InputText == inputText {
		return nil
	}

	
	entry, err := s.db.CreateUserInputHistory(ctx, db.CreateUserInputHistoryParams{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		InputText: inputText,
	})
	if err != nil {
		return fmt.Errorf("failed to create input history entry: %w", err)
	}

	// Convert to domain object and publish event
	historyEntry := s.fromDBItem(entry)
	s.broker.Publish(EventInputHistoryCreated, historyEntry)

	return nil
}

func (s *inputHistoryService) GetHistory(ctx context.Context, sessionID string, limit int) ([]string, error) {
	entries, err := s.GetHistoryEntries(ctx, sessionID, limit)
	if err != nil {
		return nil, err
	}

	history := make([]string, len(entries))
	for i, entry := range entries {
		history[i] = entry.InputText
	}

	return history, nil
}

func (s *inputHistoryService) GetHistoryEntries(ctx context.Context, sessionID string, limit int) ([]InputHistoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dbEntries, err := s.db.GetUserInputHistoryBySession(ctx, db.GetUserInputHistoryBySessionParams{
		SessionID: sessionID,
		Limit:     int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get input history: %w", err)
	}

	entries := make([]InputHistoryEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		entries[i] = s.fromDBItem(dbEntry)
	}

	return entries, nil
}


func (s *inputHistoryService) DeleteSessionHistory(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.db.DeleteSessionUserInputHistory(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session input history: %w", err)
	}

	s.broker.Publish(EventInputHistorySessionDeleted, InputHistoryEntry{SessionID: sessionID})

	return nil
}

func (s *inputHistoryService) CountHistory(ctx context.Context, sessionID string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count, err := s.db.CountUserInputHistoryBySession(ctx, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to count input history entries: %w", err)
	}

	return count, nil
}

func (s *inputHistoryService) Subscribe(ctx context.Context) <-chan pubsub.Event[InputHistoryEntry] {
	return s.broker.Subscribe(ctx)
}

func (s *inputHistoryService) CreateNavigator(sessionID string) InputHistoryNavigator {
	return &inputHistoryNavigator{
		service:   s,
		sessionID: sessionID,
	}
}

type inputHistoryNavigator struct {
	service        *inputHistoryService
	sessionID      string
	historyCache   []string
	cacheLoaded    bool
	historyIndex   int
	currentMessage string
}

func (n *inputHistoryNavigator) loadCache() {
	if n.cacheLoaded {
		return
	}
	
	ctx := context.Background()
	historyEntries, err := n.service.db.GetUserInputHistoryBySession(ctx, db.GetUserInputHistoryBySessionParams{
		SessionID: n.sessionID,
		Limit:     100,
	})
	if err != nil {
		slog.Error("Failed to load input history", "error", err)
		n.historyCache = []string{}
	} else {
		n.historyCache = make([]string, len(historyEntries))
		for i, entry := range historyEntries {
			n.historyCache[len(historyEntries)-1-i] = entry.InputText
		}
	}
	
	n.cacheLoaded = true
	n.historyIndex = len(n.historyCache)
}

func (n *inputHistoryNavigator) GetCurrentMessage() string {
	return n.currentMessage
}

func (n *inputHistoryNavigator) SetCurrentMessage(message string) {
	n.currentMessage = message
}

func (n *inputHistoryNavigator) NavigateUp() (string, bool) {
	n.loadCache()
	
	if len(n.historyCache) == 0 {
		return "", false
	}
	
	if n.historyIndex > 0 {
		n.historyIndex--
		return n.historyCache[n.historyIndex], true
	}
	
	return "", false
}

func (n *inputHistoryNavigator) NavigateDown() (string, bool) {
	n.loadCache()
	
	if n.historyIndex < len(n.historyCache)-1 {
		n.historyIndex++
		return n.historyCache[n.historyIndex], true
	} else if n.historyIndex == len(n.historyCache)-1 {
		n.historyIndex = len(n.historyCache)
		return n.currentMessage, true
	}
	
	return "", false
}

func (n *inputHistoryNavigator) AddMessage(ctx context.Context, message string) error {
	err := n.service.AddInput(ctx, n.sessionID, message)
	if err != nil {
		return err
	}
	
	if n.cacheLoaded {
		if len(n.historyCache) == 0 || n.historyCache[len(n.historyCache)-1] != message {
			n.historyCache = append(n.historyCache, message)
		}
	}
	
	n.historyIndex = len(n.historyCache)
	n.currentMessage = ""
	
	return nil
}

func (n *inputHistoryNavigator) HasHistory() bool {
	n.loadCache()
	return len(n.historyCache) > 0
}

func (n *inputHistoryNavigator) Reset() {
	n.loadCache()
	n.historyIndex = len(n.historyCache)
	n.currentMessage = ""
}

func (s *inputHistoryService) fromDBItem(item db.UserInputHistory) InputHistoryEntry {
	createdAt, err := time.Parse(time.RFC3339Nano, item.CreatedAt)
	if err != nil {
		slog.Error("Failed to parse created_at for input history", "value", item.CreatedAt, "error", err)
		createdAt = time.Now()
	}

	return InputHistoryEntry{
		ID:        item.ID,
		SessionID: item.SessionID,
		InputText: item.InputText,
		CreatedAt: createdAt,
	}
}

func AddInput(ctx context.Context, sessionID, inputText string) error {
	return GetInputHistoryService().AddInput(ctx, sessionID, inputText)
}

func GetInputHistory(ctx context.Context, sessionID string, limit int) ([]string, error) {
	return GetInputHistoryService().GetHistory(ctx, sessionID, limit)
}

func GetInputHistoryEntries(ctx context.Context, sessionID string, limit int) ([]InputHistoryEntry, error) {
	return GetInputHistoryService().GetHistoryEntries(ctx, sessionID, limit)
}


func DeleteSessionInputHistory(ctx context.Context, sessionID string) error {
	return GetInputHistoryService().DeleteSessionHistory(ctx, sessionID)
}

func CountInputHistory(ctx context.Context, sessionID string) (int64, error) {
	return GetInputHistoryService().CountHistory(ctx, sessionID)
}

func SubscribeToInputHistory(ctx context.Context) <-chan pubsub.Event[InputHistoryEntry] {
	return GetInputHistoryService().Subscribe(ctx)
}

func CreateInputHistoryNavigator(sessionID string) InputHistoryNavigator {
	return GetInputHistoryService().CreateNavigator(sessionID)
}