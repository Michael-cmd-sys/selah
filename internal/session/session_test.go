package session_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/selah/internal/dbsqlc"
	"github.com/selah/internal/testutil"
)

func seedUser(t *testing.T, q *dbsqlc.Queries, email, name string) dbsqlc.User {
	t.Helper()
	u, err := q.CreateUser(context.Background(), dbsqlc.CreateUserParams{
		Email: &email,
		Name:  name,
	})
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return u
}

func TestCreateAndCompleteSession(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	user := seedUser(t, q, "meditator@example.com", "Meditator")
	userPG := pgtype.UUID{Bytes: user.ID, Valid: true}

	verseBook := "Psalms"
	verseChapter := int32(46)
	verseNumber := int32(10)
	verseText := "Be still, and know that I am God."
	verseRef := "Psalm 46:10"

	sess, err := q.CreateMeditationSession(ctx, dbsqlc.CreateMeditationSessionParams{
		UserID:          userPG,
		Title:           "Morning Silence",
		SessionType:     "meditation",
		DurationSeconds: 900,
		VerseBook:       &verseBook,
		VerseChapter:    &verseChapter,
		VerseNumber:     &verseNumber,
		VerseText:       &verseText,
		VerseRef:        &verseRef,
	})
	if err != nil {
		t.Fatalf("CreateMeditationSession: %v", err)
	}
	if sess.Completed {
		t.Fatal("session should not be completed on creation")
	}
	if sess.Title != "Morning Silence" {
		t.Errorf("want title Morning Silence, got %s", sess.Title)
	}

	note := "Great session"
	completed, err := q.CompleteMeditationSession(ctx, dbsqlc.CompleteMeditationSessionParams{
		ID:     sess.ID,
		Note:   &note,
		UserID: userPG,
	})
	if err != nil {
		t.Fatalf("CompleteMeditationSession: %v", err)
	}
	if !completed.Completed {
		t.Error("session should be completed")
	}
	if !completed.CompletedAt.Valid {
		t.Error("completedAt should be set")
	}
}

func TestGetSessionNotFound(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	_, err := q.GetMeditationSession(ctx, dbsqlc.GetMeditationSessionParams{
		ID:     uuid.New(),
		UserID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
	})
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
}

func TestListSessions(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	user := seedUser(t, q, "lister@example.com", "Lister")
	userPG := pgtype.UUID{Bytes: user.ID, Valid: true}

	for i := range 5 {
		dur := int32(300 * (i + 1))
		_, err := q.CreateMeditationSession(ctx, dbsqlc.CreateMeditationSessionParams{
			UserID:          userPG,
			Title:           "Session",
			SessionType:     "prayer",
			DurationSeconds: dur,
		})
		if err != nil {
			t.Fatalf("CreateMeditationSession %d: %v", i, err)
		}
	}

	sessions, err := q.ListMeditationSessions(ctx, dbsqlc.ListMeditationSessionsParams{
		UserID: userPG,
		Limit:  3,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListMeditationSessions: %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("want 3 sessions (limit), got %d", len(sessions))
	}
}
