package store

import (
	"bytecourses/internal/domain"
	"context"
)

// UserStore defines user persistence behavior.
type UserStore interface {
    // CreateUser persists a new user and assigns it a unique ID.
    // The store takes ownership of the provided user pointer.
    // Returns an error if the user cannot be created.
	CreateUser(ctx context.Context, u *domain.User) error

    // GetUserByID returns the user with the given ID.
    // The returned user is a borrowed pointer owned by the store.
    // Returns (*User, true) if the user exists, else (nil, false).
	GetUserByID(ctx context.Context, id int64) (*domain.User, bool)

    // GetUserByEmail returns the user with the given email address.
    // The returned user is a borrowed pointer owned by the store.
    // Returns (*User, true) if the user exists, else (nil, false).
	GetUserByEmail(ctx context.Context, email string) (*domain.User, bool)
}
