-- +goose Up
-- +goose StatementBegin
-- User Input History
CREATE TABLE IF NOT EXISTS user_input_history (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    input_text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%f000Z', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_user_input_history_session_created 
ON user_input_history(session_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_input_history_session_created;
DROP TABLE IF EXISTS user_input_history;
-- +goose StatementEnd