package services

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"strings"
)

type UserService struct {
	users store.UserStore
}

func NewUserService(users store.UserStore) *UserService {
	return &UserService{users: users}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID int64) (*domain.User, error) {
	user, ok := s.users.GetUserByID(ctx, userID)
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

// UpdateProfileRequest represents profile update input
type UpdateProfileRequest struct {
	UserID int64
	Name   string
}

// Normalize trims whitespace
func (r *UpdateProfileRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
}

// UpdateProfile updates a user's profile information
func (s *UserService) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*domain.User, error) {
	req.Normalize()

	if req.Name == "" {
		return nil, ErrInvalidInput
	}

	user, ok := s.users.GetUserByID(ctx, req.UserID)
	if !ok {
		return nil, ErrNotFound
	}

	user.Name = req.Name

	if err := s.users.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
