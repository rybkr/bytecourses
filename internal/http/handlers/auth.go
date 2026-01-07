package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
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

func (r *registerRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Password = strings.TrimSpace(r.Password)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var request registerRequest
    if !decodeJSON(w, r, &request) {
		return
	}
    request.Normalize()

	if request.Email == "" || request.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u := domain.User{
		Email:        request.Email,
		PasswordHash: hash,
		Name:         request.Name,
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

func (r *loginRequest) Normalize() {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Password = strings.TrimSpace(r.Password)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var request loginRequest
	if !decodeJSON(w, r, &request) {
		return
	}
    request.Normalize()

	if request.Email == "" || request.Password == "" {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	u, ok := h.users.GetUserByEmail(r.Context(), request.Email)
	if !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	ok = auth.VerifyPassword(u.PasswordHash, request.Password)
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
	if !requireMethod(w, r, http.MethodPost) {
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
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	u, ok := requireUser(w, r)
	if !ok {
		return
	}
    writeJSON(w, http.StatusOK, u)
}
