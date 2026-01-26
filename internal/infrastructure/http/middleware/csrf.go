package middleware

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"bytecourses/internal/infrastructure/auth"
)

const (
	csrfCookieName = "csrf-token"
	csrfHeaderName = "X-CSRF-Token"
)

func isHTTPS(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

func CSRFProtection(sessions auth.SessionStore, baseURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, hasSession := SessionFromContext(r.Context())
			cookie, _ := r.Cookie(csrfCookieName)

			var token string

			if hasSession && sessionID != "" {
				storedToken, ok := sessions.GetCSRFToken(sessionID)
				if ok && storedToken != "" {
					token = storedToken
				} else {
					var err error
					token, err = auth.GenerateCSRFToken()
					if err != nil {
						slog.Error("failed to generate CSRF token", "error", err)
						http.Error(w, "internal server error", http.StatusInternalServerError)
						return
					}
					if err := sessions.SetCSRFToken(sessionID, token); err != nil {
						slog.Warn("failed to store CSRF token", "error", err)
					}
				}
			} else {
				if cookie != nil && cookie.Value != "" {
					token = cookie.Value
				} else {
					var err error
					token, err = auth.GenerateCSRFToken()
					if err != nil {
						slog.Error("failed to generate CSRF token", "error", err)
						http.Error(w, "internal server error", http.StatusInternalServerError)
						return
					}
				}
			}

			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				SameSite: http.SameSiteStrictMode,
				Secure:   isHTTPS(r),
				MaxAge:   60 * 60 * 24,
			})

			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			if baseURL != "" {
				origin := r.Header.Get("Origin")
				referer := r.Header.Get("Referer")

				if origin != "" {
					if !strings.HasPrefix(origin, baseURL) {
						slog.Warn("CSRF origin validation failed", "origin", origin, "baseURL", baseURL, "path", r.URL.Path)
					}
				} else if referer != "" {
					refererURL, err := url.Parse(referer)
					if err == nil {
						refererOrigin := refererURL.Scheme + "://" + refererURL.Host
						if !strings.HasPrefix(refererOrigin, baseURL) {
							slog.Warn("CSRF referer validation failed", "referer", refererOrigin, "baseURL", baseURL, "path", r.URL.Path)
						}
					}
				}
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

			if hasSession && sessionID != "" {
				storedToken, ok := sessions.GetCSRFToken(sessionID)
				if !ok || storedToken != token {
					slog.Warn("CSRF token not bound to session", "method", r.Method, "path", r.URL.Path)
					http.Error(w, "CSRF token validation failed", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
