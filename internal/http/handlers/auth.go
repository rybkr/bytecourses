package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/notify"
	"bytecourses/internal/store"
	"net/http"
	"strings"
	"time"
)

type AuthHandler struct {
	users    store.UserStore
	sessions auth.SessionStore
	resets   store.PasswordResetStore
	email    notify.EmailSender
}

func NewAuthHandler(
	users store.UserStore,
	sessions auth.SessionStore,
	resets store.PasswordResetStore,
	email notify.EmailSender,
) *AuthHandler {
	return &AuthHandler{
		users:    users,
		sessions: sessions,
		resets:   resets,
		email:    email,
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

type updateProfileRequest struct {
	Name string `json:"name"`
}

func (r *updateProfileRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
}

func isHTTPS(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-Proto") == "https"
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
		Secure:   isHTTPS(r),
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

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPatch) {
		return
	}
	u, ok := requireUser(w, r)
	if !ok {
		return
	}
	var request updateProfileRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Normalize()

	if request.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	u.Name = request.Name
	if err := h.users.UpdateUser(r.Context(), u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, u)
}

type requestPasswordResetRequest struct {
	Email string `json:"email"`
}

func (r *requestPasswordResetRequest) Normalize() {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var request requestPasswordResetRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Normalize()
	w.WriteHeader(http.StatusAccepted)

	if request.Email == "" {
		return
	}
	u, ok := h.users.GetUserByEmail(r.Context(), request.Email)
	if !ok {
		return
	}

	resetToken, err := auth.GenerateToken(32)
	if err != nil {
		return
	}
	if err := h.resets.CreateResetToken(r.Context(), u.ID, []byte(resetToken), time.Now().Add(30*time.Minute)); err != nil {
		return
	}

	resetURL := "http://localhost:8080" + "/reset-password?token=" + resetToken
	if err := h.email.Send(r.Context(), u.Email, "Reset your password", "Click here "+resetURL, ""); err != nil {
		return
	}
}

type confirmPasswordResetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (r *confirmPasswordResetRequest) Normalize() {
	r.Token = strings.TrimSpace(r.Token)
	r.NewPassword = strings.TrimSpace(r.NewPassword)
}

func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request confirmPasswordResetRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Normalize()

	if request.Token == "" || request.NewPassword == "" {
		http.Error(w, "token and new_password requestuired", http.StatusBadRequest)
		return
	}

	userID, ok := h.resets.ConsumeResetToken(r.Context(), []byte(request.Token), time.Now())
	if !ok {
		http.Error(w, "reset error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "invalid or expired token", http.StatusBadRequest)
		return
	}

	u, ok := h.users.GetUserByID(r.Context(), userID)
	if !ok {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		http.Error(w, "password rejected", http.StatusBadRequest)
		return
	}

	u.PasswordHash = hash
	if err := h.users.UpdateUser(r.Context(), u); err != nil {
		http.Error(w, "failed to update password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
