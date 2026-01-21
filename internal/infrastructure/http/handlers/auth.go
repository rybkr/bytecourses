package handlers

import (
	"net/http"
	"strings"

	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/services"
)

type AuthHandler struct {
	Service *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		Service: authService,
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
	var request RegisterRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	user, err := h.Service.Register(r.Context(), request.ToCommand())
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

func (r *LoginRequest) ToCommand() *services.LoginCommand {
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

	result, err := h.Service.Login(r.Context(), request.ToCommand())
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
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
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
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	var request UpdateProfileRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	updated, err := h.Service.UpdateProfile(r.Context(), request.ToCommand(user.ID))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

func (r *RequestPasswordResetRequest) ToCommand() *services.RequestPasswordResetCommand {
	return &services.RequestPasswordResetCommand{
		Email: strings.ToLower(strings.TrimSpace(r.Email)),
	}
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var request RequestPasswordResetRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	// Always return 202 Accepted to avoid email enumeration
	w.WriteHeader(http.StatusAccepted)

	_ = h.Service.RequestPasswordReset(r.Context(), request.ToCommand())
}

type ConfirmPasswordResetRequest struct {
	NewPassword string `json:"new_password"`
}

func (r *ConfirmPasswordResetRequest) ToCommand(token string) *services.ConfirmPasswordResetCommand {
	return &services.ConfirmPasswordResetCommand{
		Token:       token,
		NewPassword: strings.TrimSpace(r.NewPassword),
	}
}

func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var request ConfirmPasswordResetRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	token := r.URL.Query().Get("token")

	if err := h.Service.ConfirmPasswordReset(r.Context(), request.ToCommand(token)); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
