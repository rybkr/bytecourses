package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"net/http"
)

type UtilHandlers struct{}

func NewUtilHandlers() *UtilHandlers {
	return &UtilHandlers{}
}

func (h *UtilHandlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func actorFromRequest(r *http.Request, sessions auth.SessionStore, users store.UserStore) (*domain.User, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return nil, false
	}

	uid, ok := sessions.GetUserIDByToken(c.Value)
	if !ok {
		return nil, false
	}

	u, ok := users.GetUserByID(r.Context(), uid)
	return u, ok
}
