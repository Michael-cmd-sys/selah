-- +goose Up
CREATE TABLE streaks (
    user_id            UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    activity_type      TEXT NOT NULL CHECK (activity_type IN ('meditation','fasting','reading','prayer')),
    current_streak     INT NOT NULL DEFAULT 0,
    longest_streak     INT NOT NULL DEFAULT 0,
    last_activity_date DATE,
    PRIMARY KEY (user_id, activity_type)
);

CREATE TABLE milestones (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    label       TEXT NOT NULL,
    earned_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, type)
);

CREATE INDEX idx_milestones_user ON milestones (user_id);

-- +goose Down
DROP TABLE milestones;
DROP TABLE streaks;
