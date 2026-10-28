package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"todo-app/internal/store"
)

const maxCategoryNameLen = 50

type categoryRequest struct {
	Name string `json:"name"`
}

func (r categoryRequest) validName() (string, error) {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return "", errors.New("Name is required")
	}

	if utf8.RuneCountInString(name) > maxCategoryNameLen {
		return "", errors.New("Name must be 50 characters or fewer")
	}

	return name, nil
}

func parseCategoryID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Category ID must be a positive integer",
		)
		return 0, false
	}

	return id, true
}

func (h *Handler) getCategoriesHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	items, err := h.store.ListCategories(r.Context())
	if err != nil {
		log.Printf("List categories failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not retrieve categories",
		)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) createCategoryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req categoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	name, err := req.validName()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.store.CreateCategory(r.Context(), name)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "A category with that name already exists")
			return
		}

		log.Printf("Create category failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not create category",
		)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateCategoryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := parseCategoryID(w, r)
	if !ok {
		return
	}

	var req categoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	name, err := req.validName()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.store.UpdateCategory(r.Context(), id, name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Category not found")
			return
		}

		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "A category with that name already exists")
			return
		}

		log.Printf("Update category failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not update category",
		)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) deleteCategoryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := parseCategoryID(w, r)
	if !ok {
		return
	}

	if err := h.store.DeleteCategory(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Category not found")
			return
		}

		log.Printf("Delete category failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"Could not delete category",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
