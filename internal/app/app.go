package app

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/auth/memsession"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"bytecourses/internal/store/memstore"
	"context"
	"errors"
	"log"
	"time"
)

type App struct {
	UserStore     store.UserStore
	SessionStore  auth.SessionStore
	ProposalStore store.ProposalStore
}

func New(ctx context.Context, cfg Config) (*App, error) {
	a := &App{}

	switch cfg.Storage {
	case StorageMemroy:
		a.UserStore = memstore.NewUserStore()
		a.SessionStore = memsession.New(24 * time.Hour)
		a.ProposalStore = memstore.NewProposalStore()
	case StorageSQL:
		return nil, errors.New("sql backend not implemented yet")
	default:
		return nil, errors.New("unknown storage backend")
	}

	if cfg.SeedUsers {
		if err := ensureTestUsers(ctx, a.UserStore); err != nil {
			log.Fatal(err)
		}
	}

    return a, nil
}

func ensureTestUsers(ctx context.Context, users store.UserStore) error {
	adminEmail := "admin@local.bytecourses.org"
	if _, ok := users.GetUserByEmail(ctx, adminEmail); ok {
		return nil
	}
	hash, err := auth.HashPassword("admin")
	if err != nil {
		return err
	}

	if err := users.InsertUser(ctx, &domain.User{
		Email:        adminEmail,
		PasswordHash: hash,
		Role:         domain.UserRoleAdmin,
		Name:         "Admin User",
	}); err != nil {
		return err
	}

	userEmail := "user@local.bytecourses.org"
	if _, ok := users.GetUserByEmail(ctx, userEmail); ok {
		return nil
	}
	hash, err = auth.HashPassword("user")
	if err != nil {
		return err
	}

	if err := users.InsertUser(ctx, &domain.User{
		Email:        userEmail,
		PasswordHash: hash,
		Role:         domain.UserRoleStudent,
		Name:         "Guest User",
	}); err != nil {
		return err
	}

	return nil
}
