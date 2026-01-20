package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"bytecourses/internal/domain"
	infraauth "bytecourses/internal/infrastructure/auth"
	"bytecourses/internal/infrastructure/email"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/infrastructure/persistence/memory"
	"bytecourses/internal/infrastructure/persistence/postgres"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/services"
)

type Container struct {
	EventBus     events.EventBus
	SessionStore infraauth.SessionStore
	EmailSender  email.Sender
	DB           persistence.DB

	UserRepo          persistence.UserRepository
	ProposalRepo      persistence.ProposalRepository
	CourseRepo        persistence.CourseRepository
	PasswordResetRepo persistence.PasswordResetRepository

	AuthService     *services.AuthService
	ProposalService *services.ProposalService
	CourseService   *services.CourseService

	onClose func() error
}

func NewContainer(ctx context.Context, cfg Config) (*Container, error) {
	c := &Container{}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	c.EventBus = events.NewInMemoryEventBus(logger)
	c.SessionStore = infraauth.NewInMemorySessionStore(24 * time.Hour)

	if err := c.setupEmailSender(cfg); err != nil {
		return nil, err
	}
	if err := c.setupPersistence(ctx, cfg); err != nil {
		return nil, err
	}

	c.wireServices()
	c.setupEventSubscribers()

	return c, nil
}

func (c *Container) setupEmailSender(cfg Config) error {
	switch cfg.EmailService {
	case EmailServiceResend:
		apiKey := os.Getenv("RESEND_API_KEY")
		if apiKey == "" {
			return errors.New("RESEND_API_KEY not found")
		}
		fromEmail := os.Getenv("RESEND_FROM_EMAIL")
		if fromEmail == "" {
			return errors.New("RESEND_FROM_EMAIL not found")
		}
		c.EmailSender = email.NewResendSender(apiKey, fromEmail)

	case EmailServiceNone:
		c.EmailSender = email.NewNullSender()

	default:
		return errors.New("unknown email service")
	}

	return nil
}

func (c *Container) setupPersistence(ctx context.Context, cfg Config) error {
	switch cfg.Storage {
	case StorageMemory:
		c.UserRepo = memory.NewUserRepository()
		c.ProposalRepo = memory.NewProposalRepository()
		c.CourseRepo = memory.NewCourseRepository()
		c.PasswordResetRepo = memory.NewPasswordResetRepository()

	case StoragePostgres:
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			return errors.New("DATABASE_URL required for postgres storage")
		}

		db, err := postgres.Open(ctx, dbURL)
		if err != nil {
			return err
		}

		c.DB = db
		c.UserRepo = postgres.NewUserRepository(db)
		c.ProposalRepo = postgres.NewProposalRepository(db)
		c.CourseRepo = postgres.NewCourseRepository(db)
		c.PasswordResetRepo = postgres.NewPasswordResetRepository(db)
		c.onClose = db.Close

	default:
		return errors.New("unknown storage type")
	}

	return nil
}

func (c *Container) wireServices() {
	c.AuthService = services.NewAuthService(
		c.UserRepo,
		c.PasswordResetRepo,
		c.SessionStore,
		c.EmailSender,
		c.EventBus,
	)

	c.ProposalService = services.NewProposalService(
		c.ProposalRepo,
		c.UserRepo,
		c.EventBus,
	)

	c.CourseService = services.NewCourseService(
		c.CourseRepo,
		c.ProposalRepo,
		c.EventBus,
	)
}

func (c *Container) setupEventSubscribers() {
	c.EventBus.Subscribe("user.registered", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.UserRegisteredEvent)
		return c.EmailSender.SendWelcomeEmail(ctx, event.Email, event.Name)
	})

	c.EventBus.Subscribe("user.password_reset_requested", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.PasswordResetRequestedEvent)
		return c.EmailSender.SendPasswordResetEmail(ctx, event.Email)
	})
}

func (c *Container) Close() error {
	if c.onClose != nil {
		return c.onClose()
	}
	return nil
}

func seedAdmin(ctx context.Context, users persistence.UserRepository) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	if _, ok := users.GetByEmail(ctx, email); ok {
		return nil
	}

	hash, err := infraauth.HashPassword(password)
	if err != nil {
		return err
	}

	return users.Create(ctx, &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         domain.UserRoleAdmin,
		Name:         "Admin",
	})
}
