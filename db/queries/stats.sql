-- name: GetStreaks :many
SELECT * FROM streaks WHERE user_id = $1;

-- name: GetStreakByType :one
SELECT * FROM streaks WHERE user_id = $1 AND activity_type = $2;

-- name: UpsertStreak :one
INSERT INTO streaks (user_id, activity_type, current_streak, longest_streak, last_activity_date)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, activity_type) DO UPDATE SET
    current_streak     = EXCLUDED.current_streak,
    longest_streak     = GREATEST(streaks.longest_streak, EXCLUDED.longest_streak),
    last_activity_date = EXCLUDED.last_activity_date
RETURNING *;

-- name: GetMilestones :many
SELECT * FROM milestones WHERE user_id = $1 ORDER BY earned_at DESC;

-- name: AwardMilestone :one
INSERT INTO milestones (user_id, type, label)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, type) DO NOTHING
RETURNING *;

-- name: GetWeeklySessionMinutes :many
SELECT
    DATE(started_at) AS day,
    SUM(duration_seconds) / 60 AS minutes
FROM meditation_sessions
WHERE user_id = $1
  AND completed = TRUE
  AND started_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(started_at)
ORDER BY day;

-- name: GetConsistencyMap :many
SELECT DISTINCT DATE(started_at) AS active_date
FROM meditation_sessions
WHERE user_id = $1
  AND completed = TRUE
  AND started_at >= $2::timestamptz
  AND started_at < $3::timestamptz
ORDER BY active_date;
