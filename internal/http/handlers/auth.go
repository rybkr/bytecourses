package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/notify"
	"bytecourses/internal/store"
	"crypto/sha256"
	"log"
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

func baseURL(r *http.Request) string {
	scheme := "http"
	if isHTTPS(r) {
		scheme = "https"
	}
	return scheme + "://" + r.Host
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

	log.Printf("RequestPasswordReset: received request for email=%s", request.Email)

	if request.Email == "" {
		log.Printf("RequestPasswordReset: email is empty, returning")
		return
	}
	u, ok := h.users.GetUserByEmail(r.Context(), request.Email)
	if !ok {
		log.Printf("RequestPasswordReset: user not found for email=%s", request.Email)
		return
	}

	log.Printf("RequestPasswordReset: user found, userID=%d", u.ID)

	resetToken, err := auth.GenerateToken(32)
	if err != nil {
		log.Printf("RequestPasswordReset: token generation failed: %v", err)
		return
	}
	log.Printf("RequestPasswordReset: token generated successfully")

	tokenHash := sha256.Sum256([]byte(resetToken))
	if err := h.resets.CreateResetToken(r.Context(), u.ID, tokenHash[:], time.Now().Add(30*time.Minute)); err != nil {
		log.Printf("RequestPasswordReset: failed to create reset token for userID=%d: %v", u.ID, err)
		return
	}
	log.Printf("RequestPasswordReset: reset token stored for userID=%d", u.ID)

	resetURL := baseURL(r) + "/reset-password?token=" + resetToken
	log.Printf("RequestPasswordReset: attempting to send email to=%s, url=%s", u.Email, resetURL)
	if err := h.email.Send(r.Context(), u.Email, "Reset your password", "Click here "+resetURL, ""); err != nil {
		log.Printf("RequestPasswordReset: email send failed for userID=%d, email=%s: %v", u.ID, u.Email, err)
		return
	}
	log.Printf("RequestPasswordReset: email sent successfully to=%s", u.Email)
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

	log.Printf("ConfirmPasswordReset: received confirmation request, token present=%v, password present=%v", request.Token != "", request.NewPassword != "")

	if request.Token == "" || request.NewPassword == "" {
		log.Printf("ConfirmPasswordReset: missing required fields, token present=%v, password present=%v", request.Token != "", request.NewPassword != "")
		http.Error(w, "token and new_password required", http.StatusBadRequest)
		return
	}

	tokenHash := sha256.Sum256([]byte(request.Token))
	userID, ok := h.resets.ConsumeResetToken(r.Context(), tokenHash[:], time.Now())
	if !ok {
		log.Printf("ConfirmPasswordReset: invalid or expired token")
		http.Error(w, "invalid or expired token", http.StatusBadRequest)
		return
	}
	log.Printf("ConfirmPasswordReset: token validated successfully, userID=%d", userID)

	u, ok := h.users.GetUserByID(r.Context(), userID)
	if !ok {
		log.Printf("ConfirmPasswordReset: user not found for userID=%d", userID)
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	log.Printf("ConfirmPasswordReset: user found, userID=%d, email=%s", u.ID, u.Email)

	hash, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		log.Printf("ConfirmPasswordReset: password hashing failed for userID=%d: %v", u.ID, err)
		http.Error(w, "password rejected", http.StatusBadRequest)
		return
	}
	log.Printf("ConfirmPasswordReset: password hashed successfully for userID=%d", u.ID)

	u.PasswordHash = hash
	if err := h.users.UpdateUser(r.Context(), u); err != nil {
		log.Printf("ConfirmPasswordReset: failed to update password for userID=%d: %v", u.ID, err)
		http.Error(w, "failed to update password", http.StatusInternalServerError)
		return
	}
	log.Printf("ConfirmPasswordReset: password updated successfully for userID=%d, email=%s", u.ID, u.Email)

	w.WriteHeader(http.StatusNoContent)
}
