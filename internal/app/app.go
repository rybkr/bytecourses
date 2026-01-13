package app

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/auth/memsession"
	"bytecourses/internal/domain"
	"bytecourses/internal/notify"
	"bytecourses/internal/notify/nullsender"
	"bytecourses/internal/notify/resend"
	"bytecourses/internal/services"
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
	Services           *services.Services
	UserStore          store.UserStore
	ProposalStore      store.ProposalStore
	CourseStore        store.CourseStore
	SessionStore       auth.SessionStore
	PasswordResetStore store.PasswordResetStore
	DB                 store.DB
	EmailSender        notify.EmailSender
	onClose            func() error
}

func New(ctx context.Context, cfg Config) (*App, error) {
	a := &App{}

	switch cfg.Storage {
	case StorageMemroy:
		a.UserStore = memstore.NewUserStore()
		a.ProposalStore = memstore.NewProposalStore()
		a.CourseStore = memstore.NewCourseStore()
		a.SessionStore = memsession.New(24 * time.Hour)
		a.PasswordResetStore = memstore.NewPasswordResetStore()

	case StorageSQL:
		dbDsn := os.Getenv("DATABASE_URL")
		if dbDsn == "" {
			return nil, errors.New("DATABASE_URL is not set")
		}
		if db, err := sqlstore.Open(ctx, dbDsn); err == nil {
			a.UserStore = db
			a.ProposalStore = db
			a.PasswordResetStore = db
			a.SessionStore = memsession.New(24 * time.Hour)
			a.DB = db
			a.onClose = db.Close
		} else {
			return nil, err
		}

	default:
		return nil, errors.New("unrecognized memory configuration")
	}

	switch cfg.EmailService {
	case EmailServiceResend:
		apiKey := os.Getenv("RESEND_API_KEY")
		fromEmail := os.Getenv("RESEND_FROM_EMAIL")
		if apiKey == "" || fromEmail == "" {
			return nil, errors.New("RESEND_API_KEY and RESEND_FROM_EMAIL are not set")
		}
		a.EmailSender = resend.New(apiKey, fromEmail)
	case EmailServiceNone:
		a.EmailSender = nullsender.New()
	default:
		return nil, errors.New("unrecognized email service provider")
	}

	logger := services.NewLogger()
	if logger == nil || logger.Logger == nil {
		log.Fatal("failed to create logger")
	}
	a.Services = services.New(services.Dependencies{
		UserStore:          a.UserStore,
		ProposalStore:      a.ProposalStore,
		CourseStore:        a.CourseStore,
		PasswordResetStore: a.PasswordResetStore,
		SessionStore:       a.SessionStore,
		EmailSender:        a.EmailSender,
		Logger:             logger.Logger,
	})

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
