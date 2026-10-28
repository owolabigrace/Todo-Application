package store

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"todo-app/internal/todo"
)

type factory func(t *testing.T) Store

func TestMemoryStoreContract(t *testing.T) {
	runContract(t, func(t *testing.T) Store { return NewMemoryStore() })
}

func str(s string) *string { return &s }

func ids(items []todo.Todo) []int {
	out := make([]int, len(items))
	for i, it := range items {
		out[i] = it.ID
	}

	return out
}

// runContract checks behavior every Store implementation must share.
// Each subtest gets a fresh, empty store whose first ID is 1.
func runContract(t *testing.T, newStore factory) {
	ctx := context.Background()

	t.Run("create and get", func(t *testing.T) {
		s := newStore(t)

		created, err := s.Create(ctx, todo.Input{
			Title:    "Buy milk",
			Notes:    "2 litres",
			Priority: todo.PriorityHigh,
			DueDate:  str("2026-12-01"),
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		if created.ID != 1 || created.Title != "Buy milk" || created.Notes != "2 litres" ||
			created.Completed || created.Priority != todo.PriorityHigh {
			t.Errorf("unexpected created todo: %+v", created)
		}
		if created.DueDate == nil || *created.DueDate != "2026-12-01" {
			t.Errorf("DueDate = %v, want 2026-12-01", created.DueDate)
		}
		if time.Since(created.CreatedAt) > time.Minute {
			t.Errorf("CreatedAt looks wrong: %v", created.CreatedAt)
		}

		got, err := s.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.Title != "Buy milk" || got.DueDate == nil || *got.DueDate != "2026-12-01" {
			t.Errorf("Get returned %+v", got)
		}

		if _, err := s.Get(ctx, 9999); !errors.Is(err, ErrNotFound) {
			t.Errorf("Get missing id: err = %v, want ErrNotFound", err)
		}
	})

	t.Run("no due date stays nil", func(t *testing.T) {
		s := newStore(t)

		created, err := s.Create(ctx, todo.Input{Title: "x", Priority: todo.PriorityLow})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if created.DueDate != nil {
			t.Errorf("DueDate = %v, want nil", *created.DueDate)
		}
	})

	t.Run("empty list is non-nil", func(t *testing.T) {
		s := newStore(t)

		items, err := s.List(ctx, todo.Filter{})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("want non-nil empty slice, got %#v", items)
		}
	})

	t.Run("update", func(t *testing.T) {
		s := newStore(t)

		created, _ := s.Create(ctx, todo.Input{
			Title:    "old",
			Priority: todo.PriorityLow,
			DueDate:  str("2026-12-01"),
		})

		updated, err := s.Update(ctx, created.ID, todo.Input{
			Title:     "new",
			Notes:     "details",
			Completed: true,
			Priority:  todo.PriorityHigh,
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		if updated.Title != "new" || updated.Notes != "details" || !updated.Completed ||
			updated.Priority != todo.PriorityHigh {
			t.Errorf("unexpected update result: %+v", updated)
		}
		if updated.DueDate != nil {
			t.Errorf("DueDate should be cleared, got %v", *updated.DueDate)
		}
		if !updated.CreatedAt.Equal(created.CreatedAt) {
			t.Error("Update must not change CreatedAt")
		}

		if _, err := s.Update(ctx, 9999, todo.Input{Title: "x", Priority: todo.PriorityLow}); !errors.Is(err, ErrNotFound) {
			t.Errorf("Update missing id: err = %v, want ErrNotFound", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		s := newStore(t)

		created, _ := s.Create(ctx, todo.Input{Title: "temp", Priority: todo.PriorityMedium})

		if err := s.Delete(ctx, created.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := s.Get(ctx, created.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("Get after delete: err = %v, want ErrNotFound", err)
		}
		if err := s.Delete(ctx, created.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("second Delete: err = %v, want ErrNotFound", err)
		}
	})

	t.Run("filter search and sort", func(t *testing.T) {
		s := newStore(t)

		create := func(in todo.Input) todo.Todo {
			item, err := s.Create(ctx, in)
			if err != nil {
				t.Fatalf("seed Create: %v", err)
			}

			return item
		}

		create(todo.Input{Title: "Write report", Notes: "quarterly numbers", Priority: todo.PriorityHigh, DueDate: str("2026-11-01")})       // 1
		create(todo.Input{Title: "Buy milk", Priority: todo.PriorityLow})                                                                    // 2
		dave := create(todo.Input{Title: "Call Dave", Notes: "about the report", Priority: todo.PriorityMedium, DueDate: str("2026-10-30")}) // 3
		create(todo.Input{Title: "Save 50% off", Priority: todo.PriorityMedium, DueDate: str("2026-12-15")})                                 // 4
		create(todo.Input{Title: "Plan trip_2027", Priority: todo.PriorityLow})                                                              // 5

		if _, err := s.Update(ctx, dave.ID, todo.Input{
			Title:     "Call Dave",
			Notes:     "about the report",
			Completed: true,
			Priority:  todo.PriorityMedium,
			DueDate:   str("2026-10-30"),
		}); err != nil {
			t.Fatalf("seed Update: %v", err)
		}

		tests := []struct {
			name string
			f    todo.Filter
			want []int
		}{
			{"no filter", todo.Filter{}, []int{1, 2, 3, 4, 5}},
			{"status all", todo.Filter{Status: "all"}, []int{1, 2, 3, 4, 5}},
			{"active", todo.Filter{Status: "active"}, []int{1, 2, 4, 5}},
			{"completed", todo.Filter{Status: "completed"}, []int{3}},
			{"priority high", todo.Filter{Priority: "high"}, []int{1}},
			{"active and low", todo.Filter{Status: "active", Priority: "low"}, []int{2, 5}},
			{"search title and notes", todo.Filter{Search: "report"}, []int{1, 3}},
			{"search ignores case", todo.Filter{Search: "REPORT"}, []int{1, 3}},
			{"search percent is literal", todo.Filter{Search: "%"}, []int{4}},
			{"search underscore is literal", todo.Filter{Search: "_"}, []int{5}},
			{"search no match", todo.Filter{Search: "zzz"}, []int{}},
			{"created desc", todo.Filter{Sort: "created", Order: "desc"}, []int{5, 4, 3, 2, 1}},
			{"due asc, none last", todo.Filter{Sort: "due"}, []int{3, 1, 4, 2, 5}},
			{"due desc, none last", todo.Filter{Sort: "due", Order: "desc"}, []int{4, 1, 3, 2, 5}},
			{"priority desc", todo.Filter{Sort: "priority", Order: "desc"}, []int{1, 3, 4, 2, 5}},
			{"priority asc", todo.Filter{Sort: "priority", Order: "asc"}, []int{2, 5, 3, 4, 1}},
			{"title asc", todo.Filter{Sort: "title"}, []int{2, 3, 5, 4, 1}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				items, err := s.List(ctx, tc.f)
				if err != nil {
					t.Fatalf("List: %v", err)
				}

				if got := ids(items); !slices.Equal(got, tc.want) {
					t.Errorf("ids = %v, want %v", got, tc.want)
				}
			})
		}
	})
}
