package store

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"todo-app/internal/migrate"
)

func newTestStore(t *testing.T) *PostgresStore {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping PostgreSQL tests")
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()

	var name string
	if err := db.QueryRowContext(ctx, `SELECT current_database()`).Scan(&name); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	if !strings.HasSuffix(name, "_test") {
		t.Fatalf("refusing to run: database %q does not end in _test", name)
	}

	if err := migrate.Up(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if _, err := db.ExecContext(ctx, `TRUNCATE todos, categories RESTART IDENTITY`); err != nil {
		t.Fatalf("reset table: %v", err)
	}

	return NewPostgresStore(db)
}

func TestPostgresStoreContract(t *testing.T) {
	runContract(t, func(t *testing.T) Store { return newTestStore(t) })
}

func TestMigrateIsIdempotent(t *testing.T) {
	s := newTestStore(t)

	if err := migrate.Up(context.Background(), s.db); err != nil {
		t.Fatalf("running migrations a second time failed: %v", err)
	}
}
