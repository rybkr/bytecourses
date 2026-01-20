package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/auth"
	"bytecourses/internal/infrastructure/email"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/validation"
)

// AuthService handles all authentication operations.
type AuthService struct {
	users    persistence.UserRepository
	resets   persistence.PasswordResetRepository
	sessions auth.SessionStore
	email    email.Sender
	events   events.EventBus
}

// NewAuthService creates a new AuthService with the given dependencies.
func NewAuthService(
	users persistence.UserRepository,
	resets persistence.PasswordResetRepository,
	sessions auth.SessionStore,
	emailSender email.Sender,
	eventBus events.EventBus,
) *AuthService {
	return &AuthService{
		users:    users,
		resets:   resets,
		sessions: sessions,
		email:    emailSender,
		events:   eventBus,
	}
}

// RegisterInput contains the data needed to register a new user.
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (i *RegisterInput) Validate(v *validation.Validator) {
	v.Field(i.Name, "name").Required().MinLength(2).MaxLength(80)
	v.Field(i.Email, "email").Required().Email().MaxLength(254)
	v.Field(i.Password, "password").Required().MinLength(1)
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, input *RegisterInput) (*domain.User, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))

	if _, ok := s.users.GetByEmail(ctx, email); ok {
		return nil, errors.ErrConflict
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: hash,
		Name:         strings.TrimSpace(input.Name),
		Role:         domain.UserRoleStudent,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	event := domain.NewUserRegisteredEvent(user.ID, user.Email, user.Name)
	_ = s.events.Publish(ctx, event)

	return user, nil
}

// LoginInput contains the data needed to log in a user.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (i *LoginInput) Validate(v *validation.Validator) {
	v.Field(i.Email, "email").Required().Email().MaxLength(254)
	v.Field(i.Password, "password").Required()
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	User      *domain.User
	SessionID string
}

// Login authenticates a user and creates a session.
func (s *AuthService) Login(ctx context.Context, input *LoginInput) (*LoginResult, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, ok := s.users.GetByEmail(ctx, email)
	if !ok {
		return nil, errors.ErrInvalidCredentials
	}

	if err := auth.CheckPassword(user.PasswordHash, input.Password); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	sessionID, err := s.sessions.Create(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:      user,
		SessionID: sessionID,
	}, nil
}

// Logout terminates a user session.
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessions.Delete(sessionID)
}

// UpdateProfileInput contains the data needed to update a user profile.
type UpdateProfileInput struct {
	UserID int64  `json:"-"`
	Name   string `json:"name"`
}

func (i *UpdateProfileInput) Validate(v *validation.Validator) {
	v.Field(i.UserID, "user_id").Required().EntityID()
	v.Field(i.Name, "name").Required().MinLength(2).MaxLength(80)
}

// UpdateProfile updates a user's profile information.
func (s *AuthService) UpdateProfile(ctx context.Context, input *UpdateProfileInput) (*domain.User, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	user, ok := s.users.GetByID(ctx, input.UserID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	user.Name = strings.TrimSpace(input.Name)

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	event := domain.NewUserProfileUpdatedEvent(user.ID)
	_ = s.events.Publish(ctx, event)

	return user, nil
}

// RequestPasswordResetInput contains the data needed to request a password reset.
type RequestPasswordResetInput struct {
	Email string `json:"email"`
}

func (i *RequestPasswordResetInput) Validate(v *validation.Validator) {
	v.Field(i.Email, "email").Required().Email().MaxLength(254)
}

// RequestPasswordReset initiates the password reset flow.
// Always returns nil to prevent email enumeration.
func (s *AuthService) RequestPasswordReset(ctx context.Context, input *RequestPasswordResetInput) error {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, ok := s.users.GetByEmail(ctx, email)
	if !ok {
		return nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(token))

	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.resets.CreateResetToken(ctx, user.ID, hash[:], expiresAt); err != nil {
		return err
	}

	if err := s.email.SendPasswordResetEmail(ctx, email); err != nil {
		return err
	}

	event := domain.NewPasswordResetRequestedEvent(user.ID, email)
	_ = s.events.Publish(ctx, event)

	return nil
}

// ConfirmPasswordResetInput contains the data needed to confirm a password reset.
type ConfirmPasswordResetInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (i *ConfirmPasswordResetInput) Validate(v *validation.Validator) {
	v.Field(i.Token, "token").Required()
	v.Field(i.NewPassword, "new_password").Required().MinLength(1)
}

// ConfirmPasswordReset completes the password reset flow.
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, input *ConfirmPasswordResetInput) error {
	if err := validation.New().Validate(input); err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(input.Token))

	userID, ok := s.resets.ConsumeResetToken(ctx, hash[:], time.Now())
	if !ok {
		return errors.ErrInvalidToken
	}

	user, ok := s.users.GetByID(ctx, userID)
	if !ok {
		return errors.ErrNotFound
	}

	passwordHash, err := auth.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = passwordHash
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}

	event := domain.NewPasswordResetCompletedEvent(user.ID)
	_ = s.events.Publish(ctx, event)

	return nil
}

// GetCurrentUser retrieves the current user by ID.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*domain.User, error) {
	user, ok := s.users.GetByID(ctx, userID)
	if !ok {
		return nil, errors.ErrNotFound
	}
	return user, nil
}
