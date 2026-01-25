package domain

import (
	"time"
)

type SystemRole string

const (
	SystemRoleUser  SystemRole = "user"
	SystemRoleAdmin SystemRole = "admin"
)

type User struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	PasswordHash []byte     `json:"-"`
	Role         SystemRole `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (u *User) IsAdmin() bool {
	return u.Role == SystemRoleAdmin
}
