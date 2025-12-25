package store

import (
	"bytecourses/internal/domain"
	"context"
)

// Store is the persistence boundary for the application.
type Store interface {
    // CreateUser persists a new user and assigns it a unique ID.
    // The store takes ownership of the provided user pointer.
    // Returns an error if the user cannot be created.
	CreateUser(ctx context.Context, u *domain.User) error

    // GetUserByID returns the user with the given ID.
    // The returned user is a borrowed pointer owned by the store.
	GetUserByID(ctx context.Context, id int64) (*domain.User, bool, error)

    // GetUserByEmail returns the user with the given email address.
    // The returned user is a borrowed pointer owned by the store.
	GetUserByEmail(ctx context.Context, email string) (*domain.User, bool)
}
