-- name: CreateUser :one
INSERT INTO users (email, name, avatar_url, google_sub, email_verified)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByGoogleSub :one
SELECT * FROM users WHERE google_sub = $1;

-- name: UpsertGoogleUser :one
INSERT INTO users (google_sub, email, name, avatar_url, email_verified)
VALUES ($1, $2, $3, $4, TRUE)
ON CONFLICT (google_sub) DO UPDATE SET
    email      = EXCLUDED.email,
    name       = COALESCE(NULLIF(users.name, ''), EXCLUDED.name),
    avatar_url = COALESCE(users.avatar_url, EXCLUDED.avatar_url),
    updated_at = NOW()
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    name       = COALESCE(sqlc.narg(name), name),
    avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
    bio        = COALESCE(sqlc.narg(bio), bio),
    visibility = COALESCE(sqlc.narg(visibility), visibility),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateCredential :exec
INSERT INTO user_credentials (user_id, password_hash) VALUES ($1, $2);

-- name: GetCredentialByUserID :one
SELECT password_hash FROM user_credentials WHERE user_id = $1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions WHERE token = $1 AND expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= NOW();
