package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/selah/internal/auth"
	"github.com/selah/internal/bible"
	"github.com/selah/internal/community"
	"github.com/selah/internal/config"
	"github.com/selah/internal/dbsqlc"
	"github.com/selah/internal/fasting"
	apimw "github.com/selah/internal/middleware"
	"github.com/selah/internal/response"
	"github.com/selah/internal/session"
	"github.com/selah/internal/stats"
	"github.com/selah/internal/user"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("migrations", "err", err)
		os.Exit(1)
	}

	q := dbsqlc.New(pool)

	authH := auth.NewHandler(q, cfg.JWTSecret, cfg.JWTExpiry)
	userH := user.NewHandler(q)
	bibleH := bible.NewHandler(q)
	sessionH := session.NewHandler(q)
	fastingH := fasting.NewHandler(q)
	statsH := stats.NewHandler(q)
	communityH := community.NewHandler(q)

	authLimiter := apimw.NewRateLimiter(10, time.Minute)

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(apimw.Logger)
	r.Use(corsMiddleware)
	r.Use(apimw.MaxBodyBytes(1 << 20)) // 1 MB

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		response.OK(w, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		// Auth — public (rate limited: 10 req/min per IP)
		r.Group(func(r chi.Router) {
			r.Use(authLimiter.Handler)
			r.Post("/auth/register", authH.Register)
			r.Post("/auth/login", authH.Login)
			r.Post("/auth/google", authH.GoogleAuth)
		})

		// Auth — protected
		r.Group(func(r chi.Router) {
			r.Use(apimw.Authenticate(cfg.JWTSecret))
			r.Get("/auth/me", authH.Me)
		})

		// Bible — optional auth (highlights/bookmarks need auth, reading is public)
		r.Route("/bible", func(r chi.Router) {
			r.Use(apimw.OptionalAuth(cfg.JWTSecret))
			r.Get("/translations", bibleH.ListTranslations)
			r.Get("/daily-verse", bibleH.GetDailyVerse)
			r.Get("/{translation}/books", bibleH.ListBooks)
			r.Get("/{translation}/{book}/{chapter}", bibleH.GetChapter)
			r.Get("/{translation}/search", bibleH.Search)

			// Highlights & bookmarks require auth
			r.Group(func(r chi.Router) {
				r.Use(apimw.Authenticate(cfg.JWTSecret))
				r.Get("/highlights", bibleH.GetHighlights)
				r.Put("/highlights", bibleH.UpsertHighlight)
				r.Delete("/highlights/{id}", bibleH.DeleteHighlight)
				r.Get("/bookmarks", bibleH.GetBookmarks)
				r.Post("/bookmarks", bibleH.CreateBookmark)
				r.Delete("/bookmarks/{id}", bibleH.DeleteBookmark)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(apimw.Authenticate(cfg.JWTSecret))

			// User
			r.Get("/users/me", userH.GetProfile)
			r.Put("/users/me", userH.UpdateProfile)
			r.Get("/users/me/goals", userH.GetGoals)
			r.Put("/users/me/goals", userH.UpsertGoal)
			r.Delete("/users/me/goals/{type}", userH.DeleteGoal)
			r.Get("/users/me/reminders", userH.GetReminders)
			r.Put("/users/me/reminders", userH.UpsertReminder)
			r.Get("/users/search", userH.SearchUsers)

			// Meditation sessions
			r.Post("/sessions", sessionH.Create)
			r.Put("/sessions/{id}/complete", sessionH.Complete)
			r.Get("/sessions/{id}", sessionH.Get)
			r.Get("/sessions", sessionH.List)

			// Fasting
			r.Get("/fasting/current", fastingH.GetCurrent)
			r.Post("/fasting", fastingH.Start)
			r.Put("/fasting/{id}/end", fastingH.End)
			r.Post("/fasting/{id}/checkin", fastingH.AddCheckin)
			r.Get("/fasting/{id}", fastingH.GetFast)
			r.Get("/fasting", fastingH.ListFasts)

			// Stats
			r.Get("/stats", statsH.GetJourney)
			r.Get("/stats/consistency", statsH.GetConsistency)
			r.Get("/stats/milestones", statsH.GetMilestones)

			// Community
			r.Get("/community/feed", communityH.ListFeed)
			r.Post("/community/posts", communityH.CreatePost)
			r.Delete("/community/posts/{id}", communityH.DeletePost)
			r.Post("/community/posts/{id}/reactions", communityH.AddReaction)
			r.Delete("/community/posts/{id}/reactions", communityH.RemoveReaction)
			r.Get("/community/posts/{id}/comments", communityH.ListComments)
			r.Post("/community/posts/{id}/comments", communityH.CreateComment)
			r.Delete("/community/posts/{id}/comments/{commentId}", communityH.DeleteComment)
			r.Get("/community/groups", communityH.ListMyGroups)
			r.Post("/community/groups", communityH.CreateGroup)
			r.Get("/community/groups/{id}", communityH.GetGroup)
			r.Post("/community/groups/{id}/join", communityH.JoinGroup)
			r.Post("/community/groups/{id}/leave", communityH.LeaveGroup)
			r.Get("/community/challenges", communityH.ListActiveChallenges)
			r.Post("/community/challenges", communityH.CreateChallenge)
		})
	})

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("listening", "addr", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "err", err)
	}
	slog.Info("server stopped")
}

func runMigrations(dsn string) error {
	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse migration dsn: %w", err)
	}
	db := stdlib.OpenDB(*connConfig)
	defer db.Close()

	goose.SetBaseFS(nil)
	goose.SetDialect("postgres")
	if err := goose.Up(db, "db/migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

