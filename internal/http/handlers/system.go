package handlers

import (
	"bytecourses/internal/store"
	"context"
	"net/http"
	"time"
)

type DBStatser interface {
	Stats() *store.DBStats
	Ping(ctx context.Context) error
}

type SystemHandlers struct {
	db DBStatser
}

func NewSystemHandlers(db DBStatser) *SystemHandlers {
	return &SystemHandlers{
		db: db,
	}
}

func (h *SystemHandlers) Health(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *SystemHandlers) Diagnostics(w http.ResponseWriter, r *http.Request) {
	out := map[string]any{"storage": "memory"}

	if h.db != nil {
		out["storage"] = "sql"
		out["db"] = h.db.Stats()
	}

	writeJSON(w, http.StatusOK, out)
}
