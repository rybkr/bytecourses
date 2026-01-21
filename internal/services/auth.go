package services

import (
	"context"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/auth"
	"bytecourses/internal/infrastructure/email"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/errors"
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
	userRepo persistence.UserRepository,
	resetRepo persistence.PasswordResetRepository,
	sessionStore auth.SessionStore,
	emailSender email.Sender,
	eventBus events.EventBus,
) *AuthService {
	return &AuthService{
		Users:    userRepo,
		Resets:   resetRepo,
		Sessions: sessionStore,
		Email:    emailSender,
		Events:   eventBus,
	}
}

var (
	_ Command = (*RegisterCommand)(nil)
	_ Command = (*LoginCommand)(nil)
	_ Command = (*LogoutCommand)(nil)
	_ Command = (*UpdateProfileCommand)(nil)
	_ Command = (*RequestPasswordResetCommand)(nil)
	_ Command = (*ConfirmPasswordResetCommand)(nil)
)

type RegisterCommand struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *RegisterCommand) Validate(v *validation.Validator) {
	v.Field(c.Name, "name").Required().MinLength(2).MaxLength(80).IsTrimmed()
	v.Field(c.Email, "email").Required().Email()
	v.Field(c.Password, "password").Required().Password()
}

func (s *AuthService) Register(ctx context.Context, cmd *RegisterCommand) (*domain.User, error) {
	if err := validation.Validate(cmd); err != nil {
		return nil, err
	}
	if _, found := s.Users.GetByEmail(ctx, cmd.Email); found {
		return nil, errors.ErrConflict
	}

	passwordHash, err := auth.HashPassword(cmd.Password)
	if err != nil {
		return nil, err
	}

	user := domain.User{
		Name:         cmd.Name,
		Email:        cmd.Email,
		PasswordHash: passwordHash,
		Role:         domain.UserRoleStudent,
	}
	if err := s.Users.Create(ctx, &user); err != nil {
		return nil, err
	}

	event := domain.NewUserRegisteredEvent(user.ID, user.Email, user.Name)
	_ = s.Events.Publish(ctx, event)

	return &user, nil
}

type LoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *LoginCommand) Validate(v *validation.Validator) {
	v.Field(c.Email, "email").Required().Email()
	// Do not validate password rules on login, we cannot block users if policies change
	v.Field(c.Password, "password").Required() // No .Password()
}

func (s *AuthService) Login(ctx context.Context, cmd *LoginCommand) (string, error) {
	if err := validation.Validate(cmd); err != nil {
		return "", err
	}

	user, ok := s.Users.GetByEmail(ctx, cmd.Email)
	if !ok {
		return "", errors.ErrInvalidCredentials
	}
	if err := auth.CheckPassword(user.PasswordHash, cmd.Password); err != nil {
		return "", errors.ErrInvalidCredentials
	}

	sessionID, err := s.Sessions.Create(user.ID)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

type LogoutCommand struct {
	SessionID string `json:"session_id"`
}

func (c *LogoutCommand) Validate(v *validation.Validator) {
	v.Field(c.SessionID, "session_id").Required()
}

func (s *AuthService) Logout(ctx context.Context, cmd *LogoutCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}
	return s.Sessions.Delete(cmd.SessionID)
}

type UpdateProfileCommand struct {
	UserID int64  `json:"-"`
	Name   string `json:"name"`
}

func (i *UpdateProfileCommand) Validate(v *validation.Validator) {
	v.Field(i.UserID, "user_id").Required().EntityID()
	v.Field(i.Name, "name").Required().MinLength(2).MaxLength(80).IsTrimmed()
}

func (s *AuthService) UpdateProfile(ctx context.Context, cmd *UpdateProfileCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	user, ok := s.Users.GetByID(ctx, cmd.UserID)
	if !ok {
		return errors.ErrNotFound
	}

	user.Name = cmd.Name
	if err := s.Users.Update(ctx, user); err != nil {
		return err
	}

	event := domain.NewUserProfileUpdatedEvent(user.ID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type RequestPasswordResetCommand struct {
	Email string `json:"email"`
}

func (i *RequestPasswordResetCommand) Validate(v *validation.Validator) {
	v.Field(i.Email, "email").Required().Email()
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, cmd *RequestPasswordResetCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	user, ok := s.Users.GetByEmail(ctx, cmd.Email)
	if !ok {
		return nil // Return nil to avoid email enumeration
	}

	token, err := auth.GenerateToken()
    if err != nil {
        return err
    }

	tokenHash := auth.HashToken(token)
	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.Resets.CreateResetToken(ctx, user.ID, tokenHash[:], expiresAt); err != nil {
		return err
	}

	event := domain.NewPasswordResetRequestedEvent(user.ID, cmd.Email, token)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type ConfirmPasswordResetCommand struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (i *ConfirmPasswordResetCommand) Validate(v *validation.Validator) {
	v.Field(i.Token, "token").Required().IsTrimmed()
	v.Field(i.NewPassword, "new_password").Required().Password()
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, cmd *ConfirmPasswordResetCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	hash := auth.HashToken(cmd.Token)

	userID, ok := s.Resets.ConsumeResetToken(ctx, hash[:], time.Now())
	if !ok {
		return errors.ErrInvalidToken
	}
    user, ok := s.Users.GetByID(ctx, userID)
    if !ok {
        return errors.ErrNotFound
    }

	passwordHash, err := auth.HashPassword(cmd.NewPassword)
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
