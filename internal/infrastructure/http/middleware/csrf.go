package middleware

import (
	"log/slog"
	"net/http"

	"bytecourses/internal/infrastructure/auth"
)

const (
	csrfCookieName = "csrf-token"
	csrfHeaderName = "X-CSRF-Token"
)

func isHTTPS(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

func CSRFProtection() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(csrfCookieName)
			var token string
			if err != nil || cookie.Value == "" {
				token, err = auth.GenerateCSRFToken()
				if err != nil {
					slog.Error("failed to generate CSRF token", "error", err)
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			} else {
				token = cookie.Value
			}

			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
				Secure:   isHTTPS(r),
				MaxAge:   60 * 60 * 24,
				// Don't set Domain field to prevent subdomain attacks
			})

			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			headerToken := r.Header.Get(csrfHeaderName)
			if headerToken == "" {
				slog.Warn("CSRF token missing from header", "method", r.Method, "path", r.URL.Path)
				http.Error(w, "CSRF token validation failed", http.StatusForbidden)
				return
			}

			if !auth.ValidateCSRFToken(token, headerToken) {
				slog.Warn("CSRF token validation failed", "method", r.Method, "path", r.URL.Path)
				http.Error(w, "CSRF token validation failed", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
