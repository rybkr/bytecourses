package services

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/notify"
	"bytecourses/internal/store"
	"context"
	"crypto/sha256"
	"log/slog"
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

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		r.Name = "Guest User"
	}
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Password = strings.TrimSpace(r.Password)
}

func (r *RegisterRequest) IsValid() bool {
	return r.Name != "" && r.Email != "" && r.Password != ""
}

func (s *AuthService) Register(ctx context.Context, request RegisterRequest) (*domain.User, error) {
	start := time.Now()
	slog.Info("auth.register.attempt")

	request.Normalize()
	if !request.IsValid() {
		slog.Warn("auth.register.invalid_input")
		return nil, ErrInvalidInput
	}

	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		slog.Error("auth.register.hash_error", "err", err)
		return nil, err
	}

	user := &domain.User{
		Email:        request.Email,
		PasswordHash: hash,
		Name:         request.Name,
		Role:         domain.UserRoleStudent,
	}
	if err := s.users.CreateUser(ctx, user); err != nil {
		slog.Error("auth.register.store_error", "err", err)
		return nil, err
	}

	slog.Info("auth.register.success",
		"user_id", user.ID,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return user, nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Normalize() {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Password = strings.TrimSpace(r.Password)
}

func (r *LoginRequest) IsValid() bool {
	return r.Email != "" && r.Password != ""
}

type LoginResult struct {
	UserID int64
	Token  string
}

func (s *AuthService) Login(ctx context.Context, request LoginRequest) (*LoginResult, error) {
	request.Normalize()
	if !request.IsValid() {
		return nil, ErrInvalidCredentials
	}

	user, ok := s.users.GetUserByEmail(ctx, request.Email)
	if !ok {
		return nil, ErrInvalidCredentials
	}
	if !auth.VerifyPassword(user.PasswordHash, request.Password) {
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

func (s *AuthService) Logout(token string) {
	s.sessions.DeleteSessionByToken(token)
}

type UpdateProfileRequest struct {
	UserID int64  `json:"-"`
	Name   string `json:"name"`
}

func (r *UpdateProfileRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
}

func (r *UpdateProfileRequest) IsValid() bool {
	return r.Name != ""
}

func (s *AuthService) UpdateProfile(ctx context.Context, request UpdateProfileRequest) (*domain.User, error) {
	request.Normalize()
	if !request.IsValid() {
		return nil, ErrInvalidInput
	}

	user, ok := s.users.GetUserByID(ctx, request.UserID)
	if !ok {
		return nil, ErrNotFound
	}

	user.Name = request.Name

	if err := s.users.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

type RequestPasswordResetRequest struct {
	Email   string `json:"email"`
	BaseURL string `json:"-"`
}

func (r *RequestPasswordResetRequest) Normalize() {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.BaseURL = strings.TrimSpace(r.BaseURL)
}

func (r *RequestPasswordResetRequest) IsValid() bool {
	return r.Email != ""
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, request RequestPasswordResetRequest) error {
	request.Normalize()
	if !request.IsValid() {
		return ErrInvalidInput
	}

	user, ok := s.users.GetUserByEmail(ctx, request.Email)
	if !ok {
		return ErrNotFound
	}

	resetToken, err := auth.GenerateToken(32)
	if err != nil {
		return err
	}
	tokenHash := sha256.Sum256([]byte(resetToken))
	if err := s.resets.CreateResetToken(ctx, user.ID, tokenHash[:], time.Now().Add(30*time.Minute)); err != nil {
		return err
	}

	resetURL := request.BaseURL + "/reset-password?token=" + resetToken
	if err := s.email.Send(ctx, user.Email, "Reset your password", "Click here "+resetURL, ""); err != nil {
		return err
	}

	return nil
}

type ConfirmPasswordResetRequest struct {
	Token       string
	NewPassword string
}

func (r *ConfirmPasswordResetRequest) Normalize() {
	r.Token = strings.TrimSpace(r.Token)
	r.NewPassword = strings.TrimSpace(r.NewPassword)
}

func (r *ConfirmPasswordResetRequest) IsValid() bool {
	return r.Token != "" && r.NewPassword != ""
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, request ConfirmPasswordResetRequest) error {
	request.Normalize()
	if !request.IsValid() {
		return ErrInvalidInput
	}

	tokenHash := sha256.Sum256([]byte(request.Token))
	userID, ok := s.resets.ConsumeResetToken(ctx, tokenHash[:], time.Now())
	if !ok {
		return ErrInvalidToken
	}

	user, ok := s.users.GetUserByID(ctx, userID)
	if !ok {
		return ErrNotFound
	}

	hash, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	if err := s.users.UpdateUser(ctx, user); err != nil {
		return err
	}

	return nil
}
