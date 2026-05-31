-- name: CreateMeditationSession :one
INSERT INTO meditation_sessions (
    user_id, title, session_type, duration_seconds,
    verse_book, verse_chapter, verse_number, verse_text, verse_ref
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: CompleteMeditationSession :one
UPDATE meditation_sessions SET
    completed    = TRUE,
    note         = COALESCE($2, note),
    completed_at = NOW()
WHERE id = $1 AND user_id = $3
RETURNING *;

-- name: GetMeditationSession :one
SELECT * FROM meditation_sessions WHERE id = $1 AND user_id = $2;

-- name: ListMeditationSessions :many
SELECT * FROM meditation_sessions
WHERE user_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCompletedSessionsOnDate :one
SELECT COUNT(*) FROM meditation_sessions
WHERE user_id = $1
  AND completed = TRUE
  AND DATE(completed_at) = $2::date;
