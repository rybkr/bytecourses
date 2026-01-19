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

	c.EventBus.Subscribe("user.password_reset_requested", func(ctx context.Context, e domain.DomainEvent) error {
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

const (
	adminEmail string = "admin@local.bytecourses.org"
	userEmail  string = "user@local.bytecourses.org"
)

func seedTestUsers(ctx context.Context, users persistence.UserRepository) error {
	if _, ok := users.GetUserByEmail(ctx, adminEmail); !ok {
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
	}

	if _, ok := users.GetUserByEmail(ctx, userEmail); !ok {
		hash, err := auth.HashPassword("user")
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
	}

	return nil
}

func seedTestProposals(ctx context.Context, users persistence.UserRepository, proposals persistence.ProposalRepository) error {
	if err := seedTestUsers(ctx, users); err != nil {
		return err
	}
	guestUser, _ := users.GetUserByEmail(ctx, userEmail)
	userID := guestUser.ID
	adminUser, _ := users.GetUserByEmail(ctx, adminEmail)
	adminID := adminUser.ID

	if err := proposals.CreateProposal(ctx, &domain.Proposal{
		Title:                "Practical Distributed Systems in Go",
		Summary:              "This course explores how to design and reason about distributed systems using Go, with an emphasis on tradeoffs, failure modes, and operational simplicity rather than academic formalisms.",
		Qualifications:       "I have designed and operated distributed Go services involving queues, background workers, retries, idempotency, and partial failure handling in production environments.",
		TargetAudience:       "Intermediate Go developers who want to understand how real distributed systems behave and how to build resilient services.",
		LearningObjectives:   "- Understand common distributed systems failure modes\n- Design idempotent APIs and background jobs\n- Apply retries, backoff, and timeouts correctly\n- Reason about consistency and tradeoffs",
		Outline:              "1. What makes systems distributed\n2. Failure modes and fallacies\n3. Timeouts, retries, and idempotency\n4. Background workers and queues\n5. Consistency models in practice\n6. Observability and debugging",
		AssumedPrerequisites: "- Solid Go fundamentals\n- Basic HTTP and concurrency knowledge",
		AuthorID:             userID,
		Status:               domain.ProposalStatusDraft,
	}); err != nil {
		return err
	}

	if err := proposals.CreateProposal(ctx, &domain.Proposal{
		Title:                "Building Secure APIs in Go",
		Summary:              "Students will learn how to design and implement secure HTTP APIs in Go, covering authentication, authorization, input validation, and common attack vectors.",
		Qualifications:       "I have implemented authentication and authorization systems for production Go APIs, including token-based auth, session security, and password handling.",
		TargetAudience:       "Backend developers who want to build APIs that are secure by default.",
		LearningObjectives:   "- Implement secure authentication flows\n- Apply authorization patterns correctly\n- Prevent common web vulnerabilities\n- Validate and sanitize input safely",
		Outline:              "1. Threat modeling basics\n2. Authentication strategies\n3. Authorization patterns\n4. Input validation and encoding\n5. Common attacks and defenses\n6. Security testing and reviews",
		AssumedPrerequisites: "- Go web development experience\n- Basic understanding of HTTP",
		AuthorID:             userID,
		Status:               domain.ProposalStatusSubmitted,
	}); err != nil {
		return err
	}

	if err := proposals.CreateProposal(ctx, &domain.Proposal{
		Title:                "Designing and Shipping a Go Web App",
		Summary:              "In this course, students will build a real-world web application in Go from scratch, focusing on clean architecture, persistence boundaries, authentication, and deployment. The course emphasizes pragmatic decision-making and incremental design rather than frameworks or tutorials.",
		Qualifications:       "I built and deployed ByteCourses from scratch, including authentication, persistence, user workflows, and deployment to production infrastructure. This work involved designing clean domain boundaries, implementing both in-memory and SQL-backed storage, writing automated tests, and operating the system in a live environment.",
		TargetAudience:       "Intermediate developers who already know basic Go and want to learn how to design, structure, and ship a maintainable backend service with real users and real constraints.",
		LearningObjectives:   "- Implement authentication, authorization, and session management\n- Design clear domain, handler, service, and persistence boundaries in Go\n- Write effective tests at multiple layers (unit, integration, e2e)\n- Deploy a production Go service with a database and migration system",
		Outline:              "1. Project goals and architectural boundaries\n2. Domain modeling and invariants\n3. HTTP handlers and middleware design\n4. Persistence interfaces and store implementations\n5. Authentication, sessions, and password security\n6. Testing strategies (memstore vs SQL, API tests)\n7. Migrations and schema evolution\n8. Deployment and operational concerns",
		AssumedPrerequisites: "- Comfortable with Go syntax and tooling\n- Basic understanding of HTTP and REST APIs\n- Familiarity with SQL fundamentals is helpful but not required",
		AuthorID:             userID,
		Status:               domain.ProposalStatusApproved,
	}); err != nil {
		return err
	}

	if err := proposals.CreateProposal(ctx, &domain.Proposal{
		Title:                "Testing Go Applications End to End",
		Summary:              "This course focuses on building confidence in Go systems through effective testing strategies across unit, integration, and end-to-end layers.",
		Qualifications:       "I have written and maintained extensive automated test suites for Go services, including database-backed integration tests and full API-level tests.",
		TargetAudience:       "Go developers who want to improve test quality and reduce production regressions.",
		LearningObjectives:   "- Structure code for testability\n- Write meaningful unit and integration tests\n- Manage test data and environments\n- Balance test speed and coverage",
		Outline:              "1. Testing philosophy and tradeoffs\n2. Unit testing domains and services\n3. Integration testing with databases\n4. End-to-end API tests\n5. Test data and fixtures\n6. CI considerations",
		AssumedPrerequisites: "- Comfortable writing Go code\n- Familiarity with Go’s testing tools",
		AuthorID:             userID,
		Status:               domain.ProposalStatusChangesRequested,
		ReviewNotes:          "It seems that there should be more prerequisites than you have listed here.",
		ReviewerID:           &adminID,
	}); err != nil {
		return err
	}

	return nil
}
