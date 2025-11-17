package store

import (
	"context"
	"github.com/rybkr/bytecourses/internal/models"
	"golang.org/x/crypto/bcrypt"
	"log"
)

func (s *Store) CreateUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		return nil, err
	}

	var user models.User
	query := `
        INSERT INTO users (email, password_hash, role)
        VALUES ($1, $2, $3)
        RETURNING id, email, role, created_at`

	err = s.db.QueryRow(ctx, query, email, string(hash), role).Scan(
		&user.ID, &user.Email, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		log.Printf("failed to create user: %v", err)
		return nil, err
	}

	log.Printf("user created: id=%d, email=%s, role=%s", user.ID, user.Email, user.Role)
	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`

	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		log.Printf("failed to get user by email: %v", err)
		return nil, err
	}

	return &user, nil
}

func (s *Store) ValidateUser(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Printf("invalid password for user: %s", email)
		return nil, err
	}

	log.Printf("user validated: id=%d, email=%s", user.ID, user.Email)
	return user, nil
}
