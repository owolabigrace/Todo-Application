package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"todo-app/internal/todo"
)

func mustCreate(t *testing.T, h http.Handler, body string) todo.Todo {
	t.Helper()

	rec := doRequest(t, h, "POST", "/api/todos", body)
	requireStatus(t, rec, http.StatusCreated)

	return decodeTodo(t, rec)
}

func listTitles(t *testing.T, h http.Handler, query string) []string {
	t.Helper()

	rec := doRequest(t, h, "GET", "/api/todos"+query, "")
	requireStatus(t, rec, http.StatusOK)

	var items []todo.Todo
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatalf("could not decode list: %v", err)
	}

	titles := make([]string, len(items))
	for i, it := range items {
		titles[i] = it.Title
	}

	return titles
}

func TestCreateTodoWithDetails(t *testing.T) {
	h := newTestHandler()

	item := mustCreate(t, h, `{"title":"Pay rent","notes":"  bank transfer  ","priority":"high","due_date":"2026-11-01"}`)

	if item.Notes != "bank transfer" {
		t.Errorf("Notes = %q, want trimmed", item.Notes)
	}
	if item.Priority != "high" {
		t.Errorf("Priority = %q, want high", item.Priority)
	}
	if item.DueDate == nil || *item.DueDate != "2026-11-01" {
		t.Errorf("DueDate = %v, want 2026-11-01", item.DueDate)
	}
}

func TestCreateTodoDefaults(t *testing.T) {
	item := mustCreate(t, newTestHandler(), `{"title":"x"}`)

	if item.Priority != "medium" {
		t.Errorf("Priority = %q, want medium", item.Priority)
	}
	if item.Notes != "" || item.DueDate != nil {
		t.Errorf("unexpected defaults: %+v", item)
	}
}

func TestCreateTodoDetailValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"invalid priority", `{"title":"x","priority":"urgent"}`, http.StatusBadRequest},
		{"wrong date format", `{"title":"x","due_date":"01/11/2026"}`, http.StatusBadRequest},
		{"impossible date", `{"title":"x","due_date":"2026-02-30"}`, http.StatusBadRequest},
		{"notes too long", `{"title":"x","notes":"` + strings.Repeat("a", 2001) + `"}`, http.StatusBadRequest},
		{"notes at limit", `{"title":"x","notes":"` + strings.Repeat("a", 2000) + `"}`, http.StatusCreated},
		{"empty due date means none", `{"title":"x","due_date":""}`, http.StatusCreated},
		{"null due date means none", `{"title":"x","due_date":null}`, http.StatusCreated},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(t, newTestHandler(), "POST", "/api/todos", tc.body)
			requireStatus(t, rec, tc.want)
		})
	}
}

func TestUpdateTodoDetails(t *testing.T) {
	h := newTestHandler()
	created := mustCreate(t, h, `{"title":"one","due_date":"2026-11-01"}`)
	path := fmt.Sprintf("/api/todos/%d", created.ID)

	rec := doRequest(t, h, "PUT", path, `{"title":"one","notes":"n","completed":true,"priority":"low","due_date":null}`)
	requireStatus(t, rec, http.StatusOK)

	got := decodeTodo(t, rec)
	if got.Notes != "n" || got.Priority != "low" || !got.Completed {
		t.Errorf("unexpected update result: %+v", got)
	}
	if got.DueDate != nil {
		t.Errorf("DueDate should be cleared, got %v", *got.DueDate)
	}

	requireStatus(t, doRequest(t, h, "PUT", path, `{"title":"one","priority":"nope"}`), http.StatusBadRequest)
}

func TestListFilters(t *testing.T) {
	h := newTestHandler()

	mustCreate(t, h, `{"title":"Write report","priority":"high","due_date":"2026-11-01"}`)
	mustCreate(t, h, `{"title":"Buy milk","priority":"low"}`)
	dave := mustCreate(t, h, `{"title":"Call Dave","notes":"about the report"}`)

	rec := doRequest(t, h, "PUT", fmt.Sprintf("/api/todos/%d", dave.ID),
		`{"title":"Call Dave","notes":"about the report","completed":true}`)
	requireStatus(t, rec, http.StatusOK)

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"default", "", []string{"Write report", "Buy milk", "Call Dave"}},
		{"active", "?status=active", []string{"Write report", "Buy milk"}},
		{"completed", "?status=completed", []string{"Call Dave"}},
		{"priority", "?priority=high", []string{"Write report"}},
		{"search", "?q=report", []string{"Write report", "Call Dave"}},
		{"sort title", "?sort=title", []string{"Buy milk", "Call Dave", "Write report"}},
		{"sort due", "?sort=due", []string{"Write report", "Buy milk", "Call Dave"}},
		{"sort priority desc", "?sort=priority&order=desc", []string{"Write report", "Call Dave", "Buy milk"}},
		{"combined", "?status=active&sort=title&order=desc", []string{"Write report", "Buy milk"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := listTitles(t, h, tc.query); !slices.Equal(got, tc.want) {
				t.Errorf("titles = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestListInvalidQuery(t *testing.T) {
	queries := []string{
		"?status=done",
		"?priority=urgent",
		"?sort=random",
		"?order=sideways",
		"?q=" + strings.Repeat("a", 201),
	}

	for _, q := range queries {
		name := q
		if len(name) > 30 {
			name = name[:30]
		}

		t.Run(name, func(t *testing.T) {
			requireStatus(t, doRequest(t, newTestHandler(), "GET", "/api/todos"+q, ""), http.StatusBadRequest)
		})
	}
}

func TestUnknownPathIsNotFound(t *testing.T) {
	h := newTestHandler()

	requireStatus(t, doRequest(t, h, "GET", "/", ""), http.StatusOK)
	requireStatus(t, doRequest(t, h, "GET", "/api/nonsense", ""), http.StatusNotFound)
}
