package handlers

import (
	"net/http"
	"strings"

	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) ToMessage() *services.RegisterCommand {
	return &services.RegisterCommand{
		Name:     strings.TrimSpace(r.Name),
		Email:    strings.ToLower(strings.TrimSpace(r.Email)),
		Password: strings.TrimSpace(r.Password),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	user, err := h.authService.Register(r.Context(), request.ToMessage())
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) ToMessage() *services.LoginCommand {
	return &services.LoginCommand{
		Email:    strings.ToLower(strings.TrimSpace(r.Email)),
		Password: strings.TrimSpace(r.Password),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	result, err := h.authService.Login(r.Context(), request.ToMessage())
	if err != nil {
		handleError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    result.SessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
	})

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := middleware.SessionFromContext(r.Context())
	if ok && sessionID != "" {
		_ = h.authService.Logout(r.Context(), sessionID)
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
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, user)
}

type updateProfileRequest struct {
	Name string `json:"name"`
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var req updateProfileRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	updated, err := h.authService.UpdateProfile(r.Context(), &services.UpdateProfileInput{
		UserID: user.ID,
		Name:   req.Name,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

type requestPasswordResetRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req requestPasswordResetRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	// Always return 202 Accepted to avoid email enumeration
	w.WriteHeader(http.StatusAccepted)

	_ = h.authService.RequestPasswordReset(r.Context(), &services.RequestPasswordResetInput{
		Email: req.Email,
	})
}

type confirmPasswordResetRequest struct {
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req confirmPasswordResetRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	token := r.URL.Query().Get("token")

	if err := h.authService.ConfirmPasswordReset(r.Context(), &services.ConfirmPasswordResetInput{
		Token:       token,
		NewPassword: req.NewPassword,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
