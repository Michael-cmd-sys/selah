package fasting_test

import (
	"context"
	"testing"

	"github.com/selah/internal/dbsqlc"
	"github.com/selah/internal/testutil"
)

func sp(s string) *string { return &s }

func seedUser(t *testing.T, q *dbsqlc.Queries, email, name string) dbsqlc.User {
	t.Helper()
	user, err := q.CreateUser(context.Background(), dbsqlc.CreateUserParams{
		Email: sp(email),
		Name:  name,
	})
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return user
}

func TestStartAndEndFast(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	user := seedUser(t, q, "faster@example.com", "Faster")

	fast, err := q.StartFast(ctx, dbsqlc.StartFastParams{
		UserID:      user.ID,
		TargetHours: 16,
	})
	if err != nil {
		t.Fatalf("StartFast: %v", err)
	}
	if fast.EndedAt.Valid {
		t.Error("fast should not be ended on creation")
	}

	active, err := q.GetActiveFast(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetActiveFast: %v", err)
	}
	if active.ID != fast.ID {
		t.Error("active fast ID mismatch")
	}

	ended, err := q.EndFast(ctx, dbsqlc.EndFastParams{ID: fast.ID, UserID: user.ID})
	if err != nil {
		t.Fatalf("EndFast: %v", err)
	}
	if !ended.EndedAt.Valid {
		t.Error("fast should be marked ended")
	}

	_, err = q.GetActiveFast(ctx, user.ID)
	if err == nil {
		t.Error("expected no active fast after ending")
	}
}

func TestFastCheckins(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	user := seedUser(t, q, "checkin@example.com", "CheckIn")

	fast, err := q.StartFast(ctx, dbsqlc.StartFastParams{UserID: user.ID, TargetHours: 24})
	if err != nil {
		t.Fatalf("StartFast: %v", err)
	}

	checkin, err := q.AddCheckin(ctx, dbsqlc.AddCheckinParams{
		FastID:     fast.ID,
		UserID:     user.ID,
		Mood:       "peaceful",
		Reflection: sp("Feeling close to God"),
	})
	if err != nil {
		t.Fatalf("AddCheckin: %v", err)
	}
	if checkin.Mood != "peaceful" {
		t.Errorf("want mood peaceful, got %s", checkin.Mood)
	}

	latest, err := q.GetLatestCheckin(ctx, fast.ID)
	if err != nil {
		t.Fatalf("GetLatestCheckin: %v", err)
	}
	if latest.ID != checkin.ID {
		t.Error("latest checkin ID mismatch")
	}
}

func TestListFasts(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	user := seedUser(t, q, "list@example.com", "Lister")

	for range 3 {
		f, _ := q.StartFast(ctx, dbsqlc.StartFastParams{UserID: user.ID, TargetHours: 16})
		q.EndFast(ctx, dbsqlc.EndFastParams{ID: f.ID, UserID: user.ID})
	}

	fasts, err := q.ListFasts(ctx, dbsqlc.ListFastsParams{UserID: user.ID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListFasts: %v", err)
	}
	if len(fasts) != 3 {
		t.Errorf("want 3 fasts, got %d", len(fasts))
	}
}
