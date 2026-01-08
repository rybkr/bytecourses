package handlers

import (
	"bytecourses/internal/store"
	"context"
	"net/http"
	"time"
)

type SystemHandlers struct {
	users store.UserStore
}

func NewSystemHandlers(users store.UserStore) *SystemHandlers {
	return &SystemHandlers{
		users: users,
	}
}

func (h *SystemHandlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if p, ok := h.users.(interface{ Ping(context.Context) error }); ok {
		if err := p.Ping(ctx); err != nil {
			http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SystemHandlers) Diagnostics(w http.ResponseWriter, r *http.Request) {
	out := map[string]any{"storage": "memory"}

	if sp, ok := h.users.(store.StatsProvider); ok {
		out["storage"] = "sql"
		out["db"] = sp.Stats()
	}

	writeJSON(w, http.StatusOK, out)
}
