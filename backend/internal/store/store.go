package store

import (
	"context"
	"errors"

	"todo-app/internal/todo"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrDuplicate       = errors.New("name already exists")
	ErrInvalidCategory = errors.New("category does not exist")
)

type Store interface {
	List(ctx context.Context, f todo.Filter) ([]todo.Todo, error)
	Get(ctx context.Context, id int) (todo.Todo, error)
	Create(ctx context.Context, in todo.Input) (todo.Todo, error)
	Update(ctx context.Context, id int, in todo.Input) (todo.Todo, error)
	Delete(ctx context.Context, id int) error

	ListCategories(ctx context.Context) ([]todo.Category, error)
	CreateCategory(ctx context.Context, name string) (todo.Category, error)
	UpdateCategory(ctx context.Context, id int, name string) (todo.Category, error)
	DeleteCategory(ctx context.Context, id int) error
}
