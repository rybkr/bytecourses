package middleware

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"net/http"
)

func RequireUser(w http.ResponseWriter, r *http.Request, sessions auth.SessionStore, users store.UserStore) (*domain.User, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}

	uid, ok := sessions.GetUserIDByToken(c.Value)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}

	u, ok := users.GetUserByID(r.Context(), uid)
	return u, ok
}

func RequireAdminUser(w http.ResponseWriter, r *http.Request, sessions auth.SessionStore, users store.UserStore) (*domain.User, bool) {
	user, ok := RequireUser(w, r, sessions, users)
	if !ok {
		return nil, false
	}
	if user.Role != domain.UserRoleAdmin {
		return nil, false
	}
	return user, true
}
