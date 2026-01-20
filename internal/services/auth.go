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

type AuthService struct {
	Users    persistence.UserRepository
	Resets   persistence.PasswordResetRepository
	Sessions auth.SessionStore
	Email    email.Sender
	Events   events.EventBus
}

func NewAuthService(
	users persistence.UserRepository,
	resets persistence.PasswordResetRepository,
	sessions auth.SessionStore,
	emailSender email.Sender,
	eventBus events.EventBus,
) *AuthService {
	return &AuthService{
		Users:    users,
		Resets:   resets,
		Sessions: sessions,
		Email:    emailSender,
		Events:   eventBus,
	}
}

type RegisterCommand struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *RegisterCommand) Validate(v *validation.Validator) {
	v.Field(c.Name, "name").Required().MinLength(2).MaxLength(80)
	v.Field(c.Email, "email").Required().Email()
	v.Field(c.Password, "password").Required().Password()
}

func (s *AuthService) Register(ctx context.Context, command *RegisterCommand) (*domain.User, error) {
	if err := validation.Validate(command); err != nil {
		return nil, err
	}

	if _, ok := s.Users.GetByEmail(ctx, command.Email); ok {
		return nil, errors.ErrConflict
	}

	hash, err := auth.HashPassword(command.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        command.Email,
		PasswordHash: hash,
		Name:         command.Name,
		Role:         domain.UserRoleStudent,
	}
	if err := s.Users.Create(ctx, user); err != nil {
		return nil, err
	}

	event := domain.NewUserRegisteredEvent(user.ID, user.Email, user.Name)
	_ = s.Events.Publish(ctx, event)

	return user, nil
}

type LoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *LoginCommand) Validate(v *validation.Validator) {
	v.Field(c.Email, "email").Required().Email()
	v.Field(c.Password, "password").Required() // Do not validate password rules on login
}

type LoginResult struct {
	User      *domain.User
	SessionID string
}

func (s *AuthService) Login(ctx context.Context, command *LoginCommand) (*LoginResult, error) {
	if err := validation.Validate(command); err != nil {
		return nil, err
	}

	user, ok := s.Users.GetByEmail(ctx, command.Email)
	if !ok {
		return nil, errors.ErrInvalidCredentials
	}

	if err := auth.CheckPassword(user.PasswordHash, command.Password); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	sessionID, err := s.Sessions.Create(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:      user,
		SessionID: sessionID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.Sessions.Delete(sessionID)
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

	user, ok := s.Users.GetByID(ctx, input.UserID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	user.Name = strings.TrimSpace(input.Name)

	if err := s.Users.Update(ctx, user); err != nil {
		return nil, err
	}

	event := domain.NewUserProfileUpdatedEvent(user.ID)
	_ = s.Events.Publish(ctx, event)

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

	user, ok := s.Users.GetByEmail(ctx, email)
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
	if err := s.Resets.CreateResetToken(ctx, user.ID, hash[:], expiresAt); err != nil {
		return err
	}

	if err := s.Email.SendPasswordResetEmail(ctx, email); err != nil {
		return err
	}

	event := domain.NewPasswordResetRequestedEvent(user.ID, email)
	_ = s.Events.Publish(ctx, event)

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

	userID, ok := s.Resets.ConsumeResetToken(ctx, hash[:], time.Now())
	if !ok {
		return errors.ErrInvalidToken
	}

	user, ok := s.Users.GetByID(ctx, userID)
	if !ok {
		return errors.ErrNotFound
	}

	passwordHash, err := auth.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = passwordHash
	if err := s.Users.Update(ctx, user); err != nil {
		return err
	}

	event := domain.NewPasswordResetCompletedEvent(user.ID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

// GetCurrentUser retrieves the current user by ID.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*domain.User, error) {
	user, ok := s.Users.GetByID(ctx, userID)
	if !ok {
		return nil, errors.ErrNotFound
	}
	return user, nil
}
