-- name: StartFast :one
INSERT INTO fasts (user_id, target_hours, started_at)
VALUES ($1, $2, NOW())
RETURNING *;

-- name: EndFast :one
UPDATE fasts SET ended_at = NOW()
WHERE id = $1 AND user_id = $2 AND ended_at IS NULL
RETURNING *;

-- name: GetActiveFast :one
SELECT * FROM fasts
WHERE user_id = $1 AND ended_at IS NULL
ORDER BY started_at DESC
LIMIT 1;

-- name: GetFastByID :one
SELECT * FROM fasts WHERE id = $1 AND user_id = $2;

-- name: ListFasts :many
SELECT * FROM fasts
WHERE user_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: AddCheckin :one
INSERT INTO fast_checkins (fast_id, user_id, mood, reflection)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLatestCheckin :one
SELECT * FROM fast_checkins
WHERE fast_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListCheckins :many
SELECT * FROM fast_checkins
WHERE fast_id = $1
ORDER BY created_at DESC;
