package books

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Manager interface {
	Create(ctx context.Context, book Book) (Book, error)
	GetById(ctx context.Context, bookID string) (Book, error)
}

type Controller struct {
	manager  Manager
	validate *validator.Validate
}

func NewController(manager Manager, validator *validator.Validate) *Controller {
	return &Controller{manager: manager, validate: validator}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validate.Struct(book); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, strings.ReplaceAll(err.Error(), "\n", " ")), http.StatusBadRequest)
		return
	}

	createdBook, err := c.manager.Create(r.Context(), book)
	if err != nil {
		http.Error(w, `{"error": "unexpected error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBook)
}

func (c *Controller) GetById(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookID")
	if bookID == "" {
		http.Error(w, `{"error": "book id is required"}`, http.StatusBadRequest)
		return
	}

	book, err := c.manager.GetById(r.Context(), bookID)
	if err != nil {
		http.Error(w, `{"error": "unexpected error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}
