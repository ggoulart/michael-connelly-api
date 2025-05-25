package health

import (
	"encoding/json"
	"net/http"
)

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"api": true})
}
