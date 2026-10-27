package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"todo-app/internal/todo"
)

type MemoryStore struct {
	mu     sync.RWMutex
	todos  map[int]todo.Todo
	nextID int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		todos:  make(map[int]todo.Todo),
		nextID: 1,
	}
}

func (s *MemoryStore) List(ctx context.Context) ([]todo.Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]todo.Todo, 0, len(s.todos))

	for _, item := range s.todos {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func (s *MemoryStore) Get(ctx context.Context, id int) (todo.Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.todos[id]
	if !exists {
		return todo.Todo{}, ErrNotFound
	}

	return item, nil
}

func (s *MemoryStore) Create(ctx context.Context, title string) (todo.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := todo.Todo{
		ID:        s.nextID,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now().UTC(),
	}

	s.todos[item.ID] = item
	s.nextID++

	return item, nil
}

func (s *MemoryStore) Update(
	ctx context.Context,
	id int,
	title string,
	completed bool,
) (todo.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.todos[id]
	if !exists {
		return todo.Todo{}, ErrNotFound
	}

	item.Title = title
	item.Completed = completed

	s.todos[id] = item

	return item, nil
}

func (s *MemoryStore) Delete(ctx context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.todos[id]; !exists {
		return ErrNotFound
	}

	delete(s.todos, id)

	return nil
}

var _ Store = (*MemoryStore)(nil)
