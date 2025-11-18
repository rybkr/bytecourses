package helpers

import (
	"encoding/json"
	"log"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode json response: %v", err)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func ValidationError(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusBadRequest, map[string]interface{}{
		"error":  "validation_failed",
		"fields": fields,
	})
}

func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, map[string]string{"error": message})
}

func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, map[string]string{"error": message})
}

func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, map[string]string{"error": message})
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

func InternalServerError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, map[string]string{"error": message})
}
