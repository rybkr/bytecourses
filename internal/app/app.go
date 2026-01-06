package app

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/auth/memsession"
	"bytecourses/internal/store"
	"bytecourses/internal/store/memstore"
	"context"
	"errors"
    "time"
)

type App struct {
	UserStore     store.UserStore
	SessionStore  auth.SessionStore
	ProposalStore store.ProposalStore
}

func New(ctx context.Context, cfg Config) (*App, error) {
	switch cfg.Storage {
	case StorageMemroy:
		return &App{
			UserStore:     memstore.NewUserStore(),
			SessionStore:  memsession.New(24 * time.Hour),
			ProposalStore: memstore.NewProposalStore(),
		}, nil
	case StorageSQL:
		return nil, errors.New("sql backend not implemented yet")
	default:
		return nil, errors.New("unknown storage backend")
	}
}
