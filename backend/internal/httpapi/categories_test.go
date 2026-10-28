package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"todo-app/internal/todo"
)

func decodeCategory(t *testing.T, rec *httptest.ResponseRecorder) todo.Category {
	t.Helper()

	var c todo.Category
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatalf("could not decode category: %v", err)
	}

	return c
}

func listCategoryNames(t *testing.T, h http.Handler) []string {
	t.Helper()

	rec := doRequest(t, h, "GET", "/api/categories", "")
	requireStatus(t, rec, http.StatusOK)

	var items []todo.Category
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatalf("could not decode categories: %v", err)
	}

	names := make([]string, len(items))
	for i, c := range items {
		names[i] = c.Name
	}

	return names
}

func TestCategoryCRUD(t *testing.T) {
	h := newTestHandler()

	rec := doRequest(t, h, "POST", "/api/categories", `{"name":"  Work  "}`)
	requireStatus(t, rec, http.StatusCreated)

	work := decodeCategory(t, rec)
	if work.ID != 1 || work.Name != "Work" {
		t.Errorf("unexpected category: %+v", work)
	}

	requireStatus(t, doRequest(t, h, "POST", "/api/categories", `{"name":"Home"}`), http.StatusCreated)

	if got := listCategoryNames(t, h); !slices.Equal(got, []string{"Home", "Work"}) {
		t.Errorf("names = %v, want [Home Work]", got)
	}

	rec = doRequest(t, h, "PUT", "/api/categories/1", `{"name":"Office"}`)
	requireStatus(t, rec, http.StatusOK)

	if got := decodeCategory(t, rec); got.Name != "Office" {
		t.Errorf("Name = %q, want Office", got.Name)
	}

	if got := listCategoryNames(t, h); !slices.Equal(got, []string{"Home", "Office"}) {
		t.Errorf("names = %v, want [Home Office]", got)
	}

	requireStatus(t, doRequest(t, h, "DELETE", "/api/categories/1", ""), http.StatusNoContent)
	requireStatus(t, doRequest(t, h, "DELETE", "/api/categories/1", ""), http.StatusNotFound)
	requireStatus(t, doRequest(t, h, "PUT", "/api/categories/1", `{"name":"x"}`), http.StatusNotFound)
}

func TestCategoryValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"invalid JSON", `{bad`, http.StatusBadRequest},
		{"missing name", `{}`, http.StatusBadRequest},
		{"blank name", `{"name":"   "}`, http.StatusBadRequest},
		{"name too long", `{"name":"` + strings.Repeat("a", 51) + `"}`, http.StatusBadRequest},
		{"name at limit", `{"name":"` + strings.Repeat("a", 50) + `"}`, http.StatusCreated},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(t, newTestHandler(), "POST", "/api/categories", tc.body)
			requireStatus(t, rec, tc.want)
		})
	}
}

func TestCategoryDuplicate(t *testing.T) {
	h := newTestHandler()

	requireStatus(t, doRequest(t, h, "POST", "/api/categories", `{"name":"Work"}`), http.StatusCreated)
	requireStatus(t, doRequest(t, h, "POST", "/api/categories", `{"name":"work"}`), http.StatusConflict)
	requireStatus(t, doRequest(t, h, "POST", "/api/categories", `{"name":"Home"}`), http.StatusCreated)
	requireStatus(t, doRequest(t, h, "PUT", "/api/categories/2", `{"name":"WORK"}`), http.StatusConflict)
}

func TestCategoryInvalidID(t *testing.T) {
	h := newTestHandler()

	requireStatus(t, doRequest(t, h, "PUT", "/api/categories/abc", `{"name":"x"}`), http.StatusBadRequest)
	requireStatus(t, doRequest(t, h, "DELETE", "/api/categories/0", ""), http.StatusBadRequest)
}

func TestTodoCategoryLink(t *testing.T) {
	h := newTestHandler()

	rec := doRequest(t, h, "POST", "/api/categories", `{"name":"Work"}`)
	requireStatus(t, rec, http.StatusCreated)
	work := decodeCategory(t, rec)

	item := mustCreate(t, h, fmt.Sprintf(`{"title":"a","category_id":%d}`, work.ID))
	if item.CategoryID == nil || *item.CategoryID != work.ID {
		t.Fatalf("CategoryID = %v, want %d", item.CategoryID, work.ID)
	}

	mustCreate(t, h, `{"title":"b"}`)

	if got := listTitles(t, h, "?category=1"); !slices.Equal(got, []string{"a"}) {
		t.Errorf("category=1 titles = %v, want [a]", got)
	}
	if got := listTitles(t, h, "?category=none"); !slices.Equal(got, []string{"b"}) {
		t.Errorf("category=none titles = %v, want [b]", got)
	}

	rec = doRequest(t, h, "PUT", "/api/todos/1", `{"title":"a"}`)
	requireStatus(t, rec, http.StatusOK)
	if got := decodeTodo(t, rec); got.CategoryID != nil {
		t.Errorf("PUT without category_id should clear it, got %v", *got.CategoryID)
	}

	rec = doRequest(t, h, "PUT", "/api/todos/1", fmt.Sprintf(`{"title":"a","category_id":%d}`, work.ID))
	requireStatus(t, rec, http.StatusOK)

	requireStatus(t, doRequest(t, h, "DELETE", "/api/categories/1", ""), http.StatusNoContent)

	rec = doRequest(t, h, "GET", "/api/todos/1", "")
	requireStatus(t, rec, http.StatusOK)
	if got := decodeTodo(t, rec); got.CategoryID != nil {
		t.Errorf("CategoryID should be nil after category deleted, got %v", *got.CategoryID)
	}
}

func TestTodoCategoryValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"nonexistent category", `{"title":"x","category_id":99}`},
		{"zero category", `{"title":"x","category_id":0}`},
		{"negative category", `{"title":"x","category_id":-1}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(t, newTestHandler(), "POST", "/api/todos", tc.body)
			requireStatus(t, rec, http.StatusBadRequest)
		})
	}

	h := newTestHandler()
	mustCreate(t, h, `{"title":"x"}`)
	requireStatus(t, doRequest(t, h, "PUT", "/api/todos/1", `{"title":"x","category_id":99}`), http.StatusBadRequest)
}

func TestListCategoryFilterInvalid(t *testing.T) {
	for _, q := range []string{"?category=abc", "?category=0", "?category=-1"} {
		t.Run(q, func(t *testing.T) {
			requireStatus(t, doRequest(t, newTestHandler(), "GET", "/api/todos"+q, ""), http.StatusBadRequest)
		})
	}
}
