-- name: GetTranslations :many
SELECT * FROM bible_translations ORDER BY language, name;

-- name: GetTranslationByCode :one
SELECT * FROM bible_translations WHERE code = $1;

-- name: UpsertTranslation :one
INSERT INTO bible_translations (code, short_name, name, language)
VALUES ($1, $2, $3, $4)
ON CONFLICT (code) DO UPDATE SET
    short_name = EXCLUDED.short_name,
    name       = EXCLUDED.name,
    language   = EXCLUDED.language
RETURNING *;

-- name: GetChapter :many
SELECT verse, text
FROM bible_verses
WHERE translation_id = $1 AND book = $2 AND chapter = $3
ORDER BY verse;

-- name: SearchVerses :many
SELECT bv.book, bv.chapter, bv.verse, bv.text, bt.code AS translation_code
FROM bible_verses bv
JOIN bible_translations bt ON bt.id = bv.translation_id
WHERE bv.translation_id = $1 AND bv.text ILIKE '%' || $2 || '%'
ORDER BY bv.book, bv.chapter, bv.verse
LIMIT 50;

-- name: GetDailyVerse :one
SELECT bv.book, bv.chapter, bv.verse, bv.text, bt.code AS translation_code
FROM bible_daily_cache dc
JOIN bible_verses bv ON bv.translation_id = dc.translation_id
    AND bv.book = dc.book AND bv.chapter = dc.chapter AND bv.verse = dc.verse
JOIN bible_translations bt ON bt.id = dc.translation_id
WHERE dc.cache_date = $1 AND dc.translation_id = $2;

-- name: UpsertDailyVerse :exec
INSERT INTO bible_daily_cache (cache_date, translation_id, book, chapter, verse)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (cache_date, translation_id) DO UPDATE SET
    book    = EXCLUDED.book,
    chapter = EXCLUDED.chapter,
    verse   = EXCLUDED.verse;

-- name: GetHighlights :many
SELECT * FROM bible_highlights WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpsertHighlight :one
INSERT INTO bible_highlights (user_id, translation_id, book, chapter, verse, color, note)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, translation_id, book, chapter, verse) DO UPDATE SET
    color = EXCLUDED.color,
    note  = EXCLUDED.note
RETURNING *;

-- name: DeleteHighlight :exec
DELETE FROM bible_highlights WHERE id = $1 AND user_id = $2;

-- name: GetBookmarks :many
SELECT * FROM bible_bookmarks WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateBookmark :one
INSERT INTO bible_bookmarks (user_id, translation_id, book, chapter, verse)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, translation_id, book, chapter, verse) DO NOTHING
RETURNING *;

-- name: DeleteBookmark :exec
DELETE FROM bible_bookmarks WHERE id = $1 AND user_id = $2;

-- name: UpsertReadingProgress :exec
INSERT INTO reading_progress (user_id, translation_id, book, chapter, last_read_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (user_id, translation_id, book) DO UPDATE SET
    chapter      = EXCLUDED.chapter,
    last_read_at = NOW();

-- name: GetReadingProgress :many
SELECT * FROM reading_progress WHERE user_id = $1 ORDER BY last_read_at DESC;
