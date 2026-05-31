-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          TEXT UNIQUE,
    name           TEXT NOT NULL DEFAULT '',
    avatar_url     TEXT,
    bio            TEXT,
    level          INT NOT NULL DEFAULT 1,
    xp             INT NOT NULL DEFAULT 0,
    visibility     TEXT NOT NULL DEFAULT 'friends' CHECK (visibility IN ('public','friends','private')),
    google_sub     TEXT UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email     ON users (email)      WHERE email IS NOT NULL;
CREATE INDEX idx_users_google_sub ON users (google_sub) WHERE google_sub IS NOT NULL;

CREATE TABLE sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    ip_address TEXT,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_token   ON sessions (token);
CREATE INDEX idx_sessions_user_id ON sessions (user_id);

CREATE TABLE user_credentials (
    user_id       UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE user_credentials;
DROP TABLE sessions;
DROP TABLE users;
