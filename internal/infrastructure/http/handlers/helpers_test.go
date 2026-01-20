package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/pkg/errors"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{
		"message": "hello",
	}

	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Fatalf("writeJSON: expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("writeJSON: expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("writeJSON: failed to decode response: %v", err)
	}
	if result["message"] != "hello" {
		t.Fatalf("writeJSON: expected message 'hello', got %s", result["message"])
	}
}

func TestDecodeJSON(t *testing.T) {
	body := `{"name":"test"}`
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	var dst struct {
		Name string `json:"name"`
	}
	ok := decodeJSON(w, req, &dst)

	if !ok {
		t.Fatalf("decodeJSON: should succeed for valid JSON")
	}
	if dst.Name != "test" {
		t.Fatalf("decodeJSON: expected name 'test', got %s", dst.Name)
	}
	if w.Code >= 400 {
		t.Fatalf("decodeJSON: should not return error status, got %d", w.Code)
	}
}

func TestDecodeJSONInvalid(t *testing.T) {
	body := `{invalid json}`
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	var dst struct{}
	ok := decodeJSON(w, req, &dst)

	if ok {
		t.Fatalf("decodeJSON: should return false for invalid JSON")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("decodeJSON: expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if w.Body.String() != "invalid json\n" {
		t.Fatalf("decodeJSON: expected error message 'invalid json', got %s", w.Body.String())
	}
}

func TestDecodeJSONEmptyBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte{}))
	w := httptest.NewRecorder()

	var dst struct{}
	ok := decodeJSON(w, req, &dst)

	if ok {
		t.Fatalf("decodeJSON: should return false for empty body")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("decodeJSON: expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRequireUser(t *testing.T) {
	user := &domain.User{
		ID:    1,
		Email: "test@example.com",
	}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)
	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	gotUser, ok := requireUser(w, req)

	if !ok {
		t.Fatalf("requireUser: should succeed when user in context")
	}
	if gotUser.ID != user.ID {
		t.Fatalf("requireUser: expected user ID %d, got %d", user.ID, gotUser.ID)
	}
	if w.Code >= 400 {
		t.Fatalf("requireUser: should not return error status, got %d", w.Code)
	}
}

func TestRequireUserMissing(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	user, ok := requireUser(w, req)

	if ok {
		t.Fatalf("requireUser: should return false when user not in context")
	}
	if user != nil {
		t.Fatalf("requireUser: expected nil user, got %v", user)
	}
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("requireUser: expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	if w.Body.String() != "internal error\n" {
		t.Fatalf("requireUser: expected error message 'internal error', got %s", w.Body.String())
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
		checkJSON      bool
	}{
		{
			name:           "ValidationErrors",
			err:            &errors.ValidationErrors{Errors: []errors.ValidationError{{Field: "email", Message: "required"}}},
			expectedStatus: http.StatusBadRequest,
			checkJSON:      true,
		},
		{
			name:           "ErrNotFound",
			err:            errors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "not found\n",
		},
		{
			name:           "ErrUnauthorized",
			err:            errors.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "unauthorized\n",
		},
		{
			name:           "ErrForbidden",
			err:            errors.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "forbidden\n",
		},
		{
			name:           "ErrConflict",
			err:            errors.ErrConflict,
			expectedStatus: http.StatusConflict,
			expectedBody:   "conflict\n",
		},
		{
			name:           "ErrInvalidInput",
			err:            errors.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid input\n",
		},
		{
			name:           "ErrInvalidCredentials",
			err:            errors.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid credentials\n",
		},
		{
			name:           "ErrInvalidToken",
			err:            errors.ErrInvalidToken,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid or expired token\n",
		},
		{
			name:           "ErrInvalidStatusTransition",
			err:            errors.ErrInvalidStatusTransition,
			expectedStatus: http.StatusConflict,
			expectedBody:   "invalid status transition\n",
		},
		{
			name:           "UnknownError",
			err:            stderrors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal error\n",
		},
		{
			name:           "NilError",
			err:            nil,
			expectedStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handleError(w, tt.err)

			if tt.err == nil {
				if w.Code >= 400 {
					t.Fatalf("handleError: should not write error status for nil error, got %d", w.Code)
				}
				return
			}

			if w.Code != tt.expectedStatus {
				t.Fatalf("handleError: expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkJSON {
				if w.Header().Get("Content-Type") != "application/json" {
					t.Fatalf("handleError: expected Content-Type application/json for validation errors")
				}
				var result map[string]any
				if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
					t.Fatalf("handleError: failed to decode JSON response: %v", err)
				}
				if result["error"] != "validation failed" {
					t.Fatalf("handleError: expected error field 'validation failed', got %v", result["error"])
				}
			} else if tt.expectedBody != "" {
				if w.Body.String() != tt.expectedBody {
					t.Fatalf("handleError: expected body %q, got %q", tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

func TestIsHTTPS(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if isHTTPS(req) {
		t.Fatalf("isHTTPS: should return false when X-Forwarded-Proto not set")
	}

	req.Header.Set("X-Forwarded-Proto", "https")
	if !isHTTPS(req) {
		t.Fatalf("isHTTPS: should return true when X-Forwarded-Proto is https")
	}

	req.Header.Set("X-Forwarded-Proto", "http")
	if isHTTPS(req) {
		t.Fatalf("isHTTPS: should return false when X-Forwarded-Proto is http")
	}
}
