package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthHandlers struct {
	users    store.UserStore
	sessions auth.SessionStore
}

func NewAuthHandlers(users store.UserStore, sessions auth.SessionStore) *AuthHandlers {
	return &AuthHandlers{
		users:    users,
		sessions: sessions,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
    if request.Password == "" || request.Email == "" {
        http.Error(w, "email and password required", http.StatusBadRequest)
        return
    }

	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u := domain.NewUser(strings.TrimSpace(request.Email), hash)

	if err := h.users.InsertUser(r.Context(), u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	u, ok := h.users.GetUserByEmail(r.Context(), strings.TrimSpace(request.Email))
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	ok = auth.VerifyPassword(u.PasswordHash, request.Password)
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _, err := h.sessions.InsertSession(u.ID)
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

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if c, err := r.Cookie("session"); err == nil {
		h.sessions.DeleteSession(c.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
    c, err := r.Cookie("session")
    if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
    }

    uid, ok := h.sessions.GetUserIDByToken(c.Value)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

    u, ok := h.users.GetUserByID(r.Context(), uid)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(u)
}
