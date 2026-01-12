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
	logger   *AuthLogger
}

func NewAuthService(
	users store.UserStore,
	sessions auth.SessionStore,
	resets store.PasswordResetStore,
	email notify.EmailSender,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		resets:   resets,
		email:    email,
		logger:   NewAuthLogger(logger),
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

func (s *AuthService) Register(ctx context.Context, request *RegisterRequest) (*domain.User, error) {
	request.Normalize()
	if !request.IsValid() {
		return nil, ErrInvalidInput
	}

	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		s.logger.Error("password hashing failed during registration",
			"event", "auth.registration",
			"email", request.Email,
			"error", err,
		)
		return nil, err
	}

	user := &domain.User{
		Email:        request.Email,
		PasswordHash: hash,
		Name:         request.Name,
		Role:         domain.UserRoleStudent,
	}
	if err := s.users.CreateUser(ctx, user); err != nil {
		s.logger.ErrorOp("register", request.Email, err)
		return nil, err
	}

	s.logger.InfoUser("user.registered", user)

	if err := s.email.SendWelcomeEmail(ctx, user.Email, user.Name); err != nil {
		s.logger.Error("failed to send welcome email",
			"event", "auth.registration",
			"user_id", user.ID,
			"email", user.Email,
			"error", err,
		)
	}

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

func (s *AuthService) Login(ctx context.Context, request *LoginRequest) (*LoginResult, error) {
	request.Normalize()
	if !request.IsValid() {
		return nil, ErrInvalidCredentials
	}

	user, ok := s.users.GetUserByEmail(ctx, request.Email)
	if !ok {
		s.logger.Warn("failed login attempt",
			"event", "auth.login.failed",
			"email", request.Email,
			"reason", "invalid_credentials",
		)
		return nil, ErrInvalidCredentials
	}

	if !auth.VerifyPassword(user.PasswordHash, request.Password) {
		s.logger.Warn("failed login attempt",
			"event", "auth.login.failed",
			"email", request.Email,
			"user_id", user.ID,
			"reason", "invalid_password",
		)
		return nil, ErrInvalidCredentials
	}

	token, err := s.sessions.CreateSession(user.ID)
	if err != nil {
		s.logger.Error("session creation failed during login",
			"event", "auth.login",
			"user_id", user.ID,
			"error", err,
		)
		return nil, err
	}

	s.logger.InfoUser("user.login", user)
	return &LoginResult{
		UserID: user.ID,
		Token:  token,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) {
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

func (s *AuthService) UpdateProfile(ctx context.Context, request *UpdateProfileRequest) (*domain.User, error) {
	request.Normalize()
	if !request.IsValid() {
		return nil, ErrInvalidInput
	}

	user, ok := s.users.GetUserByID(ctx, request.UserID)
	if !ok {
		return nil, ErrNotFound
	}

	oldName := user.Name
	user.Name = request.Name

	if err := s.users.UpdateUser(ctx, user); err != nil {
		s.logger.ErrorOp("update_profile", user.Email, err)
		return nil, err
	}

	s.logger.Info("user.profile.updated",
		"user_id", user.ID,
		"name_changed", request.Name != oldName,
	)

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

func (s *AuthService) RequestPasswordReset(ctx context.Context, request *RequestPasswordResetRequest) error {
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
		s.logger.Error("token generation failed for password reset",
			"event", "auth.password_reset",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	tokenHash := sha256.Sum256([]byte(resetToken))
	if err := s.resets.CreateResetToken(ctx, user.ID, tokenHash[:], time.Now().Add(30*time.Minute)); err != nil {
		s.logger.Error("failed to store reset token",
			"event", "auth.password_reset",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	resetURL := request.BaseURL + "/reset-password"
	if err := s.email.SendPasswordResetPrompt(ctx, user.Email, resetURL, resetToken); err != nil {
		s.logger.Error("failed to send password reset email",
			"event", "auth.password_reset",
			"user_id", user.ID,
			"email", user.Email,
			"error", err,
		)
		return err
	}

	s.logger.InfoUser("password_reset.requested", user)

	return nil
}

type ConfirmPasswordResetRequest struct {
    Token       string `json:"-"`
    NewPassword string `json:"password"`
}

func (r *ConfirmPasswordResetRequest) Normalize() {
	r.Token = strings.TrimSpace(r.Token)
	r.NewPassword = strings.TrimSpace(r.NewPassword)
}

func (r *ConfirmPasswordResetRequest) IsValid() bool {
	return r.Token != "" && r.NewPassword != ""
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, request *ConfirmPasswordResetRequest) error {
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
		s.logger.Error("user not found during password reset confirmation",
			"event", "auth.password_reset",
			"user_id", userID,
			"error", "user_not_found",
		)
		return ErrNotFound
	}

	hash, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		s.logger.Error("password hashing failed during reset confirmation",
			"event", "auth.password_reset",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	user.PasswordHash = hash
	if err := s.users.UpdateUser(ctx, user); err != nil {
		s.logger.ErrorOp("confirm_password_reset", user.Email, err)
		return err
	}

	s.logger.InfoUser("password_reset.confirmed", user)

	return nil
}
