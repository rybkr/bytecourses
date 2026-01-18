package bootstrap

import (
	"bytecourses/internal/application/auth"
	"bytecourses/internal/application/content"
	"bytecourses/internal/application/course"
	"bytecourses/internal/application/module"
	"bytecourses/internal/application/proposal"
	"bytecourses/internal/domain"
	infraauth "bytecourses/internal/infrastructure/auth"
	"bytecourses/internal/infrastructure/email"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/infrastructure/persistence/memory"
	"bytecourses/internal/infrastructure/persistence/postgres"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/validation"
	"context"
	"errors"
	"log/slog"
	"os"
	"time"
)

type Container struct {
	EventBus     events.EventBus
	Validator    *validation.Validator
	SessionStore infraauth.SessionStore
	EmailSender  email.Sender
	DB           persistence.DB

	UserRepo          persistence.UserRepository
	ProposalRepo      persistence.ProposalRepository
	CourseRepo        persistence.CourseRepository
	ModuleRepo        persistence.ModuleRepository
	ContentRepo       persistence.ContentRepository
	PasswordResetRepo persistence.PasswordResetRepository

	RegisterHandler             *auth.RegisterHandler
	LoginHandler                *auth.LoginHandler
	LogoutHandler               *auth.LogoutHandler
	UpdateProfileHandler        *auth.UpdateProfileHandler
	RequestPasswordResetHandler *auth.RequestPasswordResetHandler
	ConfirmPasswordResetHandler *auth.ConfirmPasswordResetHandler
	GetCurrentUserHandler       *auth.GetCurrentUserHandler

	CreateProposalHandler   *proposal.CreateHandler
	UpdateProposalHandler   *proposal.UpdateHandler
	SubmitProposalHandler   *proposal.SubmitHandler
	WithdrawProposalHandler *proposal.WithdrawHandler
	ReviewProposalHandler   *proposal.ReviewHandler
	DeleteProposalHandler   *proposal.DeleteHandler
	GetProposalHandler      *proposal.GetByIDHandler
	ListProposalsHandler    *proposal.ListAllHandler
	ListMyProposalsHandler  *proposal.ListMineHandler

	CreateCourseHandler             *course.CreateHandler
	UpdateCourseHandler             *course.UpdateHandler
	PublishCourseHandler            *course.PublishHandler
	CreateCourseFromProposalHandler *course.CreateFromProposalHandler
	GetCourseHandler                *course.GetByIDHandler
	ListCoursesHandler              *course.ListLiveHandler

	CreateModuleHandler   *module.CreateHandler
	UpdateModuleHandler   *module.UpdateHandler
	DeleteModuleHandler   *module.DeleteHandler
	ReorderModulesHandler *module.ReorderHandler
	GetModuleHandler      *module.GetByIDHandler
	ListModulesHandler    *module.ListByCourseHandler

	CreateLectureHandler  *content.CreateLectureHandler
	UpdateLectureHandler  *content.UpdateLectureHandler
	DeleteContentHandler  *content.DeleteHandler
	ReorderContentHandler *content.ReorderHandler
	GetContentHandler     *content.GetByIDHandler
	ListContentHandler    *content.ListByModuleHandler

	onClose func() error
}

func NewContainer(ctx context.Context, cfg Config) (*Container, error) {
	c := &Container{}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	c.EventBus = events.NewInMemoryEventBus(logger)
	c.Validator = validation.New()
	c.SessionStore = infraauth.NewMemorySessionStore(24 * time.Hour)

	if err := c.setupEmailSender(cfg); err != nil {
		return nil, err
	}
	if err := c.setupPersistence(ctx, cfg); err != nil {
		return nil, err
	}

	c.wireAuthHandlers()
	c.wireProposalHandlers()
	c.wireCourseHandlers()
	c.wireModuleHandlers()
	c.wireContentHandlers()

	c.setupEventSubscribers()

	if err := c.seedData(ctx, cfg); err != nil {
		return nil, err
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
		c.ModuleRepo = memory.NewModuleRepository()
		c.ContentRepo = memory.NewContentRepository()
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
		c.ModuleRepo = postgres.NewModuleRepository(db)
		c.ContentRepo = postgres.NewContentRepository(db)
		c.PasswordResetRepo = postgres.NewPasswordResetRepository(db)
		c.onClose = db.Close

	default:
		return errors.New("unknown storage type")
	}

	return nil
}

func (c *Container) wireAuthHandlers() {
	c.RegisterHandler = auth.NewRegisterHandler(c.UserRepo, c.EmailSender, c.EventBus, c.Validator)
	c.LoginHandler = auth.NewLoginHandler(c.UserRepo, c.SessionStore, c.EventBus, c.Validator)
	c.LogoutHandler = auth.NewLogoutHandler(c.SessionStore)
	c.UpdateProfileHandler = auth.NewUpdateProfileHandler(c.UserRepo, c.EventBus, c.Validator)
	c.RequestPasswordResetHandler = auth.NewRequestPasswordResetHandler(c.UserRepo, c.PasswordResetRepo, c.EmailSender, c.EventBus)
	c.ConfirmPasswordResetHandler = auth.NewConfirmPasswordResetHandler(c.UserRepo, c.PasswordResetRepo, c.EventBus, c.Validator)
	c.GetCurrentUserHandler = auth.NewGetCurrentUserHandler(c.UserRepo)
}

func (c *Container) wireProposalHandlers() {
	c.CreateProposalHandler = proposal.NewCreateHandler(c.ProposalRepo, c.EventBus, c.Validator)
	c.UpdateProposalHandler = proposal.NewUpdateHandler(c.ProposalRepo, c.EventBus, c.Validator)
	c.SubmitProposalHandler = proposal.NewSubmitHandler(c.ProposalRepo, c.EventBus)
	c.WithdrawProposalHandler = proposal.NewWithdrawHandler(c.ProposalRepo, c.EventBus)
	c.ReviewProposalHandler = proposal.NewReviewHandler(c.ProposalRepo, c.EventBus, c.Validator)
	c.DeleteProposalHandler = proposal.NewDeleteHandler(c.ProposalRepo, c.EventBus)
	c.GetProposalHandler = proposal.NewGetByIDHandler(c.ProposalRepo)
	c.ListProposalsHandler = proposal.NewListAllHandler(c.ProposalRepo, c.UserRepo)
	c.ListMyProposalsHandler = proposal.NewListMineHandler(c.ProposalRepo)
}

func (c *Container) wireCourseHandlers() {
	c.CreateCourseHandler = course.NewCreateHandler(c.CourseRepo, c.EventBus, c.Validator)
	c.UpdateCourseHandler = course.NewUpdateHandler(c.CourseRepo, c.EventBus, c.Validator)
	c.PublishCourseHandler = course.NewPublishHandler(c.CourseRepo, c.EventBus)
	c.CreateCourseFromProposalHandler = course.NewCreateFromProposalHandler(c.CourseRepo, c.ProposalRepo, c.EventBus)
	c.GetCourseHandler = course.NewGetByIDHandler(c.CourseRepo)
	c.ListCoursesHandler = course.NewListLiveHandler(c.CourseRepo)
}

func (c *Container) wireModuleHandlers() {
	c.CreateModuleHandler = module.NewCreateHandler(c.ModuleRepo, c.CourseRepo, c.EventBus, c.Validator)
	c.UpdateModuleHandler = module.NewUpdateHandler(c.ModuleRepo, c.CourseRepo, c.EventBus, c.Validator)
	c.DeleteModuleHandler = module.NewDeleteHandler(c.ModuleRepo, c.CourseRepo, c.EventBus)
	c.ReorderModulesHandler = module.NewReorderHandler(c.ModuleRepo, c.CourseRepo, c.EventBus)
	c.GetModuleHandler = module.NewGetByIDHandler(c.ModuleRepo)
	c.ListModulesHandler = module.NewListByCourseHandler(c.ModuleRepo)
}

func (c *Container) wireContentHandlers() {
	c.CreateLectureHandler = content.NewCreateLectureHandler(c.ContentRepo, c.ModuleRepo, c.CourseRepo, c.EventBus, c.Validator)
	c.UpdateLectureHandler = content.NewUpdateLectureHandler(c.ContentRepo, c.ModuleRepo, c.CourseRepo, c.EventBus, c.Validator)
	c.DeleteContentHandler = content.NewDeleteHandler(c.ContentRepo, c.ModuleRepo, c.CourseRepo, c.EventBus)
	c.ReorderContentHandler = content.NewReorderHandler(c.ContentRepo, c.ModuleRepo, c.CourseRepo, c.EventBus)
	c.GetContentHandler = content.NewGetByIDHandler(c.ContentRepo)
	c.ListContentHandler = content.NewListByModuleHandler(c.ContentRepo)
}

func (c *Container) setupEventSubscribers() {
	c.EventBus.Subscribe("user.registered", func(ctx context.Context, e domain.DomainEvent) error {
		event := e.(*domain.UserRegisteredEvent)
		return c.EmailSender.SendWelcomeEmail(ctx, event.Email, event.Name)
	})

	e.EventBus.Subscribe("user.password_reset_requested", func(ctx context.Context, e domain.DomainEvent) error {
		event := e.(*domain.PasswordResetRequestedEvent)
		return c.EmailSender.SendPasswordResetEmail(ctx, event.Email)
	})
}

func (c *Container) seedData(ctx context.Context, cfg Config) error {
	if cfg.SeedUsers {
		if err := seedTestUsers(ctx, c.UserRepo); err != nil {
			return err
		}
	}

	if cfg.SeedProposals {
		if err := seedTestProposals(ctx, c.UserRepo, c.ProposalRepo); err != nil {
			return err
		}
	}

	if err := seedAdmin(ctx, c.UserRepo); err != nil {
		return err
	}

	return nil
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

	if _, ok := users.GetUserByEmail(ctx, email); ok {
		return nil
	}

	hash, err := infraauth.HashPassword(password)
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

func seedTestUsers(ctx context.Context, users persistence.UserRepository) error {
	return nil
}

func seedTestProposals(ctx context.Context, users persistence.UserRepository, proposals persistence.ProposalRepository) error {
	return nil
}
