package middleware

import (
	"bytecourses/internal/domain"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockSessionStore struct {
	sessions map[string]int64
}

func (m *mockSessionStore) Create(userID int64) (string, error) {
	if m.sessions == nil {
		m.sessions = make(map[string]int64)
	}
	sessionID := "session123"
	m.sessions[sessionID] = userID
	return sessionID, nil
}

func (m *mockSessionStore) Get(sessionID string) (int64, bool) {
	if m.sessions == nil {
		return 0, false
	}
	userID, ok := m.sessions[sessionID]
	return userID, ok
}

func (m *mockSessionStore) Delete(sessionID string) error {
	if m.sessions == nil {
		return nil
	}
	delete(m.sessions, sessionID)
	return nil
}

type mockUserRepository struct {
	users map[int64]*domain.User
}

func (m *mockUserRepository) Create(ctx context.Context, u *domain.User) error {
	if m.users == nil {
		m.users = make(map[int64]*domain.User)
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, bool) {
	if m.users == nil {
		return nil, false
	}
	user, ok := m.users[id]
	return user, ok
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, bool) {
	if m.users == nil {
		return nil, false
	}
	for _, user := range m.users {
		if user.Email == email {
			return user, true
		}
	}
	return nil, false
}

func (m *mockUserRepository) Update(ctx context.Context, u *domain.User) error {
	if m.users == nil {
		m.users = make(map[int64]*domain.User)
	}
	m.users[u.ID] = u
	return nil
}

func TestRequireUser_ValidSession(t *testing.T) {
	sessions := &mockSessionStore{sessions: map[string]int64{"session123": 1}}
	users := &mockUserRepository{users: map[int64]*domain.User{
		1: {ID: 1, Email: "test@example.com"},
	}}

	handler := RequireUser(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatalf("RequireUser: user should be in context")
		}
		if user.ID != 1 {
			t.Fatalf("RequireUser: expected user ID 1, got %d", user.ID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session123"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("RequireUser: expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireUser_NoCookie(t *testing.T) {
	sessions := &mockSessionStore{}
	users := &mockUserRepository{}

	handler := RequireUser(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("RequireUser: should not call next handler when no cookie")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RequireUser: expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
	if w.Body.String() != "unauthorized\n" {
		t.Fatalf("RequireUser: expected error message 'unauthorized', got %s", w.Body.String())
	}
}

func TestRequireUser_InvalidSession(t *testing.T) {
	sessions := &mockSessionStore{}
	users := &mockUserRepository{}

	handler := RequireUser(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("RequireUser: should not call next handler when session invalid")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalidsession"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RequireUser: expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireUser_UserNotFound(t *testing.T) {
	sessions := &mockSessionStore{sessions: map[string]int64{"session123": 999}}
	users := &mockUserRepository{}

	handler := RequireUser(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("RequireUser: should not call next handler when user not found")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session123"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RequireUser: expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireUser_ContextInjection(t *testing.T) {
	sessions := &mockSessionStore{sessions: map[string]int64{"session123": 1}}
	users := &mockUserRepository{users: map[int64]*domain.User{
		1: {ID: 1, Email: "test@example.com"},
	}}

	var capturedUser *domain.User
	var capturedSession string

	handler := RequireUser(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatalf("RequireUser: user should be in context")
		}
		capturedUser = u

		s, ok := SessionFromContext(r.Context())
		if !ok {
			t.Fatalf("RequireUser: session should be in context")
		}
		capturedSession = s

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session123"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if capturedUser.ID != 1 {
		t.Fatalf("RequireUser: expected user ID 1, got %d", capturedUser.ID)
	}
	if capturedSession != "session123" {
		t.Fatalf("RequireUser: expected session 'session123', got %s", capturedSession)
	}
}

func TestRequireAdmin_AdminUser(t *testing.T) {
	sessions := &mockSessionStore{sessions: map[string]int64{"session123": 1}}
	users := &mockUserRepository{users: map[int64]*domain.User{
		1: {ID: 1, Email: "admin@example.com", Role: domain.UserRoleAdmin},
	}}

	handler := RequireAdmin(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatalf("RequireAdmin: user should be in context")
		}
		if user.Role != domain.UserRoleAdmin {
			t.Fatalf("RequireAdmin: expected admin role")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session123"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("RequireAdmin: expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireAdmin_NonAdmin(t *testing.T) {
	sessions := &mockSessionStore{sessions: map[string]int64{"session123": 1}}
	users := &mockUserRepository{users: map[int64]*domain.User{
		1: {ID: 1, Email: "user@example.com", Role: domain.UserRoleStudent},
	}}

	handler := RequireAdmin(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("RequireAdmin: should not call next handler for non-admin")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session123"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("RequireAdmin: expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if w.Body.String() != "forbidden\n" {
		t.Fatalf("RequireAdmin: expected error message 'forbidden', got %s", w.Body.String())
	}
}

func TestRequireLogin_ValidSession(t *testing.T) {
	sessions := &mockSessionStore{sessions: map[string]int64{"session123": 1}}
	users := &mockUserRepository{users: map[int64]*domain.User{
		1: {ID: 1, Email: "test@example.com"},
	}}

	handler := RequireLogin(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session123"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("RequireLogin: expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireLogin_NoSession(t *testing.T) {
	sessions := &mockSessionStore{}
	users := &mockUserRepository{}

	handler := RequireLogin(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("RequireLogin: should not call next handler when no session")
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("RequireLogin: expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	location := w.Header().Get("Location")
	expected := "/login?next=%2Fprotected"
	if location != expected {
		t.Fatalf("RequireLogin: expected Location %q, got %q", expected, location)
	}
}

func TestOptionalUser_WithSession(t *testing.T) {
	sessions := &mockSessionStore{sessions: map[string]int64{"session123": 1}}
	users := &mockUserRepository{users: map[int64]*domain.User{
		1: {ID: 1, Email: "test@example.com"},
	}}

	handler := OptionalUser(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatalf("OptionalUser: user should be in context when session exists")
		}
		if user.ID != 1 {
			t.Fatalf("OptionalUser: expected user ID 1, got %d", user.ID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session123"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("OptionalUser: expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestOptionalUser_NoSession(t *testing.T) {
	sessions := &mockSessionStore{}
	users := &mockUserRepository{}

	handler := OptionalUser(sessions, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := UserFromContext(r.Context())
		if ok {
			t.Fatalf("OptionalUser: user should not be in context when no session")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("OptionalUser: expected status %d, got %d", http.StatusOK, w.Code)
	}
}
