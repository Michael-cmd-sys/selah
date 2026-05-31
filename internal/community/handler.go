package community

import (
	"context"
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
	q *dbsqlc.Queries
}

func NewHandler(q *dbsqlc.Queries) *Handler {
	return &Handler{q: q}
}

// isMember returns true if userID is a member of groupID.
func (h *Handler) isMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	return h.q.IsGroupMember(ctx, dbsqlc.IsGroupMemberParams{GroupID: groupID, UserID: userID})
}

// requireGroupAccess loads the group and, if it is private, verifies the caller is a member.
// Returns false and writes the response if access is denied.
func (h *Handler) requireGroupAccess(w http.ResponseWriter, r *http.Request, groupID, userID uuid.UUID) (dbsqlc.Group, bool) {
	group, err := h.q.GetGroupByID(r.Context(), groupID)
	if err != nil {
		response.NotFound(w, "group not found")
		return dbsqlc.Group{}, false
	}
	if group.IsPrivate {
		ok, err := h.isMember(r.Context(), groupID, userID)
		if err != nil || !ok {
			response.NotFound(w, "group not found")
			return dbsqlc.Group{}, false
		}
	}
	return group, true
}

// requirePostGroupAccess loads the post and, if it belongs to a private group,
// verifies the caller is a member. Returns false and writes the response if denied.
func (h *Handler) requirePostGroupAccess(w http.ResponseWriter, r *http.Request, postID, userID uuid.UUID) (dbsqlc.GetPostByIDRow, bool) {
	post, err := h.q.GetPostByID(r.Context(), postID)
	if err != nil {
		response.NotFound(w, "post not found")
		return dbsqlc.GetPostByIDRow{}, false
	}
	if post.GroupID.Valid {
		ok, err := h.isMember(r.Context(), post.GroupID.Bytes, userID)
		if err != nil || !ok {
			response.Forbidden(w, "not a member of this group")
			return dbsqlc.GetPostByIDRow{}, false
		}
	}
	return post, true
}

// --- Groups ---

type createGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsPrivate   bool    `json:"isPrivate"`
	AvatarURL   *string `json:"avatarUrl"`
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	group, err := h.q.CreateGroup(r.Context(), dbsqlc.CreateGroupParams{
		Name:        req.Name,
		Description: req.Description,
		CreatorID:   userID,
		IsPrivate:   req.IsPrivate,
		AvatarUrl:   req.AvatarURL,
	})
	if err != nil {
		response.Internal(w, "could not create group")
		return
	}

	_ = h.q.JoinGroup(r.Context(), dbsqlc.JoinGroupParams{GroupID: group.ID, UserID: userID})
	response.Created(w, group)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid group id")
		return
	}
	group, ok := h.requireGroupAccess(w, r, id, userID)
	if !ok {
		return
	}
	response.OK(w, group)
}

func (h *Handler) ListMyGroups(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	groups, err := h.q.ListUserGroups(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch groups")
		return
	}
	response.OK(w, groups)
}

func (h *Handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid group id")
		return
	}

	group, err := h.q.GetGroupByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "group not found")
		return
	}
	if group.IsPrivate {
		response.Forbidden(w, "this group is invite-only")
		return
	}

	if err := h.q.JoinGroup(r.Context(), dbsqlc.JoinGroupParams{GroupID: id, UserID: userID}); err != nil {
		response.Internal(w, "could not join group")
		return
	}
	response.NoContent(w)
}

func (h *Handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid group id")
		return
	}
	if err := h.q.LeaveGroup(r.Context(), dbsqlc.LeaveGroupParams{GroupID: id, UserID: userID}); err != nil {
		response.Internal(w, "could not leave group")
		return
	}
	response.NoContent(w)
}

// --- Challenges ---

type createChallengeRequest struct {
	GroupID     *string `json:"groupId"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Type        string  `json:"type"`
	Target      int32   `json:"target"`
	Unit        string  `json:"unit"`
	StartDate   string  `json:"startDate"` // "2006-01-02"
	EndDate     string  `json:"endDate"`   // "2006-01-02"
}

func (h *Handler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req createChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Title == "" || req.Type == "" {
		response.BadRequest(w, "title and type are required")
		return
	}

	groupID := pgtype.UUID{}
	if req.GroupID != nil {
		id, err := uuid.Parse(*req.GroupID)
		if err != nil {
			response.BadRequest(w, "invalid groupId")
			return
		}
		ok, err := h.isMember(r.Context(), id, userID)
		if err != nil || !ok {
			response.Forbidden(w, "not a member of this group")
			return
		}
		groupID = pgtype.UUID{Bytes: id, Valid: true}
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		response.BadRequest(w, "invalid startDate, use YYYY-MM-DD")
		return
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		response.BadRequest(w, "invalid endDate, use YYYY-MM-DD")
		return
	}

	challenge, err := h.q.CreateChallenge(r.Context(), dbsqlc.CreateChallengeParams{
		GroupID:     groupID,
		CreatorID:   userID,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Target:      req.Target,
		Unit:        req.Unit,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		response.Internal(w, "could not create challenge")
		return
	}
	response.Created(w, challenge)
}

func (h *Handler) ListActiveChallenges(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	challenges, err := h.q.ListActiveChallenges(r.Context(), userID)
	if err != nil {
		response.Internal(w, "could not fetch challenges")
		return
	}
	response.OK(w, challenges)
}

// --- Posts ---

type createPostRequest struct {
	GroupID   *string `json:"groupId"`
	PostType  string  `json:"postType"`
	Content   string  `json:"content"`
	VerseRef  *string `json:"verseRef"`
	VerseText *string `json:"verseText"`
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Content == "" {
		response.BadRequest(w, "content is required")
		return
	}

	groupID := pgtype.UUID{}
	if req.GroupID != nil {
		id, err := uuid.Parse(*req.GroupID)
		if err != nil {
			response.BadRequest(w, "invalid groupId")
			return
		}
		ok, err := h.isMember(r.Context(), id, userID)
		if err != nil || !ok {
			response.Forbidden(w, "not a member of this group")
			return
		}
		groupID = pgtype.UUID{Bytes: id, Valid: true}
	}

	post, err := h.q.CreatePost(r.Context(), dbsqlc.CreatePostParams{
		UserID:    userID,
		GroupID:   groupID,
		PostType:  req.PostType,
		Content:   req.Content,
		VerseRef:  req.VerseRef,
		VerseText: req.VerseText,
	})
	if err != nil {
		response.Internal(w, "could not create post")
		return
	}
	response.Created(w, post)
}

func (h *Handler) ListFeed(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var groupIDParam uuid.UUID
	if gid := r.URL.Query().Get("groupId"); gid != "" {
		id, err := uuid.Parse(gid)
		if err != nil {
			response.BadRequest(w, "invalid groupId")
			return
		}
		if _, ok := h.requireGroupAccess(w, r, id, userID); !ok {
			return
		}
		groupIDParam = id
	}

	posts, err := h.q.ListFeedPosts(r.Context(), dbsqlc.ListFeedPostsParams{
		Column1: groupIDParam,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		response.Internal(w, "could not fetch feed")
		return
	}
	response.OK(w, posts)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid post id")
		return
	}
	if err := h.q.DeletePost(r.Context(), dbsqlc.DeletePostParams{ID: id, UserID: userID}); err != nil {
		response.Internal(w, "could not delete post")
		return
	}
	response.NoContent(w)
}

// --- Reactions ---

type reactionRequest struct {
	Type string `json:"type"`
}

func (h *Handler) AddReaction(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid post id")
		return
	}

	if _, ok := h.requirePostGroupAccess(w, r, postID, userID); !ok {
		return
	}

	var req reactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.q.AddReaction(r.Context(), dbsqlc.AddReactionParams{
		PostID: postID, UserID: userID, ReactionType: req.Type,
	}); err != nil {
		response.Internal(w, "could not add reaction")
		return
	}
	response.NoContent(w)
}

func (h *Handler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid post id")
		return
	}

	if _, ok := h.requirePostGroupAccess(w, r, postID, userID); !ok {
		return
	}

	if err := h.q.RemoveReaction(r.Context(), dbsqlc.RemoveReactionParams{
		PostID: postID, UserID: userID, ReactionType: r.URL.Query().Get("type"),
	}); err != nil {
		response.Internal(w, "could not remove reaction")
		return
	}
	response.NoContent(w)
}

// --- Comments ---

type createCommentRequest struct {
	Content string `json:"content"`
}

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid post id")
		return
	}
	if _, ok := h.requirePostGroupAccess(w, r, postID, userID); !ok {
		return
	}
	comments, err := h.q.ListComments(r.Context(), postID)
	if err != nil {
		response.Internal(w, "could not fetch comments")
		return
	}
	response.OK(w, comments)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid post id")
		return
	}

	if _, ok := h.requirePostGroupAccess(w, r, postID, userID); !ok {
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Content == "" {
		response.BadRequest(w, "content is required")
		return
	}

	comment, err := h.q.CreateComment(r.Context(), dbsqlc.CreateCommentParams{
		PostID: postID, UserID: userID, Content: req.Content,
	})
	if err != nil {
		response.Internal(w, "could not create comment")
		return
	}
	response.Created(w, comment)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "commentId"))
	if err != nil {
		response.BadRequest(w, "invalid comment id")
		return
	}
	if err := h.q.DeleteComment(r.Context(), dbsqlc.DeleteCommentParams{ID: id, UserID: userID}); err != nil {
		response.Internal(w, "could not delete comment")
		return
	}
	response.NoContent(w)
}

func parseDate(s string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}
