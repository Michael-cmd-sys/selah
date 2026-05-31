package session

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/selah/internal/auth"
	"github.com/selah/internal/dbsqlc"
	"github.com/selah/internal/response"
)

type Handler struct {
	q *dbsqlc.Queries
}

func NewHandler(q *dbsqlc.Queries) *Handler {
	return &Handler{q: q}
}

type createRequest struct {
	Title           string  `json:"title"`
	SessionType     string  `json:"sessionType"`
	DurationSeconds int32   `json:"durationSeconds"`
	VerseBook       *string `json:"verseBook"`
	VerseChapter    *int32  `json:"verseChapter"`
	VerseNumber     *int32  `json:"verseNumber"`
	VerseText       *string `json:"verseText"`
	VerseRef        *string `json:"verseRef"`
}

type completeRequest struct {
	Note *string `json:"note"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.DurationSeconds <= 0 {
		response.BadRequest(w, "durationSeconds must be positive")
		return
	}
	if req.Title == "" {
		req.Title = "Session"
	}
	if req.SessionType == "" {
		req.SessionType = "meditation"
	}

	sess, err := h.q.CreateMeditationSession(r.Context(), dbsqlc.CreateMeditationSessionParams{
		UserID:          pgtype.UUID{Bytes: userID, Valid: userID != uuid.Nil},
		Title:           req.Title,
		SessionType:     req.SessionType,
		DurationSeconds: req.DurationSeconds,
		VerseBook:       req.VerseBook,
		VerseChapter:    req.VerseChapter,
		VerseNumber:     req.VerseNumber,
		VerseText:       req.VerseText,
		VerseRef:        req.VerseRef,
	})
	if err != nil {
		response.Internal(w, "could not create session")
		return
	}
	response.Created(w, sess)
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid session id")
		return
	}

	var req completeRequest
	json.NewDecoder(r.Body).Decode(&req)

	sess, err := h.q.CompleteMeditationSession(r.Context(), dbsqlc.CompleteMeditationSessionParams{
		ID:     id,
		Note:   req.Note,
		UserID: pgtype.UUID{Bytes: userID, Valid: userID != uuid.Nil},
	})
	if err != nil {
		response.NotFound(w, "session not found")
		return
	}
	response.OK(w, sess)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid session id")
		return
	}

	sess, err := h.q.GetMeditationSession(r.Context(), dbsqlc.GetMeditationSessionParams{
		ID:     id,
		UserID: pgtype.UUID{Bytes: userID, Valid: userID != uuid.Nil},
	})
	if err != nil {
		response.NotFound(w, "session not found")
		return
	}
	response.OK(w, sess)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	sessions, err := h.q.ListMeditationSessions(r.Context(), dbsqlc.ListMeditationSessionsParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		response.Internal(w, "could not fetch sessions")
		return
	}
	response.OK(w, sessions)
}
