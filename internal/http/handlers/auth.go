package handlers

import (
	"bytecourses/internal/services"
	"net/http"
)

type AuthHandler struct {
	services *services.Services
}

func NewAuthHandler(services *services.Services) *AuthHandler {
	return &AuthHandler{
		services: services,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request services.RegisterRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	user, err := h.services.Auth.Register(r.Context(), &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request services.LoginRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	loginResult, err := h.services.Auth.Login(r.Context(), &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    loginResult.Token,
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

	if c, err := r.Cookie("session"); err == nil && c.Value != "" {
		h.services.Auth.Logout(c.Value)
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

	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPatch) {
		return
	}

	u, ok := requireUser(w, r)
	if !ok {
		return
	}

	var request services.UpdateProfileRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.UserID = u.ID

	user, err := h.services.Auth.UpdateProfile(r.Context(), &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request services.RequestPasswordResetRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.BaseURL = baseURL(r)

	// Always return 202 Accepted to avoid email enumeration
	w.WriteHeader(http.StatusAccepted)

	if err := h.services.Auth.RequestPasswordReset(r.Context(), &request); err != nil {
		// Log error but don't expose to client
		// Service error is in charge of logging the error
		_ = err
	}
}

func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request services.ConfirmPasswordResetRequest
	if !decodeJSON(w, r, &request) {
		return
	}
    request.Token = r.URL.Query().Get("token")

	if err := h.services.Auth.ConfirmPasswordReset(r.Context(), &request); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
