package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/store"
)

type AuthHandlers struct {
	users    store.UserStore
	sessions auth.SessionStore
}

func NewAuthHandlers(users store.UserStore, sessions auth.SessionStore) *AuthHandlers {
	return &AuthHandlers{
		users:    users,
		sessions: sessions,
	}
}
