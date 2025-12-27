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

func requireUser(w http.ResponseWriter, r *http.Request, sessions auth.SessionStore, users store.UserStore) (domain.User, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return domain.User{}, false
	}

	uid, ok := sessions.GetUserIDByToken(c.Value)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return domain.User{}, false
	}

	u, ok := users.GetUserByID(r.Context(), uid)
	return u, ok
}

func actorFromRequest(r *http.Request, sessions auth.SessionStore, users store.UserStore) (domain.User, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return domain.User{}, false
	}

	uid, ok := sessions.GetUserIDByToken(c.Value)
	if !ok {
		return domain.User{}, false
	}

	u, ok := users.GetUserByID(r.Context(), uid)
	return u, ok
}
