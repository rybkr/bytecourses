package handlers

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/http/middleware"
	"bytecourses/internal/services"
	"encoding/json"
	"errors"
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

func proposalFromRequest(r *http.Request) (*domain.Proposal, bool) {
	return middleware.ProposalFromContext(r.Context())
}

func requireProposal(w http.ResponseWriter, r *http.Request) (*domain.Proposal, bool) {
	p, ok := proposalFromRequest(r)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return nil, false
	}
	return p, true
}

func courseFromRequest(r *http.Request) (*domain.Course, bool) {
	return middleware.CourseFromContext(r.Context())
}

func requireCourse(w http.ResponseWriter, r *http.Request) (*domain.Course, bool) {
	c, ok := courseFromRequest(r)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return nil, false
	}
	return c, true
}

func moduleFromRequest(r *http.Request) (*domain.Module, bool) {
	return middleware.ModuleFromContext(r.Context())
}

func requireModule(w http.ResponseWriter, r *http.Request) (*domain.Module, bool) {
	m, ok := moduleFromRequest(r)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return nil, false
	}
	return m, true
}

func handleServiceError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, services.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, services.ErrUnauthorized):
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	case errors.Is(err, services.ErrForbidden):
		http.Error(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, services.ErrConflict):
		http.Error(w, "conflict", http.StatusConflict)
	case errors.Is(err, services.ErrInvalidInput):
		http.Error(w, "invalid input", http.StatusBadRequest)
	case errors.Is(err, services.ErrInvalidCredentials):
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	case errors.Is(err, services.ErrInvalidToken):
		http.Error(w, "invalid or expired token", http.StatusBadRequest)
	case errors.Is(err, services.ErrInvalidStatusTransition):
		http.Error(w, "invalid status transition", http.StatusBadRequest)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
