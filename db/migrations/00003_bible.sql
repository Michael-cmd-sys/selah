-- +goose Up
CREATE TABLE bible_translations (
    id         BIGSERIAL PRIMARY KEY,
    code       TEXT NOT NULL UNIQUE,
    short_name TEXT,
    name       TEXT NOT NULL,
    language   TEXT NOT NULL DEFAULT 'en',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bible_verses (
    id             BIGSERIAL PRIMARY KEY,
    translation_id BIGINT NOT NULL REFERENCES bible_translations (id) ON DELETE CASCADE,
    book           TEXT NOT NULL,
    chapter        INT NOT NULL CHECK (chapter >= 1),
    verse          INT NOT NULL CHECK (verse >= 1),
    text           TEXT NOT NULL,
    UNIQUE (translation_id, book, chapter, verse)
);

CREATE INDEX idx_bible_verses_lookup ON bible_verses (translation_id, book, chapter);

CREATE TABLE bible_daily_cache (
    cache_date     DATE NOT NULL,
    translation_id BIGINT NOT NULL REFERENCES bible_translations (id) ON DELETE CASCADE,
    book           TEXT NOT NULL,
    chapter        INT NOT NULL,
    verse          INT NOT NULL,
    PRIMARY KEY (cache_date, translation_id)
);

CREATE TABLE bible_highlights (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    translation_id BIGINT NOT NULL REFERENCES bible_translations (id) ON DELETE CASCADE,
    book           TEXT NOT NULL,
    chapter        INT NOT NULL,
    verse          INT NOT NULL,
    color          TEXT NOT NULL DEFAULT 'yellow',
    note           TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, translation_id, book, chapter, verse)
);

CREATE INDEX idx_bible_highlights_user ON bible_highlights (user_id);

CREATE TABLE bible_bookmarks (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    translation_id BIGINT NOT NULL REFERENCES bible_translations (id) ON DELETE CASCADE,
    book           TEXT NOT NULL,
    chapter        INT NOT NULL,
    verse          INT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, translation_id, book, chapter, verse)
);

CREATE INDEX idx_bible_bookmarks_user ON bible_bookmarks (user_id);

CREATE TABLE reading_progress (
    user_id        UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    translation_id BIGINT NOT NULL REFERENCES bible_translations (id) ON DELETE CASCADE,
    book           TEXT NOT NULL,
    chapter        INT NOT NULL,
    last_read_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, translation_id, book)
);

-- +goose Down
DROP TABLE reading_progress;
DROP TABLE bible_bookmarks;
DROP TABLE bible_highlights;
DROP TABLE bible_daily_cache;
DROP TABLE bible_verses;
DROP TABLE bible_translations;
