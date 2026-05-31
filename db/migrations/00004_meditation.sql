-- +goose Up
CREATE TABLE meditation_sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID REFERENCES users (id) ON DELETE SET NULL,
    title            TEXT NOT NULL DEFAULT 'Session',
    session_type     TEXT NOT NULL DEFAULT 'meditation' CHECK (session_type IN ('meditation','prayer','reading','breathwork')),
    duration_seconds INT NOT NULL CHECK (duration_seconds > 0),
    completed        BOOLEAN NOT NULL DEFAULT FALSE,
    verse_book       TEXT,
    verse_chapter    INT,
    verse_number     INT,
    verse_text       TEXT,
    verse_ref        TEXT,
    note             TEXT,
    started_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at     TIMESTAMPTZ
);

CREATE INDEX idx_meditation_sessions_user     ON meditation_sessions (user_id, started_at DESC);
CREATE INDEX idx_meditation_sessions_completed ON meditation_sessions (user_id, completed) WHERE completed = TRUE;

-- +goose Down
DROP TABLE meditation_sessions;
