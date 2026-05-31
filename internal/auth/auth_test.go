package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/selah/internal/auth"
	"github.com/selah/internal/dbsqlc"
	"github.com/selah/internal/testutil"
)

func sp(s string) *string { return &s }

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := auth.HashPassword("strongpassword")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	ok, err := auth.VerifyPassword("strongpassword", hash)
	if err != nil || !ok {
		t.Fatalf("correct password should verify: err=%v ok=%v", err, ok)
	}

	wrong, _ := auth.VerifyPassword("wrongpassword", hash)
	if wrong {
		t.Fatal("wrong password must not verify")
	}
}

func TestJWTSignAndParse(t *testing.T) {
	secret := "jwt-test-secret"
	userID := uuid.New()

	token, err := auth.SignToken(userID, "test@example.com", secret, time.Hour)
	if err != nil {
		t.Fatalf("SignToken: %v", err)
	}

	claims, err := auth.ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("userID mismatch: want %v got %v", userID, claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("email mismatch: want test@example.com got %s", claims.Email)
	}

	_, err = auth.ParseToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("wrong secret must return error")
	}
}

func TestCreateUserAndCredentials(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	user, err := q.CreateUser(ctx, dbsqlc.CreateUserParams{
		Email: sp("alice@example.com"),
		Name:  "Alice",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.Name != "Alice" {
		t.Errorf("want name Alice, got %s", user.Name)
	}

	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := q.CreateCredential(ctx, dbsqlc.CreateCredentialParams{
		UserID:       user.ID,
		PasswordHash: hash,
	}); err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}

	stored, err := q.GetCredentialByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetCredentialByUserID: %v", err)
	}

	ok, err := auth.VerifyPassword("secret123", stored)
	if err != nil || !ok {
		t.Fatalf("stored password should verify: err=%v ok=%v", err, ok)
	}
}

func TestGetUserByEmail(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	_, err := q.CreateUser(ctx, dbsqlc.CreateUserParams{
		Email: sp("bob@example.com"),
		Name:  "Bob",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	user, err := q.GetUserByEmail(ctx, sp("bob@example.com"))
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if user.Name != "Bob" {
		t.Errorf("want Bob, got %s", user.Name)
	}
}

func TestUpsertGoogleUser(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	params := dbsqlc.UpsertGoogleUserParams{
		GoogleSub: sp("google-sub-xyz"),
		Email:     sp("carol@gmail.com"),
		Name:      "Carol",
		AvatarUrl: sp("https://example.com/avatar.jpg"),
	}

	u1, err := q.UpsertGoogleUser(ctx, params)
	if err != nil {
		t.Fatalf("UpsertGoogleUser insert: %v", err)
	}

	u2, err := q.UpsertGoogleUser(ctx, params)
	if err != nil {
		t.Fatalf("UpsertGoogleUser update: %v", err)
	}
	if u1.ID != u2.ID {
		t.Error("upsert must return the same user ID on conflict")
	}
}
