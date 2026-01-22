package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	if cfg.SeedUsers != "" {
		if err := c.seedUsers(ctx, cfg.SeedUsers); err != nil {
			return nil, fmt.Errorf("seeding users: %w", err)
		}
	}
	if cfg.SeedProposals != "" {
		if err := c.seedProposals(ctx, cfg.SeedProposals); err != nil {
			return nil, fmt.Errorf("seeding proposals: %w", err)
		}
	}

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

type seedUser struct {
	ID       int64           `json:"id"`
	Email    string          `json:"email"`
	Name     string          `json:"name"`
	Password string          `json:"password"`
	Role     domain.UserRole `json:"role"`
}

func (c *Container) seedUsers(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var users []seedUser
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	for _, u := range users {
		if _, ok := c.UserRepo.GetByEmail(ctx, u.Email); ok {
			continue
		}

		password := u.Password
		if password == "" {
			password = "password"
		}

		hash, err := infraauth.HashPassword(password)
		if err != nil {
			return fmt.Errorf("hashing password for %s: %w", u.Email, err)
		}

		user := &domain.User{
			ID:           u.ID,
			Email:        u.Email,
			Name:         u.Name,
			PasswordHash: hash,
			Role:         u.Role,
		}

		if err := c.UserRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("creating user %s: %w", u.Email, err)
		}
	}

	return nil
}

type seedProposal struct {
	ID                   int64                 `json:"id"`
	Title                string                `json:"title"`
	Summary              string                `json:"summary"`
	Qualifications       string                `json:"qualifications"`
	TargetAudience       string                `json:"target_audience"`
	LearningObjectives   string                `json:"learning_objectives"`
	Outline              string                `json:"outline"`
	AssumedPrerequisites string                `json:"assumed_prerequisites"`
	AuthorID             int64                 `json:"author_id"`
	ReviewNotes          string                `json:"review_notes"`
	ReviewerID           *int64                `json:"reviewer_id"`
	Status               domain.ProposalStatus `json:"status"`
}

func (c *Container) seedProposals(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var proposals []seedProposal
	if err := json.Unmarshal(data, &proposals); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	for _, p := range proposals {
		if _, ok := c.ProposalRepo.GetByID(ctx, p.ID); ok {
			continue
		}

		status := p.Status
		if status == "" {
			status = domain.ProposalStatusDraft
		}

		proposal := &domain.Proposal{
			ID:                   p.ID,
			Title:                p.Title,
			Summary:              p.Summary,
			Qualifications:       p.Qualifications,
			TargetAudience:       p.TargetAudience,
			LearningObjectives:   p.LearningObjectives,
			Outline:              p.Outline,
			AssumedPrerequisites: p.AssumedPrerequisites,
			AuthorID:             p.AuthorID,
			ReviewNotes:          p.ReviewNotes,
			ReviewerID:           p.ReviewerID,
			Status:               status,
		}

		if err := c.ProposalRepo.Create(ctx, proposal); err != nil {
			return fmt.Errorf("creating proposal %q: %w", p.Title, err)
		}
	}

	return nil
}
