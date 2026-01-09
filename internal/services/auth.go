package services

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/notify"
	"bytecourses/internal/store"
	"context"
	"crypto/sha256"
	"log"
	"strings"
	"time"
)

type AuthService struct {
	users    store.UserStore
	sessions auth.SessionStore
	resets   store.PasswordResetStore
	email    notify.EmailSender
}

func NewAuthService(
	users store.UserStore,
	sessions auth.SessionStore,
	resets store.PasswordResetStore,
	email notify.EmailSender,
) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		resets:   resets,
		email:    email,
	}
}

// RegisterRequest represents user registration input
type RegisterRequest struct {
	Name     string
	Email    string
	Password string
}

// Normalize trims whitespace and normalizes email
func (r *RegisterRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Password = strings.TrimSpace(r.Password)
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	req.Normalize()

	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         domain.UserRoleStudent,
	}

	if err := s.users.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// LoginRequest represents login input
type LoginRequest struct {
	Email    string
	Password string
}

// Normalize trims whitespace and normalizes email
func (r *LoginRequest) Normalize() {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Password = strings.TrimSpace(r.Password)
}

// LoginResult contains the session token for successful login
type LoginResult struct {
	UserID int64
	Token  string
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	req.Normalize()

	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, ok := s.users.GetUserByEmail(ctx, req.Email)
	if !ok {
		return nil, ErrInvalidCredentials
	}

	if !auth.VerifyPassword(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.sessions.CreateSession(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID: user.ID,
		Token:  token,
	}, nil
}

// Logout invalidates a session
func (s *AuthService) Logout(token string) {
	s.sessions.DeleteSessionByToken(token)
}

// RequestPasswordResetRequest represents password reset request input
type RequestPasswordResetRequest struct {
	Email   string
	BaseURL string // For constructing reset link
}

// Normalize trims whitespace and normalizes email
func (r *RequestPasswordResetRequest) Normalize() {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
}

// RequestPasswordReset initiates a password reset flow
func (s *AuthService) RequestPasswordReset(ctx context.Context, req RequestPasswordResetRequest) error {
	req.Normalize()

	log.Printf("RequestPasswordReset: received request for email=%s", req.Email)

	if req.Email == "" {
		log.Printf("RequestPasswordReset: email is empty, returning")
		return nil // Return nil for security (don't reveal if email exists)
	}

	user, ok := s.users.GetUserByEmail(ctx, req.Email)
	if !ok {
		log.Printf("RequestPasswordReset: user not found for email=%s", req.Email)
		return nil // Return nil for security
	}

	log.Printf("RequestPasswordReset: user found, userID=%d", user.ID)

	resetToken, err := auth.GenerateToken(32)
	if err != nil {
		log.Printf("RequestPasswordReset: token generation failed: %v", err)
		return err
	}
	log.Printf("RequestPasswordReset: token generated successfully")

	tokenHash := sha256.Sum256([]byte(resetToken))
	if err := s.resets.CreateResetToken(ctx, user.ID, tokenHash[:], time.Now().Add(30*time.Minute)); err != nil {
		log.Printf("RequestPasswordReset: failed to create reset token for userID=%d: %v", user.ID, err)
		return err
	}
	log.Printf("RequestPasswordReset: reset token stored for userID=%d", user.ID)

	resetURL := req.BaseURL + "/reset-password?token=" + resetToken
	log.Printf("RequestPasswordReset: attempting to send email to=%s, url=%s", user.Email, resetURL)
	if err := s.email.Send(ctx, user.Email, "Reset your password", "Click here "+resetURL, ""); err != nil {
		log.Printf("RequestPasswordReset: email send failed for userID=%d, email=%s: %v", user.ID, user.Email, err)
		return err
	}
	log.Printf("RequestPasswordReset: email sent successfully to=%s", user.Email)

	return nil
}

// ConfirmPasswordResetRequest represents password reset confirmation input
type ConfirmPasswordResetRequest struct {
	Token       string
	NewPassword string
}

// Normalize trims whitespace
func (r *ConfirmPasswordResetRequest) Normalize() {
	r.Token = strings.TrimSpace(r.Token)
	r.NewPassword = strings.TrimSpace(r.NewPassword)
}

// ConfirmPasswordReset completes the password reset flow
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, req ConfirmPasswordResetRequest) error {
	req.Normalize()

	log.Printf("ConfirmPasswordReset: received confirmation request, token present=%v, password present=%v", req.Token != "", req.NewPassword != "")

	if req.Token == "" || req.NewPassword == "" {
		log.Printf("ConfirmPasswordReset: missing required fields, token present=%v, password present=%v", req.Token != "", req.NewPassword != "")
		return ErrInvalidInput
	}

	tokenHash := sha256.Sum256([]byte(req.Token))
	userID, ok := s.resets.ConsumeResetToken(ctx, tokenHash[:], time.Now())
	if !ok {
		log.Printf("ConfirmPasswordReset: invalid or expired token")
		return ErrInvalidToken
	}
	log.Printf("ConfirmPasswordReset: token validated successfully, userID=%d", userID)

	user, ok := s.users.GetUserByID(ctx, userID)
	if !ok {
		log.Printf("ConfirmPasswordReset: user not found for userID=%d", userID)
		return ErrNotFound
	}
	log.Printf("ConfirmPasswordReset: user found, userID=%d, email=%s", user.ID, user.Email)

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("ConfirmPasswordReset: password hashing failed for userID=%d: %v", user.ID, err)
		return err
	}
	log.Printf("ConfirmPasswordReset: password hashed successfully for userID=%d", user.ID)

	user.PasswordHash = hash
	if err := s.users.UpdateUser(ctx, user); err != nil {
		log.Printf("ConfirmPasswordReset: failed to update password for userID=%d: %v", user.ID, err)
		return err
	}
	log.Printf("ConfirmPasswordReset: password updated successfully for userID=%d, email=%s", user.ID, user.Email)

	return nil
}
