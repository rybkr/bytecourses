package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/store"
	"encoding/json"
	"net/http"
	"strings"
)

type LoginHandler struct {
	users    store.UserStore
	sessions auth.SessionStore
}

func NewLoginHandler(users store.UserStore, sessions auth.SessionStore) *LoginHandler {
	return &LoginHandler{users: users, sessions: sessions}
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(req.Email)
	u, ok := h.users.GetUserByEmail(r.Context(), email)
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	ok, _ = auth.VerifyPassword(u.PasswordHash, req.Password)
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _, err := h.sessions.Create(u.ID)
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
