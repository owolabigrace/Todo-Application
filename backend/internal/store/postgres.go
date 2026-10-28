package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"todo-app/internal/todo"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const queryTimeout = 5 * time.Second

const todoColumns = `id, title, notes, completed, priority, to_char(due_date, 'YYYY-MM-DD'), created_at`

// Escapes LIKE wildcards so a search for "50%" matches a literal percent sign.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{
		db: db,
	}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTodo(s scanner) (todo.Todo, error) {
	var item todo.Todo
	var due sql.NullString

	if err := s.Scan(
		&item.ID,
		&item.Title,
		&item.Notes,
		&item.Completed,
		&item.Priority,
		&due,
		&item.CreatedAt,
	); err != nil {
		return todo.Todo{}, err
	}

	if due.Valid {
		d := due.String
		item.DueDate = &d
	}

	return item, nil
}

// orderClause only ever returns fixed strings, never user input.
func orderClause(f todo.Filter) string {
	dir := "ASC"
	if f.Order == "desc" {
		dir = "DESC"
	}

	switch f.Sort {
	case "due":
		return "due_date " + dir + " NULLS LAST, id ASC"
	case "priority":
		return "CASE priority WHEN 'high' THEN 3 WHEN 'medium' THEN 2 ELSE 1 END " + dir + ", id ASC"
	case "title":
		return "lower(title) " + dir + ", id ASC"
	default:
		return "id " + dir
	}
}

func buildListQuery(f todo.Filter) (string, []any) {
	var conds []string
	var args []any

	switch f.Status {
	case "active":
		conds = append(conds, "completed = FALSE")
	case "completed":
		conds = append(conds, "completed = TRUE")
	}

	if f.Priority != "" {
		args = append(args, f.Priority)
		conds = append(conds, fmt.Sprintf("priority = $%d", len(args)))
	}

	if f.Search != "" {
		args = append(args, "%"+likeEscaper.Replace(f.Search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf("(title ILIKE $%d OR notes ILIKE $%d)", n, n))
	}

	query := "SELECT " + todoColumns + " FROM todos"
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}

	query += " ORDER BY " + orderClause(f)

	return query, args
}

func (s *PostgresStore) List(ctx context.Context, f todo.Filter) ([]todo.Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	query, args := buildListQuery(f)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]todo.Todo, 0)

	for rows.Next() {
		item, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PostgresStore) Get(
	ctx context.Context,
	id int,
) (todo.Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	item, err := scanTodo(s.db.QueryRowContext(ctx,
		`SELECT `+todoColumns+` FROM todos WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		return todo.Todo{}, ErrNotFound
	}

	if err != nil {
		return todo.Todo{}, err
	}

	return item, nil
}

func (s *PostgresStore) Create(
	ctx context.Context,
	in todo.Input,
) (todo.Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	item, err := scanTodo(s.db.QueryRowContext(ctx, `
		INSERT INTO todos (title, notes, completed, priority, due_date)
		VALUES ($1, $2, $3, $4, $5::text::date)
		RETURNING `+todoColumns,
		in.Title, in.Notes, in.Completed, in.Priority, in.DueDate,
	))
	if err != nil {
		return todo.Todo{}, err
	}

	return item, nil
}

func (s *PostgresStore) Update(
	ctx context.Context,
	id int,
	in todo.Input,
) (todo.Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	item, err := scanTodo(s.db.QueryRowContext(ctx, `
		UPDATE todos
		SET title = $1, notes = $2, completed = $3, priority = $4, due_date = $5::text::date
		WHERE id = $6
		RETURNING `+todoColumns,
		in.Title, in.Notes, in.Completed, in.Priority, in.DueDate, id,
	))

	if errors.Is(err, sql.ErrNoRows) {
		return todo.Todo{}, ErrNotFound
	}

	if err != nil {
		return todo.Todo{}, err
	}

	return item, nil
}

func (s *PostgresStore) Delete(
	ctx context.Context,
	id int,
) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `
		DELETE FROM todos
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

var _ Store = (*PostgresStore)(nil)
