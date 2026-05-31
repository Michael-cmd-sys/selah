package community_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/selah/internal/dbsqlc"
	"github.com/selah/internal/testutil"
)

func sp(s string) *string { return &s }

func seedUser(t *testing.T, q *dbsqlc.Queries, email, name string) dbsqlc.User {
	t.Helper()
	u, err := q.CreateUser(context.Background(), dbsqlc.CreateUserParams{
		Email: sp(email),
		Name:  name,
	})
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return u
}

func TestCreateGroupAndJoin(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	creator := seedUser(t, q, "creator@example.com", "Creator")
	member := seedUser(t, q, "member@example.com", "Member")

	group, err := q.CreateGroup(ctx, dbsqlc.CreateGroupParams{
		Name:        "Morning Warriors",
		Description: sp("Early risers"),
		CreatorID:   creator.ID,
		IsPrivate:   false,
	})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if group.Name != "Morning Warriors" {
		t.Errorf("want Morning Warriors, got %s", group.Name)
	}

	if err := q.JoinGroup(ctx, dbsqlc.JoinGroupParams{GroupID: group.ID, UserID: creator.ID}); err != nil {
		t.Fatalf("JoinGroup (creator): %v", err)
	}
	if err := q.JoinGroup(ctx, dbsqlc.JoinGroupParams{GroupID: group.ID, UserID: member.ID}); err != nil {
		t.Fatalf("JoinGroup (member): %v", err)
	}

	isMember, err := q.IsGroupMember(ctx, dbsqlc.IsGroupMemberParams{GroupID: group.ID, UserID: member.ID})
	if err != nil {
		t.Fatalf("IsGroupMember: %v", err)
	}
	if !isMember {
		t.Error("member should be in group")
	}

	if err := q.LeaveGroup(ctx, dbsqlc.LeaveGroupParams{GroupID: group.ID, UserID: member.ID}); err != nil {
		t.Fatalf("LeaveGroup: %v", err)
	}
	isMember, _ = q.IsGroupMember(ctx, dbsqlc.IsGroupMemberParams{GroupID: group.ID, UserID: member.ID})
	if isMember {
		t.Error("member should have left group")
	}
}

func TestCreatePostAndReactions(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	author := seedUser(t, q, "author@example.com", "Author")
	reactor := seedUser(t, q, "reactor@example.com", "Reactor")

	post, err := q.CreatePost(ctx, dbsqlc.CreatePostParams{
		UserID:   author.ID,
		GroupID:  pgtype.UUID{},
		PostType: "praise",
		Content:  "Finally hit a 7-day streak! God is good!",
	})
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post.PostType != "praise" {
		t.Errorf("want post_type praise, got %s", post.PostType)
	}

	if err := q.AddReaction(ctx, dbsqlc.AddReactionParams{
		PostID: post.ID, UserID: reactor.ID, ReactionType: "amen",
	}); err != nil {
		t.Fatalf("AddReaction amen: %v", err)
	}
	if err := q.AddReaction(ctx, dbsqlc.AddReactionParams{
		PostID: post.ID, UserID: reactor.ID, ReactionType: "praying",
	}); err != nil {
		t.Fatalf("AddReaction praying: %v", err)
	}

	// Duplicate reaction is idempotent
	if err := q.AddReaction(ctx, dbsqlc.AddReactionParams{
		PostID: post.ID, UserID: reactor.ID, ReactionType: "amen",
	}); err != nil {
		t.Fatalf("duplicate reaction should not error: %v", err)
	}

	reactions, err := q.GetReactionsByPost(ctx, post.ID)
	if err != nil {
		t.Fatalf("GetReactionsByPost: %v", err)
	}
	if len(reactions) != 2 {
		t.Errorf("want 2 reaction types, got %d", len(reactions))
	}

	if err := q.RemoveReaction(ctx, dbsqlc.RemoveReactionParams{
		PostID: post.ID, UserID: reactor.ID, ReactionType: "amen",
	}); err != nil {
		t.Fatalf("RemoveReaction: %v", err)
	}

	reactions, _ = q.GetReactionsByPost(ctx, post.ID)
	if len(reactions) != 1 {
		t.Errorf("want 1 reaction after removal, got %d", len(reactions))
	}
}

func TestCreateCommentAndDelete(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	user := seedUser(t, q, "commenter@example.com", "Commenter")

	post, err := q.CreatePost(ctx, dbsqlc.CreatePostParams{
		UserID:   user.ID,
		GroupID:  pgtype.UUID{},
		PostType: "prayer_request",
		Content:  "Please pray for my interview",
	})
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	comment, err := q.CreateComment(ctx, dbsqlc.CreateCommentParams{
		PostID:  post.ID,
		UserID:  user.ID,
		Content: "Praying for you!",
	})
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	comments, err := q.ListComments(ctx, post.ID)
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if len(comments) != 1 {
		t.Errorf("want 1 comment, got %d", len(comments))
	}

	if err := q.DeleteComment(ctx, dbsqlc.DeleteCommentParams{ID: comment.ID, UserID: user.ID}); err != nil {
		t.Fatalf("DeleteComment: %v", err)
	}

	comments, _ = q.ListComments(ctx, post.ID)
	if len(comments) != 0 {
		t.Error("comment should be deleted")
	}
}
