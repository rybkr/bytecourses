package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	BaseURL      string
	DB           persistence.DB

	UserRepo          persistence.UserRepository
	ProposalRepo      persistence.ProposalRepository
	CourseRepo        persistence.CourseRepository
	ModuleRepo        persistence.ModuleRepository
	ReadingRepo       persistence.ReadingRepository
	PasswordResetRepo persistence.PasswordResetRepository
	EnrollmentRepo    persistence.EnrollmentRepository

	AuthService       *services.AuthService
	ProposalService   *services.ProposalService
	CourseService     *services.CourseService
	ModuleService     *services.ModuleService
	ContentService    *services.ContentService
	EnrollmentService *services.EnrollmentService

	onClose func() error
}

func NewContainer(ctx context.Context, cfg Config) (*Container, error) {
	c := Container{}

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

	c.BaseURL = strings.TrimSpace(cfg.BaseURL)
	if strings.HasSuffix(c.BaseURL, "/") && !strings.HasSuffix(c.BaseURL, "//") {
		c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")
	}
	if c.BaseURL != "" {
		u, err := url.Parse(c.BaseURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			slog.Warn("BASE_URL invalid or missing scheme/host", "base_url", c.BaseURL)
		}
	}

	seedAdmin(ctx, c.UserRepo)
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
	if cfg.SeedCourses != "" {
		if err := c.seedCourses(ctx, cfg.SeedCourses); err != nil {
			return nil, fmt.Errorf("seeding courses: %w", err)
		}
	}
	if cfg.SeedContent != "" {
		if err := c.seedContent(ctx, cfg.SeedContent); err != nil {
			return nil, fmt.Errorf("seeding content: %w", err)
		}
	}

	return &c, nil
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
		c.ReadingRepo = memory.NewReadingRepository()
		c.PasswordResetRepo = memory.NewPasswordResetRepository()
		c.EnrollmentRepo = memory.NewEnrollmentRepository()

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
		c.ReadingRepo = postgres.NewReadingRepository(db)
		c.PasswordResetRepo = postgres.NewPasswordResetRepository(db)
		c.EnrollmentRepo = postgres.NewEnrollmentRepository(db)
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

	c.ModuleService = services.NewModuleService(
		c.ModuleRepo,
		c.CourseRepo,
		c.EventBus,
	)

	c.ContentService = services.NewContentService(
		c.ReadingRepo,
		c.ModuleRepo,
		c.CourseRepo,
		c.EventBus,
	)

	c.EnrollmentService = services.NewEnrollmentService(
		c.EnrollmentRepo,
		c.CourseRepo,
		c.UserRepo,
		c.EventBus,
	)
}

func (c *Container) setupEventSubscribers() {
	c.EventBus.Subscribe("user.registered", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.UserRegisteredEvent)
		getStartedURL := c.BaseURL + "/"
		return c.EmailSender.SendWelcomeEmail(ctx, event.Email, event.Name, getStartedURL)
	})

	c.EventBus.Subscribe("user.password_reset_requested", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.PasswordResetRequestedEvent)
		return c.EmailSender.SendPasswordResetEmail(ctx, event.Email, event.ResetURL, event.Token)
	})

	c.EventBus.Subscribe("proposal.submitted", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.ProposalSubmittedEvent)
		author, ok := c.UserRepo.GetByID(ctx, event.AuthorID)
		if !ok {
			return nil
		}
		proposalURL := c.BaseURL + "/proposals/" + strconv.FormatInt(event.ProposalID, 10)
		return c.EmailSender.SendProposalSubmittedEmail(ctx, author.Email, author.Name, event.Title, proposalURL)
	})

	c.EventBus.Subscribe("proposal.approved", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.ProposalApprovedEvent)
		author, ok := c.UserRepo.GetByID(ctx, event.AuthorID)
		if !ok {
			return nil
		}
		courseURL := c.BaseURL + "/proposals/" + strconv.FormatInt(event.ProposalID, 10)
		return c.EmailSender.SendProposalApprovedEmail(ctx, author.Email, author.Name, event.Title, courseURL)
	})

	c.EventBus.Subscribe("proposal.rejected", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.ProposalRejectedEvent)
		author, ok := c.UserRepo.GetByID(ctx, event.AuthorID)
		if !ok {
			return nil
		}
		newProposalURL := c.BaseURL + "/proposals/new"
		return c.EmailSender.SendProposalRejectedEmail(ctx, author.Email, author.Name, event.Title, event.ReviewNotes, newProposalURL)
	})

	c.EventBus.Subscribe("proposal.changes_requested", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.ProposalChangesRequestedEvent)
		author, ok := c.UserRepo.GetByID(ctx, event.AuthorID)
		if !ok {
			return nil
		}
		proposalURL := c.BaseURL + "/proposals/" + strconv.FormatInt(event.ProposalID, 10) + "/edit"
		return c.EmailSender.SendProposalChangesRequestedEmail(ctx, author.Email, author.Name, event.Title, event.ReviewNotes, proposalURL)
	})

	c.EventBus.Subscribe("enrollment.created", func(ctx context.Context, e domain.Event) error {
		event := e.(*domain.EnrollmentCreatedEvent)
		user, ok := c.UserRepo.GetByID(ctx, event.UserID)
		if !ok {
			return nil
		}
		course, ok := c.CourseRepo.GetByID(ctx, event.CourseID)
		if !ok {
			return nil
		}
		courseURL := c.BaseURL + "/courses/" + strconv.FormatInt(event.CourseID, 10)
		return c.EmailSender.SendEnrollmentConfirmationEmail(ctx, user.Email, user.Name, course.Title, courseURL)
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
		Role:         domain.SystemRoleAdmin,
		Name:         "Admin",
	})
}

type seedUser struct {
	ID       int64             `json:"id"`
	Email    string            `json:"email"`
	Name     string            `json:"name"`
	Password string            `json:"password"`
	Role     domain.SystemRole `json:"role"`
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

type seedCourse struct {
	ID                   int64               `json:"id"`
	Title                string              `json:"title"`
	Summary              string              `json:"summary"`
	TargetAudience       string              `json:"target_audience"`
	LearningObjectives   string              `json:"learning_objectives"`
	AssumedPrerequisites string              `json:"assumed_prerequisites"`
	InstructorID         int64               `json:"instructor_id"`
	ProposalID           *int64              `json:"proposal_id"`
	Status               domain.CourseStatus `json:"status"`
}

func (c *Container) seedCourses(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var courses []seedCourse
	if err := json.Unmarshal(data, &courses); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	for _, sc := range courses {
		if _, ok := c.CourseRepo.GetByID(ctx, sc.ID); ok {
			continue
		}

		status := sc.Status
		if status == "" {
			status = domain.CourseStatusDraft
		}

		course := &domain.Course{
			ID:                   sc.ID,
			Title:                sc.Title,
			Summary:              sc.Summary,
			TargetAudience:       sc.TargetAudience,
			LearningObjectives:   sc.LearningObjectives,
			AssumedPrerequisites: sc.AssumedPrerequisites,
			InstructorID:         sc.InstructorID,
			ProposalID:           sc.ProposalID,
			Status:               status,
		}

		if err := c.CourseRepo.Create(ctx, course); err != nil {
			return fmt.Errorf("creating course %q: %w", sc.Title, err)
		}
	}

	return nil
}

type seedModule struct {
	ID          int64               `json:"id"`
	CourseID    int64               `json:"course_id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Order       int                 `json:"order"`
	Status      domain.ModuleStatus `json:"status"`
}

type seedReading struct {
	ID       int64                `json:"id"`
	ModuleID int64                `json:"module_id"`
	Title    string               `json:"title"`
	Order    int                  `json:"order"`
	Format   domain.ReadingFormat `json:"format"`
	Content  *string              `json:"content,omitempty"`
	Status   domain.ContentStatus `json:"status"`
}

type seedContentData struct {
	Modules  []seedModule  `json:"modules"`
	Readings []seedReading `json:"readings"`
}

func (c *Container) seedContent(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var seedData seedContentData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	// Map from seed module ID to actual created module ID
	// This handles the case where repositories auto-generate IDs
	moduleIDMap := make(map[int64]int64)

	// First, seed modules
	for _, sm := range seedData.Modules {
		// Verify course exists
		if _, ok := c.CourseRepo.GetByID(ctx, sm.CourseID); !ok {
			return fmt.Errorf("course %d not found for module %q", sm.CourseID, sm.Title)
		}

		// Check if module already exists by listing modules for the course
		// and matching by title and order (since IDs may be auto-generated)
		existingModules, err := c.ModuleRepo.ListByCourseID(ctx, sm.CourseID)
		if err != nil {
			return fmt.Errorf("listing modules for course %d: %w", sm.CourseID, err)
		}

		var existingModule *domain.Module
		for i := range existingModules {
			if existingModules[i].Title == sm.Title && existingModules[i].Order == sm.Order {
				existingModule = &existingModules[i]
				break
			}
		}

		if existingModule != nil {
			// Module already exists, use its ID for mapping
			if sm.ID > 0 {
				moduleIDMap[sm.ID] = existingModule.ID
			}
			continue
		}

		status := sm.Status
		if status == "" {
			status = domain.ModuleStatusDraft
		}

		module := &domain.Module{
			CourseID:    sm.CourseID,
			Title:       sm.Title,
			Description: sm.Description,
			Order:       sm.Order,
			Status:      status,
		}

		if err := c.ModuleRepo.Create(ctx, module); err != nil {
			return fmt.Errorf("creating module %q: %w", sm.Title, err)
		}

		// Map seed ID to actual ID
		if sm.ID > 0 {
			moduleIDMap[sm.ID] = module.ID
		}
	}

	// Then, seed readings
	for _, sr := range seedData.Readings {
		// Resolve actual module ID from mapping
		actualModuleID := sr.ModuleID
		if mappedID, ok := moduleIDMap[sr.ModuleID]; ok {
			actualModuleID = mappedID
		}

		// Verify module exists
		if _, ok := c.ModuleRepo.GetByID(ctx, actualModuleID); !ok {
			return fmt.Errorf("module %d not found for reading %q", sr.ModuleID, sr.Title)
		}

		// Check if reading already exists by listing readings for the module
		// and matching by title and order
		existingReadings, err := c.ReadingRepo.ListByModuleID(ctx, actualModuleID)
		if err != nil {
			return fmt.Errorf("listing readings for module %d: %w", actualModuleID, err)
		}

		var existingReading *domain.Reading
		for i := range existingReadings {
			if existingReadings[i].Title == sr.Title && existingReadings[i].Order == sr.Order {
				existingReading = &existingReadings[i]
				break
			}
		}

		if existingReading != nil {
			// Reading already exists, skip it
			continue
		}

		status := sr.Status
		if status == "" {
			status = domain.ContentStatusDraft
		}

		format := sr.Format
		if format == "" {
			format = domain.ReadingFormatMarkdown
		}

		reading := &domain.Reading{
			BaseContentItem: domain.BaseContentItem{
				ModuleID: actualModuleID,
				Title:    sr.Title,
				Order:    sr.Order,
				Status:   status,
			},
			Format:  format,
			Content: sr.Content,
		}

		if err := c.ReadingRepo.Create(ctx, reading); err != nil {
			return fmt.Errorf("creating reading %q: %w", sr.Title, err)
		}
	}

	return nil
}
