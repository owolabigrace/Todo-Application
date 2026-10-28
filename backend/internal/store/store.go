package store

import (
	"context"
	"errors"

	"todo-app/internal/todo"
)

var ErrNotFound = errors.New("todo not found")

type Store interface {
	List(ctx context.Context, f todo.Filter) ([]todo.Todo, error)
	Get(ctx context.Context, id int) (todo.Todo, error)
	Create(ctx context.Context, in todo.Input) (todo.Todo, error)
	Update(ctx context.Context, id int, in todo.Input) (todo.Todo, error)
	Delete(ctx context.Context, id int) error
}
