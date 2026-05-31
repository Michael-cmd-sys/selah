package user

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
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

type updateProfileRequest struct {
	Name       *string `json:"name"`
	AvatarURL  *string `json:"avatarUrl"`
	Bio        *string `json:"bio"`
	Visibility *string `json:"visibility"`
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	user, err := h.q.GetUserByID(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "user not found")
		return
	}
	response.OK(w, user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		req.Name = &trimmed
	}

	user, err := h.q.UpdateUser(r.Context(), dbsqlc.UpdateUserParams{
		ID:         userID,
		Name:       req.Name,
		AvatarUrl:  req.AvatarURL,
		Bio:        req.Bio,
		Visibility: req.Visibility,
	})
	if err != nil {
		response.Internal(w, "could not update profile")
		return
	}
	response.OK(w, user)
}

type goalRequest struct {
	GoalType  string `json:"goalType"`
	Target    int32  `json:"target"`
	Unit      string `json:"unit"`
	Frequency string `json:"frequency"`
}

func (h *Handler) GetGoals(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	goals, err := h.q.GetSpiritualGoals(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch goals")
		return
	}
	response.OK(w, goals)
}

func (h *Handler) UpsertGoal(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req goalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	goal, err := h.q.UpsertSpiritualGoal(r.Context(), dbsqlc.UpsertSpiritualGoalParams{
		UserID:    userID,
		GoalType:  req.GoalType,
		Target:    req.Target,
		Unit:      req.Unit,
		Frequency: req.Frequency,
	})
	if err != nil {
		response.Internal(w, "could not save goal")
		return
	}
	response.OK(w, goal)
}

func (h *Handler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	goalType := chi.URLParam(r, "type")

	if err := h.q.DeleteSpiritualGoal(r.Context(), dbsqlc.DeleteSpiritualGoalParams{
		UserID:   userID,
		GoalType: goalType,
	}); err != nil {
		response.Internal(w, "could not delete goal")
		return
	}
	response.NoContent(w)
}

type reminderRequest struct {
	Type     string  `json:"type"`
	TimeHHMM string  `json:"time"`
	Enabled  bool    `json:"enabled"`
	Label    *string `json:"label"`
}

func (h *Handler) GetReminders(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	reminders, err := h.q.GetReminders(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch reminders")
		return
	}
	response.OK(w, reminders)
}

func (h *Handler) UpsertReminder(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req reminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	reminder, err := h.q.UpsertReminder(r.Context(), dbsqlc.UpsertReminderParams{
		UserID:   userID,
		Type:     req.Type,
		TimeHhmm: req.TimeHHMM,
		Enabled:  req.Enabled,
		Label:    req.Label,
	})
	if err != nil {
		response.Internal(w, "could not save reminder")
		return
	}
	response.OK(w, reminder)
}

func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		response.BadRequest(w, "q is required")
		return
	}
	users, err := h.q.SearchUsers(r.Context(), &q)
	if err != nil {
		response.Internal(w, "search failed")
		return
	}
	response.OK(w, users)
}
