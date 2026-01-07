package middleware

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"net/http"
)

func userFromRequest(r *http.Request, sessions auth.SessionStore, users store.UserStore) (*domain.User, bool) {
    c, err := r.Cookie("session")
    if err != nil || c.Value == "" {
        return nil, false
    }

    uid, ok := sessions.GetUserIDByToken(c.Value)
    if !ok {
        return nil, false
    }

    u, ok := users.GetUserByID(r.Context(), uid)
    if !ok {
        return nil, false
    }
    return u, true
}

// RequireUser is middleware that enforces authentication and injects
// the resolved user into the request context.
func RequireUser(sessions auth.SessionStore, users store.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := userFromRequest(r, sessions, users)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), u)))
		})
	}
}

// RequireAdmin is middleware that enforces admin authorization.
// It assumes authentication and injects the admin user into context.
func RequireAdmin(sessions auth.SessionStore, users store.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := userFromRequest(r, sessions, users)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if u.Role != domain.UserRoleAdmin {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), u)))
		})
	}
}
