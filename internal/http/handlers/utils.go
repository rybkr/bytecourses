package handlers

import (
	"net/http"
)

type UtilHandlers struct{}

func NewUtilHandlers() *UtilHandlers {
	return &UtilHandlers{}
}

func (h *UtilHandlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
