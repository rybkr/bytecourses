package models

import (
	"time"
)

type UserRole string

const (
	RoleStudent    UserRole = "student"
	RoleInstructor UserRole = "instructor"
	RoleAdmin      UserRole = "admin"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Bio          string    `json:"bio"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
