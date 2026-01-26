package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	apperrors "bytecourses/internal/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
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
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return false
	}

	return true
}

func mapDatabaseError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case "23505":
		return apperrors.ErrConflict
	case "23503":
		return apperrors.ErrInvalidInput
	case "23502":
		return apperrors.ErrInvalidInput
	default:
		return err
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	err = mapDatabaseError(err)

	if validationErrs, ok := err.(*apperrors.ValidationErrors); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":  "Validation failed",
			"errors": validationErrs.Errors,
		})
		return
	}

	statusCode := apperrors.GetStatusCode(err)
	message := apperrors.GetUserMessage(err)

	if statusCode == http.StatusInternalServerError {
		slog.Error("unexpected error",
			"error", err,
			"path", r.URL.Path,
			"method", r.Method,
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func handlePageError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	err = mapDatabaseError(err)

	if validationErrs, ok := err.(*apperrors.ValidationErrors); ok {
		if len(validationErrs.Errors) > 0 {
			http.Error(w, validationErrs.Errors[0].Message, http.StatusBadRequest)
			return
		}
		http.Error(w, "Validation failed", http.StatusBadRequest)
		return
	}

	statusCode := apperrors.GetStatusCode(err)
	message := apperrors.GetUserMessage(err)

	if statusCode == http.StatusInternalServerError {
		slog.Error("unexpected error",
			"error", err,
			"path", r.URL.Path,
			"method", r.Method,
		)
	}

	http.Error(w, message, statusCode)
}

func isHTTPS(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-Proto") == "https"
}
