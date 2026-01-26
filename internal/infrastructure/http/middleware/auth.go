package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/auth"
	"bytecourses/internal/infrastructure/persistence"
)

func userFromRequest(r *http.Request, sessions auth.SessionStore, users persistence.UserRepository) (*domain.User, string, bool) {
	c, err := r.Cookie("session")
	if err != nil || c.Value == "" {
		return nil, "", false
	}

	userID, ok := sessions.Get(c.Value)
	if !ok {
		return nil, "", false
	}
	user, ok := users.GetByID(r.Context(), userID)
	if !ok {
		return nil, "", false
	}

	return user, c.Value, true
}

func RequireUser(sessions auth.SessionStore, users persistence.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, sessionID, ok := userFromRequest(r, sessions, users)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := WithUser(r.Context(), user)
			ctx = WithSession(ctx, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(sessions auth.SessionStore, users persistence.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, sessionID, ok := userFromRequest(r, sessions, users)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin() {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			ctx := WithUser(r.Context(), user)
			ctx = WithSession(ctx, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireLogin(sessions auth.SessionStore, users persistence.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, sessionID, ok := userFromRequest(r, sessions, users)
			if !ok {
				nextVal := nextRedirectTarget(r)
				http.Redirect(w, r, "/login?next="+url.QueryEscape(nextVal), http.StatusSeeOther)
				return
			}

			ctx := WithUser(r.Context(), user)
			ctx = WithSession(ctx, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func nextRedirectTarget(r *http.Request) string {
	path := r.URL.Path
	if path == "" {
		path = "/"
	}
	raw := path
	if r.URL.RawQuery != "" {
		raw = path + "?" + r.URL.RawQuery
	}
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || strings.HasPrefix(strings.ToLower(raw), "javascript:") {
		return path
	}
	switch {
	case raw == "/login", raw == "/register", strings.HasPrefix(raw, "/login?"), strings.HasPrefix(raw, "/register?"):
		return "/"
	}
	return raw
}

func OptionalUser(sessions auth.SessionStore, users persistence.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, sessionID, ok := userFromRequest(r, sessions, users)
			if ok {
				ctx := WithUser(r.Context(), user)
				ctx = WithSession(ctx, sessionID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
