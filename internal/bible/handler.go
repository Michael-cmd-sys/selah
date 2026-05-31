package bible

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/selah/internal/auth"
	"github.com/selah/internal/dbsqlc"
	"github.com/selah/internal/response"
)

type Handler struct {
	q      *dbsqlc.Queries
	client *helloaoClient
}

func NewHandler(q *dbsqlc.Queries) *Handler {
	return &Handler{
		q:      q,
		client: &helloaoClient{http: &http.Client{Timeout: 10 * time.Second}},
	}
}

func (h *Handler) ListTranslations(w http.ResponseWriter, r *http.Request) {
	translations, err := h.q.GetTranslations(r.Context())
	if err != nil {
		response.Internal(w, "could not fetch translations")
		return
	}
	response.OK(w, translations)
}

func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	translation := chi.URLParam(r, "translation")
	books, err := h.client.books(translation)
	if err != nil {
		response.Internal(w, "could not fetch books")
		return
	}
	response.OK(w, books)
}

func (h *Handler) GetChapter(w http.ResponseWriter, r *http.Request) {
	translation := chi.URLParam(r, "translation")
	book := chi.URLParam(r, "book")
	chapterStr := chi.URLParam(r, "chapter")

	chapter, err := strconv.Atoi(chapterStr)
	if err != nil || chapter < 1 {
		response.BadRequest(w, "invalid chapter number")
		return
	}

	data, err := h.client.chapter(translation, book, chapter)
	if err != nil {
		response.Internal(w, "could not fetch chapter")
		return
	}

	// Track reading progress for authenticated users
	if userID := auth.UserIDFromCtx(r.Context()); userID != uuid.Nil {
		if translationRow, dbErr := h.q.GetTranslationByCode(r.Context(), translation); dbErr == nil {
			_ = h.q.UpsertReadingProgress(r.Context(), dbsqlc.UpsertReadingProgressParams{
				UserID:        userID,
				TranslationID: translationRow.ID,
				Book:          book,
				Chapter:       int32(chapter),
			})
		}
	}

	response.OK(w, data)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	translation := chi.URLParam(r, "translation")
	q := r.URL.Query().Get("q")
	if q == "" {
		response.BadRequest(w, "q is required")
		return
	}

	// Prefer DB search for cached translations, fallback to API
	if translationRow, err := h.q.GetTranslationByCode(r.Context(), translation); err == nil {
		results, err := h.q.SearchVerses(r.Context(), dbsqlc.SearchVersesParams{
			TranslationID: translationRow.ID,
			Column2:       &q,
		})
		if err == nil {
			response.OK(w, results)
			return
		}
	}

	results, err := h.client.search(translation, q)
	if err != nil {
		response.Internal(w, "search failed")
		return
	}
	response.OK(w, results)
}

func (h *Handler) GetDailyVerse(w http.ResponseWriter, r *http.Request) {
	translation := r.URL.Query().Get("translation")
	if translation == "" {
		translation = "KJV"
	}

	translationRow, err := h.q.GetTranslationByCode(r.Context(), translation)
	if err != nil {
		response.NotFound(w, "translation not found")
		return
	}

	today := pgtype.Date{Time: time.Now().UTC(), Valid: true}
	verse, err := h.q.GetDailyVerse(r.Context(), dbsqlc.GetDailyVerseParams{
		CacheDate:     today,
		TranslationID: translationRow.ID,
	})
	if err != nil {
		response.NotFound(w, "daily verse not available")
		return
	}
	response.OK(w, verse)
}

type highlightRequest struct {
	TranslationID int64   `json:"translationId"`
	Book          string  `json:"book"`
	Chapter       int32   `json:"chapter"`
	Verse         int32   `json:"verse"`
	Color         string  `json:"color"`
	Note          *string `json:"note"`
}

func (h *Handler) GetHighlights(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	highlights, err := h.q.GetHighlights(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch highlights")
		return
	}
	response.OK(w, highlights)
}

func (h *Handler) UpsertHighlight(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req highlightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	highlight, err := h.q.UpsertHighlight(r.Context(), dbsqlc.UpsertHighlightParams{
		UserID:        userID,
		TranslationID: req.TranslationID,
		Book:          req.Book,
		Chapter:       req.Chapter,
		Verse:         req.Verse,
		Color:         req.Color,
		Note:          req.Note,
	})
	if err != nil {
		response.Internal(w, "could not save highlight")
		return
	}
	response.OK(w, highlight)
}

func (h *Handler) DeleteHighlight(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid highlight id")
		return
	}
	if err := h.q.DeleteHighlight(r.Context(), dbsqlc.DeleteHighlightParams{ID: id, UserID: userID}); err != nil {
		response.Internal(w, "could not delete highlight")
		return
	}
	response.NoContent(w)
}

type bookmarkRequest struct {
	TranslationID int64  `json:"translationId"`
	Book          string `json:"book"`
	Chapter       int32  `json:"chapter"`
	Verse         int32  `json:"verse"`
}

func (h *Handler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	bookmarks, err := h.q.GetBookmarks(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch bookmarks")
		return
	}
	response.OK(w, bookmarks)
}

func (h *Handler) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req bookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	bookmark, err := h.q.CreateBookmark(r.Context(), dbsqlc.CreateBookmarkParams{
		UserID:        userID,
		TranslationID: req.TranslationID,
		Book:          req.Book,
		Chapter:       req.Chapter,
		Verse:         req.Verse,
	})
	if err != nil {
		response.Internal(w, "could not create bookmark")
		return
	}
	response.Created(w, bookmark)
}

func (h *Handler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid bookmark id")
		return
	}
	if err := h.q.DeleteBookmark(r.Context(), dbsqlc.DeleteBookmarkParams{ID: id, UserID: userID}); err != nil {
		response.Internal(w, "could not delete bookmark")
		return
	}
	response.NoContent(w)
}
