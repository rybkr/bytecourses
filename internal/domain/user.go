package domain

import (
	"time"
)

type UserRole string

const (
	UserRoleStudent    UserRole = "student"
	UserRoleInstructor UserRole = "instructor"
	UserRoleAdmin      UserRole = "admin"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash []byte    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}
