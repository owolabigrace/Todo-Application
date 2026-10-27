package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"todo-app/internal/store"
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

	mux.HandleFunc("GET /", h.homeHandler)
	mux.HandleFunc("GET /api/todos", h.getTodosHandler)
	mux.HandleFunc("POST /api/todos", h.createTodoHandler)
	mux.HandleFunc("GET /api/todos/{id}", h.getTodoHandler)
	mux.HandleFunc("PUT /api/todos/{id}", h.updateTodoHandler)
	mux.HandleFunc("DELETE /api/todos/{id}", h.deleteTodoHandler)

	return mux
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
	items, err := h.store.List(r.Context())
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
	var input struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Invalid JSON request body",
		)
		return
	}

	input.Title = strings.TrimSpace(input.Title)

	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if len([]rune(input.Title)) > 200 {
		writeError(
			w,
			http.StatusBadRequest,
			"Title must be 200 characters or fewer",
		)
		return
	}

	item, err := h.store.Create(r.Context(), input.Title)
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

	var input struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Invalid JSON request body",
		)
		return
	}

	input.Title = strings.TrimSpace(input.Title)

	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if len([]rune(input.Title)) > 200 {
		writeError(
			w,
			http.StatusBadRequest,
			"Title must be 200 characters or fewer",
		)
		return
	}

	item, err := h.store.Update(
		r.Context(),
		id,
		input.Title,
		input.Completed,
	)
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
