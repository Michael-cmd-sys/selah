package fasting

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

type startRequest struct {
	TargetHours float64 `json:"targetHours"`
}

type checkinRequest struct {
	Mood       string  `json:"mood"`
	Reflection *string `json:"reflection"`
}

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	fast, err := h.q.GetActiveFast(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "no active fast")
		return
	}

	checkin, _ := h.q.GetLatestCheckin(r.Context(), fast.ID)
	response.OK(w, map[string]any{"fast": fast, "latestCheckin": checkin})
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req startRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.TargetHours <= 0 {
		req.TargetHours = 16
	}

	fast, err := h.q.StartFast(r.Context(), dbsqlc.StartFastParams{
		UserID:      userID,
		TargetHours: req.TargetHours,
	})
	if err != nil {
		response.Internal(w, "could not start fast")
		return
	}
	response.Created(w, fast)
}

func (h *Handler) End(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid fast id")
		return
	}

	fast, err := h.q.EndFast(r.Context(), dbsqlc.EndFastParams{ID: id, UserID: userID})
	if err != nil {
		response.NotFound(w, "active fast not found")
		return
	}
	response.OK(w, fast)
}

func (h *Handler) AddCheckin(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	fastID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid fast id")
		return
	}

	var req checkinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Mood == "" {
		response.BadRequest(w, "mood is required")
		return
	}

	checkin, err := h.q.AddCheckin(r.Context(), dbsqlc.AddCheckinParams{
		FastID:     fastID,
		UserID:     userID,
		Mood:       req.Mood,
		Reflection: req.Reflection,
	})
	if err != nil {
		response.Internal(w, "could not save check-in")
		return
	}
	response.Created(w, checkin)
}

func (h *Handler) GetFast(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid fast id")
		return
	}

	fast, err := h.q.GetFastByID(r.Context(), dbsqlc.GetFastByIDParams{ID: id, UserID: userID})
	if err != nil {
		response.NotFound(w, "fast not found")
		return
	}
	checkins, _ := h.q.ListCheckins(r.Context(), fast.ID)
	response.OK(w, map[string]any{"fast": fast, "checkins": checkins})
}

func (h *Handler) ListFasts(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	fasts, err := h.q.ListFasts(r.Context(), dbsqlc.ListFastsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		response.Internal(w, "could not fetch fasts")
		return
	}
	response.OK(w, fasts)
}
