package handlers

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/http/middleware"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, val any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(val)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return false
	}
	return true
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func requirePath(w http.ResponseWriter, r *http.Request, path string) bool {
    if r.URL.Path != path {
        http.NotFound(w, r)
        return false
    }
    return true
}

func userFromRequest(r *http.Request) (*domain.User, bool) {
	return middleware.UserFromContext(r.Context())
}

func requireUser(w http.ResponseWriter, r *http.Request) (*domain.User, bool) {
	u, ok := userFromRequest(r)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return nil, false
	}
	return u, true
}
