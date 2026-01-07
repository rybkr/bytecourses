package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/http/middleware"
	"bytecourses/internal/store"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthHandler struct {
	users    store.UserStore
	sessions auth.SessionStore
}

func NewAuthHandler(users store.UserStore, sessions auth.SessionStore) *AuthHandler {
	return &AuthHandler{
		users:    users,
		sessions: sessions,
	}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(request.Email))
	password := strings.TrimSpace(request.Password)
	name := strings.TrimSpace(request.Name)
	if email == "" || password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u := domain.User{
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		Role:         domain.UserRoleStudent,
	}

	if err := h.users.CreateUser(r.Context(), &u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(request.Email))
	password := strings.TrimSpace(request.Password)
	u, ok := h.users.GetUserByEmail(r.Context(), email)
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	ok = auth.VerifyPassword(u.PasswordHash, password)
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.sessions.CreateSession(u.ID)
	if err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if c, err := r.Cookie("session"); err == nil {
		h.sessions.DeleteSessionByToken(c.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
        http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(u)
}
