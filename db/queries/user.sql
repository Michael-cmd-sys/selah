-- name: GetSpiritualGoals :many
SELECT * FROM spiritual_goals WHERE user_id = $1;

-- name: UpsertSpiritualGoal :one
INSERT INTO spiritual_goals (user_id, goal_type, target, unit, frequency)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, goal_type) DO UPDATE SET
    target     = EXCLUDED.target,
    unit       = EXCLUDED.unit,
    frequency  = EXCLUDED.frequency,
    updated_at = NOW()
RETURNING *;

-- name: DeleteSpiritualGoal :exec
DELETE FROM spiritual_goals WHERE user_id = $1 AND goal_type = $2;

-- name: GetReminders :many
SELECT * FROM user_reminders WHERE user_id = $1;

-- name: UpsertReminder :one
INSERT INTO user_reminders (user_id, type, time_hhmm, enabled, label)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, type) DO UPDATE SET
    time_hhmm  = EXCLUDED.time_hhmm,
    enabled    = EXCLUDED.enabled,
    label      = EXCLUDED.label,
    updated_at = NOW()
RETURNING *;

-- name: SearchUsers :many
SELECT id, name, avatar_url, level
FROM users
WHERE name ILIKE '%' || $1 || '%'
  AND visibility != 'private'
LIMIT 20;

-- name: GetFriendships :many
SELECT f.friend_id, u.name, u.avatar_url, u.level, f.status, f.created_at
FROM friendships f
JOIN users u ON u.id = f.friend_id
WHERE f.user_id = $1 AND f.status = 'accepted';

-- name: SendFriendRequest :exec
INSERT INTO friendships (user_id, friend_id, status)
VALUES ($1, $2, 'pending')
ON CONFLICT (user_id, friend_id) DO NOTHING;

-- name: RespondFriendRequest :exec
UPDATE friendships SET status = $3
WHERE user_id = $1 AND friend_id = $2;

-- name: AddXP :one
UPDATE users SET
    xp    = xp + $2,
    level = GREATEST(1, FLOOR(SQRT((xp + $2)::float / 100))::int + 1),
    updated_at = NOW()
WHERE id = $1
RETURNING xp, level;
