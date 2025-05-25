-- name: CreateUserInputHistory :one
INSERT INTO user_input_history (
    id,
    session_id,
    input_text
) VALUES (
    ?, ?, ?
)
RETURNING *;

-- name: GetUserInputHistoryBySession :many
SELECT *
FROM user_input_history
WHERE session_id = ?
ORDER BY created_at DESC, ROWID DESC
LIMIT ?;


-- name: DeleteSessionUserInputHistory :exec
DELETE FROM user_input_history
WHERE session_id = ?;

-- name: CountUserInputHistoryBySession :one
SELECT COUNT(*)
FROM user_input_history
WHERE session_id = ?;