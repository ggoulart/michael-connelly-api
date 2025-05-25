package characters

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
	Create(ctx context.Context, character Character) (Character, error)
	GetById(ctx context.Context, characterID string) (Character, error)
}

type Controller struct {
	manager  Manager
	validate *validator.Validate
}

func NewController(manager Manager, validator *validator.Validate) *Controller {
	return &Controller{manager: manager, validate: validator}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var character Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validate.Struct(character); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, strings.ReplaceAll(err.Error(), "\n", " ")), http.StatusBadRequest)
		return
	}

	createdCharacter, err := c.manager.Create(r.Context(), character)
	if err != nil {
		http.Error(w, `{"error": "unexpected error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdCharacter)
}

func (c *Controller) GetById(w http.ResponseWriter, r *http.Request) {
	characterID := chi.URLParam(r, "characterID")
	if characterID == "" {
		http.Error(w, `{"error": "character id is required"}`, http.StatusBadRequest)
		return
	}

	character, err := c.manager.GetById(r.Context(), characterID)
	if err != nil {
		http.Error(w, `{"error": "unexpected error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(character)
}
