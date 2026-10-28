package store

import (
	"context"
	"sort"
	"strings"
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

func copyString(s *string) *string {
	if s == nil {
		return nil
	}

	c := *s

	return &c
}

func priorityRank(p string) int {
	switch p {
	case todo.PriorityHigh:
		return 3
	case todo.PriorityMedium:
		return 2
	default:
		return 1
	}
}

func matches(item todo.Todo, f todo.Filter) bool {
	switch f.Status {
	case "active":
		if item.Completed {
			return false
		}
	case "completed":
		if !item.Completed {
			return false
		}
	}

	if f.Priority != "" && item.Priority != f.Priority {
		return false
	}

	if f.Search != "" {
		q := strings.ToLower(f.Search)
		if !strings.Contains(strings.ToLower(item.Title), q) &&
			!strings.Contains(strings.ToLower(item.Notes), q) {
			return false
		}
	}

	return true
}

func sortTodos(items []todo.Todo, key, order string) {
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })

	desc := order == "desc"

	switch key {
	case "due":
		sort.SliceStable(items, func(i, j int) bool {
			a, b := items[i].DueDate, items[j].DueDate

			switch {
			case a == nil:
				return false // no due date always sorts last
			case b == nil:
				return true
			case desc:
				return *a > *b
			default:
				return *a < *b
			}
		})
	case "priority":
		sort.SliceStable(items, func(i, j int) bool {
			a, b := priorityRank(items[i].Priority), priorityRank(items[j].Priority)
			if desc {
				return a > b
			}

			return a < b
		})
	case "title":
		sort.SliceStable(items, func(i, j int) bool {
			a, b := strings.ToLower(items[i].Title), strings.ToLower(items[j].Title)
			if desc {
				return a > b
			}

			return a < b
		})
	default:
		if desc {
			sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
		}
	}
}

func (s *MemoryStore) List(ctx context.Context, f todo.Filter) ([]todo.Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]todo.Todo, 0, len(s.todos))

	for _, item := range s.todos {
		if matches(item, f) {
			result = append(result, item)
		}
	}

	sortTodos(result, f.Sort, f.Order)

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

func (s *MemoryStore) Create(ctx context.Context, in todo.Input) (todo.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := todo.Todo{
		ID:        s.nextID,
		Title:     in.Title,
		Notes:     in.Notes,
		Completed: in.Completed,
		Priority:  in.Priority,
		DueDate:   copyString(in.DueDate),
		CreatedAt: time.Now().UTC(),
	}

	s.todos[item.ID] = item
	s.nextID++

	return item, nil
}

func (s *MemoryStore) Update(
	ctx context.Context,
	id int,
	in todo.Input,
) (todo.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.todos[id]
	if !exists {
		return todo.Todo{}, ErrNotFound
	}

	item.Title = in.Title
	item.Notes = in.Notes
	item.Completed = in.Completed
	item.Priority = in.Priority
	item.DueDate = copyString(in.DueDate)

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
