package services

import (
	"context"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/validation"
)

var (
	_ Command = (*CreateReadingCommand)(nil)
	_ Command = (*UpdateReadingCommand)(nil)
	_ Command = (*DeleteReadingCommand)(nil)
	_ Command = (*PublishReadingCommand)(nil)
)

var (
	_ Query = (*ListReadingsQuery)(nil)
	_ Query = (*GetReadingQuery)(nil)
)

type ContentService struct {
	Readings persistence.ReadingRepository
	Modules  persistence.ModuleRepository
	Courses  persistence.CourseRepository
	Events   events.EventBus
}

func NewContentService(
	readings persistence.ReadingRepository,
	modules persistence.ModuleRepository,
	courses persistence.CourseRepository,
	eventBus events.EventBus,
) *ContentService {
	return &ContentService{
		Readings: readings,
		Modules:  modules,
		Courses:  courses,
		Events:   eventBus,
	}
}

type CreateReadingCommand struct {
	ModuleID int64  `json:"module_id"`
	Title    string `json:"title"`
	Order    int    `json:"order"`
	Format   string `json:"format"`
	Content  string `json:"content"`
	UserID   int64  `json:"user_id"`
}

func (c *CreateReadingCommand) Validate(v *validation.Validator) {
	v.Field(c.ModuleID, "module_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(1).MaxLength(255).IsTrimmed()
	v.Field(c.Format, "format").Required()
	v.Field(c.UserID, "user_id").EntityID()
	if c.Format == string(domain.ReadingFormatMarkdown) {
		v.Field(c.Content, "content").Required()
	}
}

func (s *ContentService) CreateReading(ctx context.Context, cmd *CreateReadingCommand) (*domain.Reading, error) {
	if err := validation.Validate(cmd); err != nil {
		return nil, err
	}

	module, ok := s.Modules.GetByID(ctx, cmd.ModuleID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return nil, errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return nil, errors.ErrNotFound
	}

	if cmd.Format != string(domain.ReadingFormatMarkdown) {
		return nil, errors.ErrInvalidInput
	}

	reading := domain.Reading{
		BaseContentItem: domain.BaseContentItem{
			ModuleID: cmd.ModuleID,
			Title:    cmd.Title,
			Order:    cmd.Order,
			Status:   domain.ContentStatusDraft,
		},
		Format:  domain.ReadingFormat(cmd.Format),
		Content: &cmd.Content,
	}
	if err := s.Readings.Create(ctx, &reading); err != nil {
		return nil, err
	}

	event := domain.NewReadingCreatedEvent(reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return &reading, nil
}

type UpdateReadingCommand struct {
	ReadingID int64  `json:"reading_id"`
	Title     string `json:"title"`
	Order     int    `json:"order"`
	Format    string `json:"format"`
	Content   string `json:"content"`
	UserID    int64  `json:"user_id"`
}

func (c *UpdateReadingCommand) Validate(v *validation.Validator) {
	v.Field(c.ReadingID, "reading_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(1).MaxLength(255).IsTrimmed()
	v.Field(c.Format, "format").Required()
	v.Field(c.UserID, "user_id").EntityID()
	if c.Format == string(domain.ReadingFormatMarkdown) {
		v.Field(c.Content, "content").Required()
	}
}

func (s *ContentService) UpdateReading(ctx context.Context, cmd *UpdateReadingCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	reading, ok := s.Readings.GetByID(ctx, cmd.ReadingID)
	if !ok {
		return errors.ErrNotFound
	}

	module, ok := s.Modules.GetByID(ctx, reading.ModuleID)
	if !ok {
		return errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrNotFound
	}

	if cmd.Format != string(domain.ReadingFormatMarkdown) {
		return errors.ErrInvalidInput
	}

	reading.Title = cmd.Title
	reading.Order = cmd.Order
	reading.Format = domain.ReadingFormat(cmd.Format)
	reading.Content = &cmd.Content
	if err := s.Readings.Update(ctx, reading); err != nil {
		return err
	}

	event := domain.NewReadingUpdatedEvent(reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type DeleteReadingCommand struct {
	ReadingID int64 `json:"reading_id"`
	UserID    int64 `json:"user_id"`
}

func (c *DeleteReadingCommand) Validate(v *validation.Validator) {
	v.Field(c.ReadingID, "reading_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ContentService) DeleteReading(ctx context.Context, cmd *DeleteReadingCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	reading, ok := s.Readings.GetByID(ctx, cmd.ReadingID)
	if !ok {
		return errors.ErrNotFound
	}

	module, ok := s.Modules.GetByID(ctx, reading.ModuleID)
	if !ok {
		return errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrNotFound
	}

	if err := s.Readings.DeleteByID(ctx, cmd.ReadingID); err != nil {
		return err
	}

	event := domain.NewReadingDeletedEvent(reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type PublishReadingCommand struct {
	ReadingID int64 `json:"reading_id"`
	UserID    int64 `json:"user_id"`
}

func (c *PublishReadingCommand) Validate(v *validation.Validator) {
	v.Field(c.ReadingID, "reading_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ContentService) PublishReading(ctx context.Context, cmd *PublishReadingCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	reading, ok := s.Readings.GetByID(ctx, cmd.ReadingID)
	if !ok {
		return errors.ErrNotFound
	}

	module, ok := s.Modules.GetByID(ctx, reading.ModuleID)
	if !ok {
		return errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrNotFound
	}
	if reading.Status != domain.ContentStatusDraft {
		return errors.ErrInvalidStatusTransition
	}

	reading.Status = domain.ContentStatusPublished
	if err := s.Readings.Update(ctx, reading); err != nil {
		return err
	}

	event := domain.NewReadingPublishedEvent(reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type ListReadingsQuery struct {
	ModuleID int64           `json:"module_id"`
	UserID   int64           `json:"user_id"`
	UserRole domain.UserRole `json:"user_role"`
}

func (s *ContentService) ListReadings(ctx context.Context, query *ListReadingsQuery) ([]domain.Reading, error) {
	module, ok := s.Modules.GetByID(ctx, query.ModuleID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	switch query.UserRole {
	case domain.UserRoleStudent,
		domain.UserRoleInstructor:
		if course.InstructorID != query.UserID {
			return nil, errors.ErrNotFound
		}

	case domain.UserRoleAdmin:

	default:
		return nil, errors.ErrForbidden
	}

	readings, err := s.Readings.ListByModuleID(ctx, query.ModuleID)
	if err != nil {
		return nil, err
	}

	return readings, nil
}

type GetReadingQuery struct {
	ReadingID int64           `json:"reading_id"`
	ModuleID  int64           `json:"module_id"`
	UserID    int64           `json:"user_id"`
	UserRole  domain.UserRole `json:"user_role"`
}

func (s *ContentService) GetReading(ctx context.Context, query *GetReadingQuery) (*domain.Reading, error) {
	module, ok := s.Modules.GetByID(ctx, query.ModuleID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	switch query.UserRole {
	case domain.UserRoleStudent,
		domain.UserRoleInstructor:
		if course.InstructorID != query.UserID {
			return nil, errors.ErrNotFound
		}

	case domain.UserRoleAdmin:

	default:
		return nil, errors.ErrForbidden
	}

	reading, ok := s.Readings.GetByID(ctx, query.ReadingID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if reading.ModuleID != query.ModuleID {
		return nil, errors.ErrNotFound
	}

	return reading, nil
}
