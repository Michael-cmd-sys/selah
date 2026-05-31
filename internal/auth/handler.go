package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/selah/internal/dbsqlc"
	"github.com/selah/internal/response"
)

type Handler struct {
	q         *dbsqlc.Queries
	jwtSecret string
	jwtExpiry time.Duration
}

func NewHandler(q *dbsqlc.Queries, jwtSecret string, jwtExpiry time.Duration) *Handler {
	return &Handler{q: q, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type googleRequest struct {
	IDToken string `json:"idToken"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  publicUser `json:"user"`
}

type publicUser struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     *string   `json:"email,omitempty"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	Level     int32     `json:"level"`
	XP        int32     `json:"xp"`
}

func toPublicUser(u dbsqlc.User) publicUser {
	return publicUser{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarUrl,
		Level:     u.Level,
		XP:        u.Xp,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" || req.Email == "" {
		response.BadRequest(w, "name and email are required")
		return
	}
	if len(req.Password) < 8 {
		response.BadRequest(w, "password must be at least 8 characters")
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		response.Internal(w, "could not process password")
		return
	}

	user, err := h.q.CreateUser(r.Context(), dbsqlc.CreateUserParams{
		Email:         &req.Email,
		Name:          req.Name,
		AvatarUrl:     nil,
		GoogleSub:     nil,
		EmailVerified: false,
	})
	if err != nil {
		response.Error(w, http.StatusConflict, "email already registered")
		return
	}

	if err := h.q.CreateCredential(r.Context(), dbsqlc.CreateCredentialParams{
		UserID:       user.ID,
		PasswordHash: hash,
	}); err != nil {
		response.Internal(w, "could not save credentials")
		return
	}

	token, err := SignToken(user.ID, req.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		response.Internal(w, "could not create token")
		return
	}

	response.Created(w, authResponse{Token: token, User: toPublicUser(user)})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.q.GetUserByEmail(r.Context(), &req.Email)
	if err != nil {
		response.Unauthorized(w, "invalid credentials")
		return
	}

	hash, err := h.q.GetCredentialByUserID(r.Context(), user.ID)
	if err != nil {
		response.Unauthorized(w, "invalid credentials")
		return
	}

	ok, err := VerifyPassword(req.Password, hash)
	if err != nil || !ok {
		response.Unauthorized(w, "invalid credentials")
		return
	}

	token, err := SignToken(user.ID, req.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		response.Internal(w, "could not create token")
		return
	}

	response.OK(w, authResponse{Token: token, User: toPublicUser(user)})
}

func (h *Handler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	var req googleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.IDToken == "" {
		response.BadRequest(w, "idToken is required")
		return
	}

	info, err := VerifyGoogleIDToken(req.IDToken)
	if err != nil {
		response.Unauthorized(w, "invalid google token")
		return
	}

	var avatarURL *string
	if info.Picture != "" {
		avatarURL = &info.Picture
	}

	user, err := h.q.UpsertGoogleUser(r.Context(), dbsqlc.UpsertGoogleUserParams{
		GoogleSub: &info.Sub,
		Email:     &info.Email,
		Name:      info.Name,
		AvatarUrl: avatarURL,
	})
	if err != nil {
		response.Internal(w, "could not authenticate user")
		return
	}

	token, err := SignToken(user.ID, info.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		response.Internal(w, "could not create token")
		return
	}

	response.OK(w, authResponse{Token: token, User: toPublicUser(user)})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromCtx(r.Context())
	user, err := h.q.GetUserByID(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "user not found")
		return
	}
	response.OK(w, toPublicUser(user))
}

type ctxKey string

const ctxUserID ctxKey = "userID"

func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxUserID, id)
}

func UserIDFromCtx(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ctxUserID).(uuid.UUID)
	return id
}
