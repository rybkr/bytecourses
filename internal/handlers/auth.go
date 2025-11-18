package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rybkr/bytecourses/internal/helpers"
	"github.com/rybkr/bytecourses/internal/models"
	"github.com/rybkr/bytecourses/internal/store"
	"github.com/rybkr/bytecourses/internal/validation"
)

type AuthHandler struct {
	store *store.Store
}

func NewAuthHandler(store *store.Store) *AuthHandler {
	return &AuthHandler{store: store}
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode signup request: %v", err)
		helpers.BadRequest(w, "invalid request body")
		return
	}

	v := validation.New()
	v.Required(req.Email, "email")
	v.Email(req.Email, "email")
	v.Required(req.Password, "password")
	v.MinLength(req.Password, 8, "password")

	if !v.Valid() {
		log.Printf("signup validation failed: %v", v.Errors)
		helpers.ValidationError(w, v.Errors)
		return
	}

	role := models.RoleStudent
	if req.Role == "instructor" {
		role = models.RoleInstructor
	}

	user, err := h.store.CreateUser(r.Context(), req.Email, req.Password, role)
	if err != nil {
		log.Printf("failed to create user in handler: %v", err)
		helpers.InternalServerError(w, "failed to create user")
		return
	}

	token, err := generateToken(user)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Created(w, AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode login request: %v", err)
		helpers.BadRequest(w, "invalid request body")
		return
	}

	v := validation.New()
	v.Required(req.Email, "email")
	v.Required(req.Password, "password")

	if !v.Valid() {
		helpers.ValidationError(w, v.Errors)
		return
	}

	user, err := h.store.ValidateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		helpers.Unauthorized(w, "invalid credentials")
		return
	}

	token, err := generateToken(user)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Success(w, AuthResponse{Token: token, User: user})
}

func generateToken(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
