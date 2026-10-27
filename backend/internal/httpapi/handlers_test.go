package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"todo-app/internal/store"
	"todo-app/internal/todo"
)

func newTestHandler() http.Handler {
	return NewHandler(store.NewMemoryStore()).Routes()
}

func doRequest(
	t *testing.T,
	h http.Handler,
	method, path, body string,
) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func decodeTodo(t *testing.T, rec *httptest.ResponseRecorder) todo.Todo {
	t.Helper()

	var item todo.Todo
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
		t.Fatalf("could not decode todo: %v", err)
	}

	return item
}

func requireStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, want, rec.Body.String())
	}
}

func TestCreateTodo(t *testing.T) {
	h := newTestHandler()

	rec := doRequest(t, h, "POST", "/api/todos", `{"title":"  Buy milk  "}`)
	requireStatus(t, rec, http.StatusCreated)

	item := decodeTodo(t, rec)
	if item.ID != 1 {
		t.Errorf("ID = %d, want 1", item.ID)
	}
	if item.Title != "Buy milk" {
		t.Errorf("Title = %q, want trimmed %q", item.Title, "Buy milk")
	}
	if item.Completed {
		t.Error("new todo should not be completed")
	}
}

func TestCreateTodoValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"invalid JSON", `{not json`, http.StatusBadRequest},
		{"missing title", `{}`, http.StatusBadRequest},
		{"blank title", `{"title":"   "}`, http.StatusBadRequest},
		{"title too long", `{"title":"` + strings.Repeat("a", 201) + `"}`, http.StatusBadRequest},
		{"title at limit", `{"title":"` + strings.Repeat("a", 200) + `"}`, http.StatusCreated},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(t, newTestHandler(), "POST", "/api/todos", tc.body)
			requireStatus(t, rec, tc.want)
		})
	}
}

func TestListTodos(t *testing.T) {
	h := newTestHandler()

	rec := doRequest(t, h, "GET", "/api/todos", "")
	requireStatus(t, rec, http.StatusOK)

	var items []todo.Todo
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatalf("could not decode list: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("empty store returned %d todos", len(items))
	}

	doRequest(t, h, "POST", "/api/todos", `{"title":"one"}`)
	doRequest(t, h, "POST", "/api/todos", `{"title":"two"}`)

	rec = doRequest(t, h, "GET", "/api/todos", "")
	requireStatus(t, rec, http.StatusOK)

	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatalf("could not decode list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d todos, want 2", len(items))
	}
	if items[0].Title != "one" || items[1].Title != "two" {
		t.Errorf("unexpected order: %+v", items)
	}
}

func TestGetTodo(t *testing.T) {
	h := newTestHandler()
	doRequest(t, h, "POST", "/api/todos", `{"title":"one"}`)

	rec := doRequest(t, h, "GET", "/api/todos/1", "")
	requireStatus(t, rec, http.StatusOK)
	if got := decodeTodo(t, rec); got.Title != "one" {
		t.Errorf("Title = %q, want %q", got.Title, "one")
	}

	requireStatus(t, doRequest(t, h, "GET", "/api/todos/99", ""), http.StatusNotFound)
	requireStatus(t, doRequest(t, h, "GET", "/api/todos/abc", ""), http.StatusBadRequest)
	requireStatus(t, doRequest(t, h, "GET", "/api/todos/0", ""), http.StatusBadRequest)
	requireStatus(t, doRequest(t, h, "GET", "/api/todos/-1", ""), http.StatusBadRequest)
}

func TestUpdateTodo(t *testing.T) {
	h := newTestHandler()
	doRequest(t, h, "POST", "/api/todos", `{"title":"one"}`)

	rec := doRequest(t, h, "PUT", "/api/todos/1", `{"title":"renamed","completed":true}`)
	requireStatus(t, rec, http.StatusOK)

	got := decodeTodo(t, rec)
	if got.Title != "renamed" || !got.Completed {
		t.Errorf("unexpected update result: %+v", got)
	}

	// The change must persist.
	rec = doRequest(t, h, "GET", "/api/todos/1", "")
	requireStatus(t, rec, http.StatusOK)
	if got := decodeTodo(t, rec); got.Title != "renamed" || !got.Completed {
		t.Errorf("update did not persist: %+v", got)
	}
}

func TestUpdateTodoErrors(t *testing.T) {
	h := newTestHandler()
	doRequest(t, h, "POST", "/api/todos", `{"title":"one"}`)

	tests := []struct {
		name string
		path string
		body string
		want int
	}{
		{"not found", "/api/todos/99", `{"title":"x","completed":false}`, http.StatusNotFound},
		{"invalid id", "/api/todos/abc", `{"title":"x"}`, http.StatusBadRequest},
		{"invalid JSON", "/api/todos/1", `{bad`, http.StatusBadRequest},
		{"blank title", "/api/todos/1", `{"title":" ","completed":true}`, http.StatusBadRequest},
		{"title too long", "/api/todos/1", `{"title":"` + strings.Repeat("a", 201) + `"}`, http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requireStatus(t, doRequest(t, h, "PUT", tc.path, tc.body), tc.want)
		})
	}
}

func TestDeleteTodo(t *testing.T) {
	h := newTestHandler()
	doRequest(t, h, "POST", "/api/todos", `{"title":"one"}`)

	requireStatus(t, doRequest(t, h, "DELETE", "/api/todos/1", ""), http.StatusNoContent)
	requireStatus(t, doRequest(t, h, "GET", "/api/todos/1", ""), http.StatusNotFound)
	requireStatus(t, doRequest(t, h, "DELETE", "/api/todos/1", ""), http.StatusNotFound)
	requireStatus(t, doRequest(t, h, "DELETE", "/api/todos/abc", ""), http.StatusBadRequest)
}
