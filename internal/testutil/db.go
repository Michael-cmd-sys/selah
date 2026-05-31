package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/selah/internal/dbsqlc"
)

// NewTestDB creates an isolated test database, runs migrations, and returns
// a Queries handle. The database is dropped when the test ends.
func NewTestDB(t *testing.T) *dbsqlc.Queries {
	t.Helper()
	_ = godotenv.Load("../../.env")

	base := os.Getenv("DATABASE_URL")
	if base == "" {
		base = "postgres://selah:dev_password@localhost:5432/selah?sslmode=disable"
	}

	dbName := fmt.Sprintf("selah_test_%d", os.Getpid())
	adminDSN := base
	testDSN := replaceDSNDB(base, dbName)

	// Create test database
	adminCfg, err := pgx.ParseConfig(adminDSN)
	if err != nil {
		t.Fatalf("parse admin dsn: %v", err)
	}
	adminDB := stdlib.OpenDB(*adminCfg)
	if _, err := adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)); err != nil {
		t.Fatalf("create test db: %v", err)
	}
	adminDB.Close()

	// Run migrations
	migCfg, err := pgx.ParseConfig(testDSN)
	if err != nil {
		t.Fatalf("parse test dsn: %v", err)
	}
	migDB := stdlib.OpenDB(*migCfg)
	goose.SetDialect("postgres")
	if err := goose.Up(migDB, "../../db/migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	migDB.Close()

	// Connect with pgx pool
	pool, err := pgxpool.New(context.Background(), testDSN)
	if err != nil {
		t.Fatalf("pgx pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		if cleanupCfg, err := pgx.ParseConfig(adminDSN); err == nil {
			db := stdlib.OpenDB(*cleanupCfg)
			db.Exec(fmt.Sprintf(`DROP DATABASE "%s" WITH (FORCE)`, dbName))
			db.Close()
		}
	})

	return dbsqlc.New(pool)
}

// replaceDSNDB swaps the database name in a postgres DSN.
func replaceDSNDB(dsn, newDB string) string {
	// Works for both URL-style (postgres://user:pass@host/db?...) DSNs.
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '/' {
			base := dsn[:i+1]
			rest := dsn[i+1:]
			// strip existing db name before any query string
			for j, c := range rest {
				if c == '?' {
					return base + newDB + rest[j:]
				}
			}
			return base + newDB
		}
	}
	return dsn
}
