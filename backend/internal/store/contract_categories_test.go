package store

import (
	"context"
	"errors"
	"slices"
	"testing"

	"todo-app/internal/todo"
)

func intPtr(n int) *int { return &n }

func TestMemoryCategoryContract(t *testing.T) {
	runCategoryContract(t, func(t *testing.T) Store { return NewMemoryStore() })
}

func TestPostgresCategoryContract(t *testing.T) {
	runCategoryContract(t, func(t *testing.T) Store { return newTestStore(t) })
}

func categoryNames(items []todo.Category) []string {
	out := make([]string, len(items))
	for i, c := range items {
		out[i] = c.Name
	}

	return out
}

func runCategoryContract(t *testing.T, newStore factory) {
	ctx := context.Background()

	t.Run("crud and ordering", func(t *testing.T) {
		s := newStore(t)

		items, err := s.ListCategories(ctx)
		if err != nil {
			t.Fatalf("ListCategories: %v", err)
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("want non-nil empty slice, got %#v", items)
		}

		work, err := s.CreateCategory(ctx, "Work")
		if err != nil {
			t.Fatalf("CreateCategory: %v", err)
		}
		if work.ID != 1 || work.Name != "Work" || work.CreatedAt.IsZero() {
			t.Errorf("unexpected category: %+v", work)
		}

		if _, err := s.CreateCategory(ctx, "home"); err != nil {
			t.Fatalf("CreateCategory: %v", err)
		}

		items, _ = s.ListCategories(ctx)
		if got := categoryNames(items); !slices.Equal(got, []string{"home", "Work"}) {
			t.Errorf("names = %v, want [home Work]", got)
		}

		renamed, err := s.UpdateCategory(ctx, work.ID, "WORK")
		if err != nil {
			t.Fatalf("UpdateCategory to different case: %v", err)
		}
		if renamed.Name != "WORK" {
			t.Errorf("Name = %q, want WORK", renamed.Name)
		}

		if _, err := s.UpdateCategory(ctx, 9999, "x"); !errors.Is(err, ErrNotFound) {
			t.Errorf("Update missing: err = %v, want ErrNotFound", err)
		}

		if err := s.DeleteCategory(ctx, work.ID); err != nil {
			t.Fatalf("DeleteCategory: %v", err)
		}
		if err := s.DeleteCategory(ctx, work.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("second Delete: err = %v, want ErrNotFound", err)
		}
	})

	t.Run("duplicate names", func(t *testing.T) {
		s := newStore(t)

		if _, err := s.CreateCategory(ctx, "Work"); err != nil {
			t.Fatalf("CreateCategory: %v", err)
		}
		if _, err := s.CreateCategory(ctx, "work"); !errors.Is(err, ErrDuplicate) {
			t.Errorf("duplicate create: err = %v, want ErrDuplicate", err)
		}

		home, err := s.CreateCategory(ctx, "Home")
		if err != nil {
			t.Fatalf("CreateCategory: %v", err)
		}
		if _, err := s.UpdateCategory(ctx, home.ID, "WORK"); !errors.Is(err, ErrDuplicate) {
			t.Errorf("duplicate update: err = %v, want ErrDuplicate", err)
		}
	})

	t.Run("todo category link", func(t *testing.T) {
		s := newStore(t)

		work, _ := s.CreateCategory(ctx, "Work")

		item, err := s.Create(ctx, todo.Input{
			Title:      "report",
			Priority:   todo.PriorityMedium,
			CategoryID: &work.ID,
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if item.CategoryID == nil || *item.CategoryID != work.ID {
			t.Fatalf("CategoryID = %v, want %d", item.CategoryID, work.ID)
		}

		got, _ := s.Get(ctx, item.ID)
		if got.CategoryID == nil || *got.CategoryID != work.ID {
			t.Errorf("Get CategoryID = %v, want %d", got.CategoryID, work.ID)
		}

		if _, err := s.Create(ctx, todo.Input{
			Title:      "bad",
			Priority:   todo.PriorityMedium,
			CategoryID: intPtr(9999),
		}); !errors.Is(err, ErrInvalidCategory) {
			t.Errorf("Create with bad category: err = %v, want ErrInvalidCategory", err)
		}

		if _, err := s.Update(ctx, item.ID, todo.Input{
			Title:      "report",
			Priority:   todo.PriorityMedium,
			CategoryID: intPtr(9999),
		}); !errors.Is(err, ErrInvalidCategory) {
			t.Errorf("Update with bad category: err = %v, want ErrInvalidCategory", err)
		}

		cleared, err := s.Update(ctx, item.ID, todo.Input{
			Title:    "report",
			Priority: todo.PriorityMedium,
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if cleared.CategoryID != nil {
			t.Errorf("CategoryID should be cleared, got %v", *cleared.CategoryID)
		}

		if _, err := s.Update(ctx, item.ID, todo.Input{
			Title:      "report",
			Priority:   todo.PriorityMedium,
			CategoryID: &work.ID,
		}); err != nil {
			t.Fatalf("Update: %v", err)
		}

		if err := s.DeleteCategory(ctx, work.ID); err != nil {
			t.Fatalf("DeleteCategory: %v", err)
		}

		after, err := s.Get(ctx, item.ID)
		if err != nil {
			t.Fatalf("todo should survive category deletion: %v", err)
		}
		if after.CategoryID != nil {
			t.Errorf("CategoryID should be nil after category deleted, got %v", *after.CategoryID)
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		s := newStore(t)

		work, _ := s.CreateCategory(ctx, "Work")
		home, _ := s.CreateCategory(ctx, "Home")

		seed := []todo.Input{
			{Title: "A", Priority: todo.PriorityMedium, CategoryID: &work.ID},
			{Title: "B", Priority: todo.PriorityMedium, CategoryID: &home.ID},
			{Title: "C", Priority: todo.PriorityMedium},
			{Title: "D", Priority: todo.PriorityMedium, CategoryID: &work.ID},
		}

		for _, in := range seed {
			if _, err := s.Create(ctx, in); err != nil {
				t.Fatalf("seed Create: %v", err)
			}
		}

		tests := []struct {
			name string
			f    todo.Filter
			want []int
		}{
			{"work", todo.Filter{CategoryID: &work.ID}, []int{1, 4}},
			{"home", todo.Filter{CategoryID: &home.ID}, []int{2}},
			{"uncategorized", todo.Filter{Uncategorized: true}, []int{3}},
			{"unknown category", todo.Filter{CategoryID: intPtr(9999)}, []int{}},
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
