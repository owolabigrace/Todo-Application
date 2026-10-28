package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

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

	if _, err := db.ExecContext(ctx, `TRUNCATE todos RESTART IDENTITY`); err != nil {
		t.Fatalf("reset table: %v", err)
	}

	return NewPostgresStore(db)
}

func TestPostgresCreateAndGet(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	created, err := s.Create(ctx, "Buy milk")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID != 1 || created.Title != "Buy milk" || created.Completed {
		t.Errorf("unexpected created todo: %+v", created)
	}
	if time.Since(created.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt looks wrong: %v", created.CreatedAt)
	}

	got, err := s.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "Buy milk" {
		t.Errorf("Title = %q, want %q", got.Title, "Buy milk")
	}
}

func TestPostgresListOrdered(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	items, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("empty list should be a non-nil empty slice, got %#v", items)
	}

	for _, title := range []string{"one", "two", "three"} {
		if _, err := s.Create(ctx, title); err != nil {
			t.Fatalf("Create %s: %v", title, err)
		}
	}

	items, err = s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 3 || items[0].Title != "one" || items[2].Title != "three" {
		t.Errorf("unexpected list: %+v", items)
	}
}

func TestPostgresUpdate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	created, _ := s.Create(ctx, "old")

	updated, err := s.Update(ctx, created.ID, "new", true)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "new" || !updated.Completed {
		t.Errorf("unexpected update result: %+v", updated)
	}
	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Error("Update must not change CreatedAt")
	}

	if _, err := s.Update(ctx, 9999, "x", false); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update missing id: err = %v, want ErrNotFound", err)
	}
}

func TestPostgresDelete(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	created, _ := s.Create(ctx, "temp")

	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after delete: err = %v, want ErrNotFound", err)
	}
	if err := s.Delete(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("second Delete: err = %v, want ErrNotFound", err)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	s := newTestStore(t)

	if err := migrate.Up(context.Background(), s.db); err != nil {
		t.Fatalf("running migrations a second time failed: %v", err)
	}
}
