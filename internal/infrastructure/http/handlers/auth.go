package handlers

import (
	"net/http"
	"strings"

	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/services"
)

type AuthHandler struct {
	Service *services.AuthService
	BaseURL string
}

func NewAuthHandler(authService *services.AuthService, baseURL string) *AuthHandler {
	return &AuthHandler{
		Service: authService,
		BaseURL: baseURL,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) ToCommand() *services.RegisterCommand {
	return &services.RegisterCommand{
		Name:     strings.TrimSpace(r.Name),
		Email:    strings.ToLower(strings.TrimSpace(r.Email)),
		Password: strings.TrimSpace(r.Password),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	user, err := h.Service.Register(r.Context(), req.ToCommand())
	if err != nil {
		handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) ToCommand() *services.LoginCommand {
	return &services.LoginCommand{
		Email:    strings.ToLower(strings.TrimSpace(r.Email)),
		Password: strings.TrimSpace(r.Password),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	sessionID, err := h.Service.Login(r.Context(), req.ToCommand())
	if err != nil {
		handleError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   isHTTPS(r),
		MaxAge:   60 * 60 * 24, // 1 day
	})

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := middleware.SessionFromContext(r.Context())
	if ok && sessionID != "" {
		_ = h.Service.Logout(r.Context(), &services.LogoutCommand{
			SessionID: strings.TrimSpace(sessionID),
		})
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   isHTTPS(r),
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

func (r *UpdateProfileRequest) ToCommand(userID int64) *services.UpdateProfileCommand {
	return &services.UpdateProfileCommand{
		Name:   strings.TrimSpace(r.Name),
		UserID: userID,
	}
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	var req UpdateProfileRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	err := h.Service.UpdateProfile(r.Context(), req.ToCommand(user.ID))
	if err != nil {
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	err := h.Service.DeleteUser(r.Context(), &services.DeleteUserCommand{
		UserID: user.ID,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

func (r *RequestPasswordResetRequest) ToCommand(baseURL string) *services.RequestPasswordResetCommand {
	return &services.RequestPasswordResetCommand{
		Email:   strings.ToLower(strings.TrimSpace(r.Email)),
		BaseURL: strings.TrimSpace(baseURL),
	}
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req RequestPasswordResetRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	// Always return 202 Accepted to avoid email enumeration
	w.WriteHeader(http.StatusAccepted)

	_ = h.Service.RequestPasswordReset(r.Context(), req.ToCommand(h.BaseURL))
}

type ConfirmPasswordResetRequest struct {
	NewPassword string `json:"new_password"`
}

func (r *ConfirmPasswordResetRequest) ToCommand(token string) *services.ConfirmPasswordResetCommand {
	return &services.ConfirmPasswordResetCommand{
		Token:       strings.TrimSpace(token),
		NewPassword: strings.TrimSpace(r.NewPassword),
	}
}

func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req ConfirmPasswordResetRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	token := r.URL.Query().Get("token")
	if err := h.Service.ConfirmPasswordReset(r.Context(), req.ToCommand(token)); err != nil {
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
