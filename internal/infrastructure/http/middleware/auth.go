package middleware

import (
	"net/http"
	"net/url"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/auth"
	"bytecourses/internal/infrastructure/persistence"
)

func userFromRequest(r *http.Request, sessions auth.SessionStore, users persistence.UserRepository) (*domain.User, string, bool) {
	c, err := r.Cookie("session")
	if err != nil || c.Value == "" {
		return nil, "", false
	}

	uid, ok := sessions.Get(c.Value)
	if !ok {
		return nil, "", false
	}

	u, ok := users.GetByID(r.Context(), uid)
	if !ok {
		return nil, "", false
	}
	return u, c.Value, true
}

func RequireUser(sessions auth.SessionStore, users persistence.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, sessionID, ok := userFromRequest(r, sessions, users)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := WithUser(r.Context(), u)
			ctx = WithSession(ctx, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(sessions auth.SessionStore, users persistence.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, sessionID, ok := userFromRequest(r, sessions, users)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if u.Role != domain.UserRoleAdmin {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			ctx := WithUser(r.Context(), u)
			ctx = WithSession(ctx, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireLogin(sessions auth.SessionStore, users persistence.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, sessionID, ok := userFromRequest(r, sessions, users)
			if !ok {
				http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.Path), http.StatusSeeOther)
				return
			}
			ctx := WithUser(r.Context(), u)
			ctx = WithSession(ctx, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalUser(sessions auth.SessionStore, users persistence.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, sessionID, ok := userFromRequest(r, sessions, users)
			if ok {
				ctx := WithUser(r.Context(), u)
				ctx = WithSession(ctx, sessionID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
