-- +goose Up
CREATE TABLE fasts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    target_hours NUMERIC(5,2) NOT NULL DEFAULT 16,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at     TIMESTAMPTZ,
    shared       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fasts_user ON fasts (user_id, started_at DESC);

-- Mood + reflection snapshots taken during a fast (latest = current state shown in UI)
CREATE TABLE fast_checkins (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fast_id    UUID NOT NULL REFERENCES fasts (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    mood       TEXT NOT NULL CHECK (mood IN ('peaceful','focused','zealous','heavy','weary','joyful')),
    reflection TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fast_checkins_fast ON fast_checkins (fast_id, created_at DESC);

-- +goose Down
DROP TABLE fast_checkins;
DROP TABLE fasts;
