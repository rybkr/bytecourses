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
	_ Command = (*CreateContentCommand)(nil)
	_ Command = (*UpdateContentCommand)(nil)
	_ Command = (*DeleteContentCommand)(nil)
	_ Command = (*PublishContentCommand)(nil)
)

var (
	_ Query = (*ListContentQuery)(nil)
	_ Query = (*GetContentQuery)(nil)
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

type CreateContentCommand struct {
	Type     domain.ContentType `json:"type"`
	ModuleID int64              `json:"module_id"`
	Title    string             `json:"title"`
	Order    int                `json:"order"`
	Format   string             `json:"format"`
	Content  string             `json:"content"`
	UserID   int64              `json:"user_id"`
}

func (c *CreateContentCommand) Validate(v *validation.Validator) {
	v.Field(string(c.Type), "type").Required()
	v.Field(c.ModuleID, "module_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(1).MaxLength(255).IsTrimmed()
	v.Field(c.UserID, "user_id").EntityID()
	if c.Type == domain.ContentTypeReading {
		v.Field(c.Format, "format").Required()
		if c.Format == string(domain.ReadingFormatMarkdown) {
			v.Field(c.Content, "content").Required()
		}
	}
}

func (s *ContentService) Create(ctx context.Context, cmd *CreateContentCommand) (domain.ContentItem, error) {
	if err := validation.Validate(cmd); err != nil {
		return nil, err
	}

	switch cmd.Type {
	case domain.ContentTypeReading:
		return s.createReading(ctx, cmd)
	default:
		return nil, errors.ErrInvalidInput
	}
}

func (s *ContentService) createReading(ctx context.Context, cmd *CreateContentCommand) (*domain.Reading, error) {
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

	event := domain.NewContentCreatedEvent(domain.ContentTypeReading, reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return &reading, nil
}

type UpdateContentCommand struct {
	Type      domain.ContentType `json:"type"`
	ContentID int64              `json:"content_id"`
	Title     string             `json:"title"`
	Order     int                `json:"order"`
	Format    string             `json:"format"`
	Content   string             `json:"content"`
	UserID    int64              `json:"user_id"`
}

func (c *UpdateContentCommand) Validate(v *validation.Validator) {
	v.Field(string(c.Type), "type").Required()
	v.Field(c.ContentID, "content_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(1).MaxLength(255).IsTrimmed()
	v.Field(c.UserID, "user_id").EntityID()
	if c.Type == domain.ContentTypeReading {
		v.Field(c.Format, "format").Required()
		if c.Format == string(domain.ReadingFormatMarkdown) {
			v.Field(c.Content, "content").Required()
		}
	}
}

func (s *ContentService) Update(ctx context.Context, cmd *UpdateContentCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	switch cmd.Type {
	case domain.ContentTypeReading:
		return s.updateReading(ctx, cmd)
	default:
		return errors.ErrInvalidInput
	}
}

func (s *ContentService) updateReading(ctx context.Context, cmd *UpdateContentCommand) error {
	reading, ok := s.Readings.GetByID(ctx, cmd.ContentID)
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

	event := domain.NewContentUpdatedEvent(domain.ContentTypeReading, reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type DeleteContentCommand struct {
	Type      domain.ContentType `json:"type"`
	ContentID int64              `json:"content_id"`
	UserID    int64              `json:"user_id"`
}

func (c *DeleteContentCommand) Validate(v *validation.Validator) {
	v.Field(string(c.Type), "type").Required()
	v.Field(c.ContentID, "content_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ContentService) Delete(ctx context.Context, cmd *DeleteContentCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	switch cmd.Type {
	case domain.ContentTypeReading:
		return s.deleteReading(ctx, cmd)
	default:
		return errors.ErrInvalidInput
	}
}

func (s *ContentService) deleteReading(ctx context.Context, cmd *DeleteContentCommand) error {
	reading, ok := s.Readings.GetByID(ctx, cmd.ContentID)
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

	if err := s.Readings.DeleteByID(ctx, cmd.ContentID); err != nil {
		return err
	}

	event := domain.NewContentDeletedEvent(domain.ContentTypeReading, reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type PublishContentCommand struct {
	Type      domain.ContentType `json:"type"`
	ContentID int64              `json:"content_id"`
	UserID    int64              `json:"user_id"`
}

func (c *PublishContentCommand) Validate(v *validation.Validator) {
	v.Field(string(c.Type), "type").Required()
	v.Field(c.ContentID, "content_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ContentService) Publish(ctx context.Context, cmd *PublishContentCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	switch cmd.Type {
	case domain.ContentTypeReading:
		return s.publishReading(ctx, cmd)
	default:
		return errors.ErrInvalidInput
	}
}

func (s *ContentService) publishReading(ctx context.Context, cmd *PublishContentCommand) error {
	reading, ok := s.Readings.GetByID(ctx, cmd.ContentID)
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

	event := domain.NewContentPublishedEvent(domain.ContentTypeReading, reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type ListContentQuery struct {
	ModuleID int64           `json:"module_id"`
	UserID   int64           `json:"user_id"`
	UserRole domain.UserRole `json:"user_role"`
}

func (s *ContentService) List(ctx context.Context, query *ListContentQuery) ([]domain.ContentItem, error) {
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

	items := make([]domain.ContentItem, len(readings))
	for i := range readings {
		items[i] = &readings[i]
	}

	return items, nil
}

type GetContentQuery struct {
	ContentID int64           `json:"content_id"`
	ModuleID  int64           `json:"module_id"`
	UserID    int64           `json:"user_id"`
	UserRole  domain.UserRole `json:"user_role"`
}

func (s *ContentService) Get(ctx context.Context, query *GetContentQuery) (domain.ContentItem, error) {
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

	reading, ok := s.Readings.GetByID(ctx, query.ContentID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if reading.ModuleID != query.ModuleID {
		return nil, errors.ErrNotFound
	}

	return reading, nil
}
