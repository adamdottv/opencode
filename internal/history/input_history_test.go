package history

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/ncruces/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create a temporary in-memory database
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Apply migrations - create the tables
	_, err = sqlDB.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			parent_session_id TEXT,
			title TEXT NOT NULL,
			message_count INTEGER NOT NULL DEFAULT 0,
			prompt_tokens INTEGER NOT NULL DEFAULT 0,
			completion_tokens INTEGER NOT NULL DEFAULT 0,
			cost REAL NOT NULL DEFAULT 0.0,
			summary TEXT,
			summarized_at TEXT,
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%f000Z', 'now')),
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%f000Z', 'now'))
		);

		CREATE TABLE user_input_history (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			input_text TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%f000Z', 'now'))
		);

		CREATE INDEX idx_user_input_history_session_created 
		ON user_input_history(session_id, created_at DESC);
	`)
	require.NoError(t, err)

	cleanup := func() {
		sqlDB.Close()
	}

	return sqlDB, cleanup
}

func setupTestService(t *testing.T) (InputHistoryService, func()) {
	t.Helper()

	sqlDB, dbCleanup := setupTestDB(t)

	config := InputHistoryConfig{
		EnablePersistence: true,
	}

	// Reset global service for testing
	globalInputHistoryService = nil
	err := InitInputHistoryService(sqlDB, config)
	require.NoError(t, err)

	service := GetInputHistoryService()

	cleanup := func() {
		globalInputHistoryService = nil
		dbCleanup()
	}

	return service, cleanup
}

func TestInitInputHistoryService(t *testing.T) {
	t.Parallel()

	t.Run("successful initialization", func(t *testing.T) {
		t.Parallel()
		sqlDB, cleanup := setupTestDB(t)
		defer cleanup()

		// Reset global service
		globalInputHistoryService = nil

		config := DefaultInputHistoryConfig()
		err := InitInputHistoryService(sqlDB, config)
		assert.NoError(t, err)

		service := GetInputHistoryService()
		assert.NotNil(t, service)

		// Reset for other tests
		globalInputHistoryService = nil
	})

	t.Run("double initialization fails", func(t *testing.T) {
		t.Parallel()
		sqlDB, cleanup := setupTestDB(t)
		defer cleanup()

		// Reset global service
		globalInputHistoryService = nil

		config := DefaultInputHistoryConfig()
		err := InitInputHistoryService(sqlDB, config)
		require.NoError(t, err)

		// Second initialization should fail
		err = InitInputHistoryService(sqlDB, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already initialized")

		// Reset for other tests
		globalInputHistoryService = nil
	})

	t.Run("panic when service not initialized", func(t *testing.T) {
		t.Parallel()
		
		// Reset global service
		globalInputHistoryService = nil

		assert.Panics(t, func() {
			GetInputHistoryService()
		})
	})
}

func TestAddInput(t *testing.T) {
	t.Parallel()

	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	sessionID := "test-session-1"

	t.Run("add valid input", func(t *testing.T) {
		err := service.AddInput(ctx, sessionID, "test input")
		assert.NoError(t, err)

		history, err := service.GetHistory(ctx, sessionID, 10)
		require.NoError(t, err)
		assert.Len(t, history, 1)
		assert.Equal(t, "test input", history[0])
	})

	t.Run("skip empty input", func(t *testing.T) {
		initialCount, err := service.CountHistory(ctx, sessionID)
		require.NoError(t, err)

		err = service.AddInput(ctx, sessionID, "")
		assert.NoError(t, err)

		err = service.AddInput(ctx, sessionID, "   ")
		assert.NoError(t, err)

		finalCount, err := service.CountHistory(ctx, sessionID)
		require.NoError(t, err)
		assert.Equal(t, initialCount, finalCount)
	})

	t.Run("skip duplicate consecutive inputs", func(t *testing.T) {
		sessionID := "test-session-2"

		err := service.AddInput(ctx, sessionID, "duplicate input")
		require.NoError(t, err)

		err = service.AddInput(ctx, sessionID, "duplicate input")
		require.NoError(t, err)

		history, err := service.GetHistory(ctx, sessionID, 10)
		require.NoError(t, err)
		assert.Len(t, history, 1)
		assert.Equal(t, "duplicate input", history[0])
	})

	t.Run("allow non-consecutive duplicates", func(t *testing.T) {
		sessionID := "test-session-3"

		err := service.AddInput(ctx, sessionID, "input 1")
		require.NoError(t, err)

		err = service.AddInput(ctx, sessionID, "input 2")
		require.NoError(t, err)

		err = service.AddInput(ctx, sessionID, "input 1")
		require.NoError(t, err)

		history, err := service.GetHistory(ctx, sessionID, 10)
		require.NoError(t, err)
		assert.Len(t, history, 3)
		assert.Equal(t, []string{"input 1", "input 2", "input 1"}, history)
	})
}

func TestGetHistory(t *testing.T) {
	t.Parallel()

	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	sessionID := "test-session-get"

	// Add test inputs
	inputs := []string{"first", "second", "third", "fourth", "fifth"}
	for _, input := range inputs {
		err := service.AddInput(ctx, sessionID, input)
		require.NoError(t, err)
	}

	t.Run("get all history", func(t *testing.T) {
		history, err := service.GetHistory(ctx, sessionID, 10)
		require.NoError(t, err)
		assert.Len(t, history, 5)
		// Should be in reverse chronological order (most recent first)
		assert.Equal(t, []string{"fifth", "fourth", "third", "second", "first"}, history)
	})

	t.Run("get limited history", func(t *testing.T) {
		history, err := service.GetHistory(ctx, sessionID, 3)
		require.NoError(t, err)
		assert.Len(t, history, 3)
		assert.Equal(t, []string{"fifth", "fourth", "third"}, history)
	})

	t.Run("get empty history for non-existent session", func(t *testing.T) {
		history, err := service.GetHistory(ctx, "non-existent", 10)
		require.NoError(t, err)
		assert.Len(t, history, 0)
	})
}

func TestGetHistoryEntries(t *testing.T) {
	t.Parallel()

	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	sessionID := "test-session-entries"

	// Add test input
	err := service.AddInput(ctx, sessionID, "test input")
	require.NoError(t, err)

	entries, err := service.GetHistoryEntries(ctx, sessionID, 10)
	require.NoError(t, err)
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.NotEmpty(t, entry.ID)
	assert.Equal(t, sessionID, entry.SessionID)
	assert.Equal(t, "test input", entry.InputText)
	assert.False(t, entry.CreatedAt.IsZero())
}


func TestDeleteSessionHistory(t *testing.T) {
	t.Parallel()

	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	sessionID := "test-session-delete"

	// Add test inputs
	for i := 1; i <= 3; i++ {
		err := service.AddInput(ctx, sessionID, fmt.Sprintf("input %d", i))
		require.NoError(t, err)
	}

	// Verify inputs exist
	count, err := service.CountHistory(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Delete session history
	err = service.DeleteSessionHistory(ctx, sessionID)
	require.NoError(t, err)

	// Verify history is deleted
	count, err = service.CountHistory(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	history, err := service.GetHistory(ctx, sessionID, 10)
	require.NoError(t, err)
	assert.Len(t, history, 0)
}

func TestCountHistory(t *testing.T) {
	t.Parallel()

	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	sessionID := "test-session-count"

	// Initially should be 0
	count, err := service.CountHistory(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Add inputs and verify count increases
	for i := 1; i <= 3; i++ {
		err := service.AddInput(ctx, sessionID, fmt.Sprintf("input %d", i))
		require.NoError(t, err)

		count, err := service.CountHistory(ctx, sessionID)
		require.NoError(t, err)
		assert.Equal(t, int64(i), count)
	}
}

func TestConfigurationDisablePersistence(t *testing.T) {
	t.Parallel()

	sqlDB, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	config := InputHistoryConfig{
		EnablePersistence: false, // Disabled
	}

	// Reset global service
	globalInputHistoryService = nil
	err := InitInputHistoryService(sqlDB, config)
	require.NoError(t, err)

	service := GetInputHistoryService()
	ctx := context.Background()
	sessionID := "test-session-disabled"

	// Adding input should succeed but not persist
	err = service.AddInput(ctx, sessionID, "test input")
	assert.NoError(t, err)

	// Should not find any history
	count, err := service.CountHistory(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Reset for other tests
	globalInputHistoryService = nil
}

func TestPackageLevelFunctions(t *testing.T) {
	// Don't run in parallel to avoid conflicts with global service
	sqlDB, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	config := InputHistoryConfig{
		EnablePersistence: true,
	}

	// Reset and initialize global service for this test
	globalInputHistoryService = nil
	err := InitInputHistoryService(sqlDB, config)
	require.NoError(t, err)

	defer func() {
		globalInputHistoryService = nil
	}()

	ctx := context.Background()
	sessionID := "test-session-package"

	// Test package-level convenience functions
	err = AddInput(ctx, sessionID, "package test")
	assert.NoError(t, err)

	history, err := GetInputHistory(ctx, sessionID, 10)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "package test", history[0])

	entries, err := GetInputHistoryEntries(ctx, sessionID, 10)
	require.NoError(t, err)
	assert.Len(t, entries, 1)

	count, err := CountInputHistory(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)


	err = DeleteSessionInputHistory(ctx, sessionID)
	assert.NoError(t, err)

	count, err = CountInputHistory(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

