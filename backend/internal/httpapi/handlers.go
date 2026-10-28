package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"todo-app/internal/store"
	"todo-app/internal/todo"
)

const (
	maxTitleLen  = 200
	maxNotesLen  = 2000
	maxSearchLen = 200
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{
		store: s,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", h.homeHandler)
	mux.HandleFunc("GET /api/todos", h.getTodosHandler)
	mux.HandleFunc("POST /api/todos", h.createTodoHandler)
	mux.HandleFunc("GET /api/todos/{id}", h.getTodoHandler)
	mux.HandleFunc("PUT /api/todos/{id}", h.updateTodoHandler)
	mux.HandleFunc("DELETE /api/todos/{id}", h.deleteTodoHandler)

	return mux
}

// todoRequest is the JSON body for creating and updating a todo.
type todoRequest struct {
	Title     string  `json:"title"`
	Notes     string  `json:"notes"`
	Completed bool    `json:"completed"`
	Priority  string  `json:"priority"`
	DueDate   *string `json:"due_date"`
}

func (r todoRequest) toInput() (todo.Input, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return todo.Input{}, errors.New("Title is required")
	}

	if utf8.RuneCountInString(title) > maxTitleLen {
		return todo.Input{}, errors.New("Title must be 200 characters or fewer")
	}

	notes := strings.TrimSpace(r.Notes)
	if utf8.RuneCountInString(notes) > maxNotesLen {
		return todo.Input{}, errors.New("Notes must be 2000 characters or fewer")
	}

	priority := strings.TrimSpace(r.Priority)
	if priority == "" {
		priority = todo.PriorityMedium
	}

	if !todo.ValidPriority(priority) {
		return todo.Input{}, errors.New("Priority must be low, medium or high")
	}

	var due *string

	if r.DueDate != nil {
		d := strings.TrimSpace(*r.DueDate)
		if d != "" {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return todo.Input{}, errors.New("Due date must be a valid date in YYYY-MM-DD format")
			}

			due = &d
		}
	}

	return todo.Input{
		Title:     title,
		Notes:     notes,
		Completed: r.Completed,
		Priority:  priority,
		DueDate:   due,
	}, nil
}

func parseFilter(q url.Values) (todo.Filter, error) {
	f := todo.Filter{
		Status:   q.Get("status"),
		Priority: q.Get("priority"),
		Search:   strings.TrimSpace(q.Get("q")),
		Sort:     q.Get("sort"),
		Order:    q.Get("order"),
	}

	switch f.Status {
	case "", "all", "active", "completed":
	default:
		return f, errors.New("status must be all, active or completed")
	}

	if f.Priority != "" && !todo.ValidPriority(f.Priority) {
		return f, errors.New("priority must be low, medium or high")
	}

	switch f.Sort {
	case "", "created", "due", "priority", "title":
	default:
		return f, errors.New("sort must be created, due, priority or title")
	}

	switch f.Order {
	case "", "asc", "desc":
	default:
		return f, errors.New("order must be asc or desc")
	}

	if utf8.RuneCountInString(f.Search) > maxSearchLen {
		return f, errors.New("search text must be 200 characters or fewer")
	}

	return f, nil
}

func (h *Handler) homeHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Welcome to my Todo List API!",
	})
}

func (h *Handler) getTodosHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	filter, err := parseFilter(r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.store.List(r.Context(), filter)
	if err != nil {
		log.Printf("List todos failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not retrieve Todos",
		)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) createTodoHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req todoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Invalid JSON request body",
		)
		return
	}

	input, err := req.toInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.store.Create(r.Context(), input)
	if err != nil {
		log.Printf("Create todo failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not create Todo",
		)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) getTodoHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Todo ID must be a positive integer",
		)
		return
	}

	item, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Todo not found")
			return
		}

		log.Printf("Get todo failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not retrieve Todo",
		)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) updateTodoHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Todo ID must be a positive integer",
		)
		return
	}

	var req todoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Invalid JSON request body",
		)
		return
	}

	input, err := req.toInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.store.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Todo not found")
			return
		}

		log.Printf("Update todo failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not update Todo",
		)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) deleteTodoHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Todo ID must be a positive integer",
		)
		return
	}

	if err := h.store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Todo not found")
			return
		}

		log.Printf("Delete todo failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not delete Todo",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(value string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid Todo ID")
	}

	return id, nil
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func NewServer(s store.Store) *http.Server {
	handler := NewHandler(s)

	return &http.Server{
		Addr:              ":8080",
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
