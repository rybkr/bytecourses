package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"bytecourses/internal/pkg/errors"
)

func writeJSON(w http.ResponseWriter, status int, val any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(val); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return false
	}

	return true
}

func handleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	if validationErrs, ok := err.(*errors.ValidationErrors); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":  "validation failed",
			"errors": validationErrs.Errors,
		})
		return
	}

	switch err {
	case errors.ErrNotFound:
		http.Error(w, "not found", http.StatusNotFound)
	case errors.ErrUnauthorized:
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	case errors.ErrForbidden:
		http.Error(w, "forbidden", http.StatusForbidden)
	case errors.ErrConflict:
		http.Error(w, "conflict", http.StatusConflict)
	case errors.ErrInvalidInput:
		http.Error(w, "invalid input", http.StatusBadRequest)
	case errors.ErrInvalidCredentials:
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	case errors.ErrInvalidToken:
		http.Error(w, "invalid or expired token", http.StatusBadRequest)
	case errors.ErrInvalidStatusTransition:
		http.Error(w, "invalid status transition", http.StatusConflict)
	case errors.ErrInvalidLogin:
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func isHTTPS(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-Proto") == "https"
}
