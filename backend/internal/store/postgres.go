package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"todo-app/internal/todo"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const queryTimeout = 5 * time.Second

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{
		db: db,
	}
}

func (s *PostgresStore) List(ctx context.Context) ([]todo.Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, completed, created_at
		FROM todos
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]todo.Todo, 0)

	for rows.Next() {
		var item todo.Todo

		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Completed,
			&item.CreatedAt,
		); err != nil {
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

	var item todo.Todo

	err := s.db.QueryRowContext(ctx, `
		SELECT id, title, completed, created_at
		FROM todos
		WHERE id = $1
	`, id).Scan(
		&item.ID,
		&item.Title,
		&item.Completed,
		&item.CreatedAt,
	)

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
	title string,
) (todo.Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var item todo.Todo

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO todos (title)
		VALUES ($1)
		RETURNING id, title, completed, created_at
	`, title).Scan(
		&item.ID,
		&item.Title,
		&item.Completed,
		&item.CreatedAt,
	)

	if err != nil {
		return todo.Todo{}, err
	}

	return item, nil
}

func (s *PostgresStore) Update(
	ctx context.Context,
	id int,
	title string,
	completed bool,
) (todo.Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var item todo.Todo

	err := s.db.QueryRowContext(ctx, `
		UPDATE todos
		SET title = $1, completed = $2
		WHERE id = $3
		RETURNING id, title, completed, created_at
	`, title, completed, id).Scan(
		&item.ID,
		&item.Title,
		&item.Completed,
		&item.CreatedAt,
	)

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
