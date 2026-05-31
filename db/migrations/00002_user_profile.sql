-- +goose Up
CREATE TABLE spiritual_goals (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    goal_type  TEXT NOT NULL CHECK (goal_type IN ('meditation','fasting','reading','prayer')),
    target     INT NOT NULL DEFAULT 0,
    unit       TEXT NOT NULL DEFAULT 'minutes' CHECK (unit IN ('minutes','hours','days','sessions')),
    frequency  TEXT NOT NULL DEFAULT 'daily' CHECK (frequency IN ('daily','weekly')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, goal_type)
);

CREATE TABLE user_reminders (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type       TEXT NOT NULL CHECK (type IN ('morning','evening','custom')),
    time_hhmm  TEXT NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    label      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, type)
);

CREATE TABLE friendships (
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    friend_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status     TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','blocked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, friend_id)
);

CREATE INDEX idx_friendships_friend_id ON friendships (friend_id);

-- +goose Down
DROP TABLE friendships;
DROP TABLE user_reminders;
DROP TABLE spiritual_goals;
