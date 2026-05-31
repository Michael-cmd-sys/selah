package stats

import (
	"net/http"
	"time"

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

func (h *Handler) GetJourney(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	streaks, err := h.q.GetStreaks(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch streaks")
		return
	}

	milestones, err := h.q.GetMilestones(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch milestones")
		return
	}

	weekly, err := h.q.GetWeeklySessionMinutes(r.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		response.Internal(w, "could not fetch weekly data")
		return
	}

	response.OK(w, map[string]any{
		"streaks":    streaks,
		"milestones": milestones,
		"weekly":     weekly,
	})
}

func (h *Handler) GetConsistency(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	// Default: current month
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	days, err := h.q.GetConsistencyMap(r.Context(), dbsqlc.GetConsistencyMapParams{
		UserID:  pgtype.UUID{Bytes: userID, Valid: true},
		Column2: pgtype.Timestamptz{Time: from, Valid: true},
		Column3: pgtype.Timestamptz{Time: to, Valid: true},
	})
	if err != nil {
		response.Internal(w, "could not fetch consistency data")
		return
	}
	response.OK(w, map[string]any{"activeDates": days})
}

func (h *Handler) GetMilestones(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	milestones, err := h.q.GetMilestones(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch milestones")
		return
	}
	response.OK(w, milestones)
}
