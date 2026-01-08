package app

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/auth/memsession"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"bytecourses/internal/store/memstore"
	"bytecourses/internal/store/sqlstore"
	"context"
	"errors"
	"log"
	"os"
	"time"
)

type App struct {
	UserStore     store.UserStore
	SessionStore  auth.SessionStore
	ProposalStore store.ProposalStore
	onClose       func() error
}

func New(ctx context.Context, cfg Config) (*App, error) {
	a := &App{}

	switch cfg.Storage {
	case StorageMemroy:
		a.UserStore = memstore.NewUserStore()
		a.ProposalStore = memstore.NewProposalStore()
		a.SessionStore = memsession.New(24 * time.Hour)

	case StorageSQL:
        dbDsn := os.Getenv("DATABASE_URL")
        if dbDsn == "" {
            log.Fatal("DATABASE_URL not set")
        }

		s, err := sqlstore.Open(ctx, dbDsn)
		if err != nil {
			return nil, err
		}

		a.UserStore = s
		a.ProposalStore = s
		a.SessionStore = memsession.New(24 * time.Hour)
		a.onClose = s.Close

	default:
		return nil, errors.New("unknown storage backend")
	}

	if cfg.SeedUsers {
		if err := ensureTestUsers(ctx, a.UserStore); err != nil {
			log.Fatal(err)
		}
	}
	if err := seedAdmin(ctx, a.UserStore); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Close() error {
	if a.onClose != nil {
		return a.onClose()
	}
	return nil
}

func seedAdmin(ctx context.Context, users store.UserStore) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	if _, ok := users.GetUserByEmail(ctx, email); ok {
		return nil
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	return users.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         domain.UserRoleAdmin,
		Name:         "Admin",
	})
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

	if err := users.CreateUser(ctx, &domain.User{
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

	if err := users.CreateUser(ctx, &domain.User{
		Email:        userEmail,
		PasswordHash: hash,
		Role:         domain.UserRoleStudent,
		Name:         "Guest User",
	}); err != nil {
		return err
	}

	return nil
}
