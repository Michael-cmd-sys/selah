-- +goose Up
CREATE TABLE groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    creator_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    is_private  BOOLEAN NOT NULL DEFAULT FALSE,
    avatar_url  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE group_members (
    group_id   UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role       TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('admin','member')),
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_members_user ON group_members (user_id);

CREATE TABLE challenges (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID REFERENCES groups (id) ON DELETE CASCADE,
    creator_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT,
    type        TEXT NOT NULL CHECK (type IN ('fasting','meditation','reading','prayer','custom')),
    target      INT NOT NULL DEFAULT 1,
    unit        TEXT NOT NULL DEFAULT 'days',
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_challenges_group ON challenges (group_id, start_date);

CREATE TABLE challenge_participants (
    challenge_id UUID NOT NULL REFERENCES challenges (id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    progress     INT NOT NULL DEFAULT 0,
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (challenge_id, user_id)
);

CREATE TABLE posts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    group_id   UUID REFERENCES groups (id) ON DELETE CASCADE,
    post_type  TEXT NOT NULL CHECK (post_type IN ('praise','prayer_request','scripture','general')),
    content    TEXT NOT NULL,
    verse_ref  TEXT,
    verse_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_group     ON posts (group_id, created_at DESC);
CREATE INDEX idx_posts_user      ON posts (user_id, created_at DESC);
CREATE INDEX idx_posts_feed      ON posts (created_at DESC);

CREATE TABLE post_reactions (
    post_id       UUID NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reaction_type TEXT NOT NULL CHECK (reaction_type IN ('amen','praying')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id, reaction_type)
);

CREATE TABLE post_comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id    UUID NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_post_comments_post ON post_comments (post_id, created_at ASC);

-- +goose Down
DROP TABLE post_comments;
DROP TABLE post_reactions;
DROP TABLE posts;
DROP TABLE challenge_participants;
DROP TABLE challenges;
DROP TABLE group_members;
DROP TABLE groups;
