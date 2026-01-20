package handlers

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/auth"
	"bytecourses/internal/infrastructure/email"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/infrastructure/persistence/memory"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/services"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func setupAuthService() *services.AuthService {
	userRepo := memory.NewUserRepository()
	resetRepo := memory.NewPasswordResetRepository()
	sessionStore := auth.NewInMemorySessionStore(24 * time.Hour)
	emailSender := email.NewNullSender()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	eventBus := events.NewInMemoryEventBus(logger)

	return services.NewAuthService(userRepo, resetRepo, sessionStore, emailSender, eventBus)
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("ValidRegistration", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		body := `{"email":"test@example.com","password":"password123","name":"Test User"}`
		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Register: expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var result domain.User
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("Register: failed to decode response: %v", err)
		}
		if result.Email != "test@example.com" {
			t.Fatalf("Register: expected email 'test@example.com', got %s", result.Email)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		body := `{invalid json}`
		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Register: expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		// First registration
		body := `{"email":"existing@example.com","password":"password123","name":"Test User"}`
		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.Register(w, req)

		// Second registration with same email
		req = httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		handler.Register(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("Register: expected status %d for conflict, got %d", http.StatusConflict, w.Code)
		}
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("ValidLogin", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		// Register a user first
		_, err := authService.Register(context.Background(), &services.RegisterInput{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		})
		if err != nil {
			t.Fatalf("Setup: failed to register user: %v", err)
		}

		body := `{"email":"test@example.com","password":"password123"}`
		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Login: expected status %d, got %d", http.StatusOK, w.Code)
		}

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("Login: expected 1 cookie, got %d", len(cookies))
		}
		cookie := cookies[0]
		if cookie.Name != "session" {
			t.Fatalf("Login: expected cookie name 'session', got %s", cookie.Name)
		}
		if !cookie.HttpOnly {
			t.Fatalf("Login: cookie should be HttpOnly")
		}
		if cookie.SameSite != http.SameSiteLaxMode {
			t.Fatalf("Login: expected SameSite Lax, got %v", cookie.SameSite)
		}
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		body := `{"email":"test@example.com","password":"wrongpassword"}`
		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Login: expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		body := `{invalid json}`
		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Login: expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("WithSession", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		ctx := middleware.WithSession(context.Background(), "session123")
		req := httptest.NewRequest("POST", "/logout", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Logout(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("Logout: expected status %d, got %d", http.StatusNoContent, w.Code)
		}

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("Logout: expected 1 cookie, got %d", len(cookies))
		}
		cookie := cookies[0]
		if cookie.Name != "session" {
			t.Fatalf("Logout: expected cookie name 'session', got %s", cookie.Name)
		}
		if cookie.Value != "" {
			t.Fatalf("Logout: expected empty cookie value, got %s", cookie.Value)
		}
		if cookie.MaxAge != -1 {
			t.Fatalf("Logout: expected MaxAge -1, got %d", cookie.MaxAge)
		}
	})

	t.Run("WithoutSession", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		req := httptest.NewRequest("POST", "/logout", nil)
		w := httptest.NewRecorder()

		handler.Logout(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("Logout: expected status %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}

func TestAuthHandler_Me(t *testing.T) {
	t.Run("AuthenticatedUser", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		user := &domain.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		}
		ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)
		req := httptest.NewRequest("GET", "/me", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Me(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Me: expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result domain.User
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("Me: failed to decode response: %v", err)
		}
		if result.ID != user.ID {
			t.Fatalf("Me: expected user ID %d, got %d", user.ID, result.ID)
		}
	})

	t.Run("NoUser", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		req := httptest.NewRequest("GET", "/me", nil)
		w := httptest.NewRecorder()

		handler.Me(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("Me: expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestAuthHandler_UpdateProfile(t *testing.T) {
	t.Run("ValidUpdate", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		// Register a user first
		user, err := authService.Register(context.Background(), &services.RegisterInput{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Old Name",
		})
		if err != nil {
			t.Fatalf("Setup: failed to register user: %v", err)
		}

		ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

		body := `{"name":"New Name"}`
		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader([]byte(body))).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("UpdateProfile: expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var result domain.User
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("UpdateProfile: failed to decode response: %v", err)
		}
		if result.Name != "New Name" {
			t.Fatalf("UpdateProfile: expected name 'New Name', got %s", result.Name)
		}
	})

	t.Run("NoUserInContext", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		body := `{"name":"New Name"}`
		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("UpdateProfile: expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestAuthHandler_RequestPasswordReset(t *testing.T) {
	authService := setupAuthService()
	handler := NewAuthHandler(authService)

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest("POST", "/password-reset/request", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.RequestPasswordReset(w, req)

	// Always returns 202 to prevent email enumeration
	if w.Code != http.StatusAccepted {
		t.Fatalf("RequestPasswordReset: expected status %d, got %d", http.StatusAccepted, w.Code)
	}
}

func TestAuthHandler_ConfirmPasswordReset(t *testing.T) {
	t.Run("InvalidToken", func(t *testing.T) {
		authService := setupAuthService()
		handler := NewAuthHandler(authService)

		body := `{"new_password":"newpassword123"}`
		req := httptest.NewRequest("POST", "/password-reset/confirm?token=invalidtoken", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ConfirmPasswordReset(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ConfirmPasswordReset: expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
