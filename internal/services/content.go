package services

import (
	"context"
	"io"
	"sort"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/infrastructure/storage"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/validation"
)

var (
	_ Command = (*CreateContentCommand)(nil)
	_ Command = (*UpdateContentCommand)(nil)
	_ Command = (*DeleteContentCommand)(nil)
	_ Command = (*PublishContentCommand)(nil)
	_ Command = (*UnpublishContentCommand)(nil)
)

var (
	_ Query = (*ListContentQuery)(nil)
	_ Query = (*GetContentQuery)(nil)
)

type ContentService struct {
	Readings    persistence.ReadingRepository
	Files       persistence.FileRepository
	Modules     persistence.ModuleRepository
	Courses     persistence.CourseRepository
	Events      events.EventBus
	FileStorage storage.FileStorage
}

func NewContentService(
	readings persistence.ReadingRepository,
	files persistence.FileRepository,
	modules persistence.ModuleRepository,
	courses persistence.CourseRepository,
	eventBus events.EventBus,
	fileStorage storage.FileStorage,
) *ContentService {
	return &ContentService{
		Readings:    readings,
		Files:       files,
		Modules:     modules,
		Courses:     courses,
		Events:      eventBus,
		FileStorage: fileStorage,
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
		format := domain.ReadingFormat(c.Format)
		if format == domain.ReadingFormatMarkdown {
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

type CreateFileCommand struct {
	ModuleID int64  `json:"module_id"`
	Title    string `json:"title"`
	Order    int    `json:"order"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
	UserID   int64  `json:"user_id"`
	Content  io.Reader
}

func (c *CreateFileCommand) Validate(v *validation.Validator) {
	v.Field(c.ModuleID, "module_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(1).MaxLength(255).IsTrimmed()
	v.Field(c.FileName, "file_name").Required()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ContentService) CreateFile(ctx context.Context, cmd *CreateFileCommand) (*domain.File, error) {
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
		return nil, errors.ErrForbidden
	}

	storagePath, err := s.FileStorage.Save(ctx, cmd.FileName, cmd.Content)
	if err != nil {
		return nil, err
	}

	file := domain.File{
		BaseContentItem: domain.BaseContentItem{
			ModuleID: cmd.ModuleID,
			Title:    cmd.Title,
			Order:    cmd.Order,
			Status:   domain.ContentStatusDraft,
		},
		FileName:    cmd.FileName,
		FileSize:    cmd.FileSize,
		MimeType:    cmd.MimeType,
		StoragePath: storagePath,
	}
	if err := s.Files.Create(ctx, &file); err != nil {
		s.FileStorage.Delete(ctx, storagePath)
		return nil, err
	}

	event := domain.NewContentCreatedEvent(domain.ContentTypeFile, file.ID, file.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return &file, nil
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

	format := domain.ReadingFormat(cmd.Format)
	if format != domain.ReadingFormatMarkdown && format != domain.ReadingFormatPlain && format != domain.ReadingFormatHTML {
		return nil, errors.ErrInvalidInput
	}

	reading := domain.Reading{
		BaseContentItem: domain.BaseContentItem{
			ModuleID: cmd.ModuleID,
			Title:    cmd.Title,
			Order:    cmd.Order,
			Status:   domain.ContentStatusDraft,
		},
		Format:  format,
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
		format := domain.ReadingFormat(c.Format)
		if format == domain.ReadingFormatMarkdown {
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
	case domain.ContentTypeFile:
		return s.updateFile(ctx, cmd)
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

	format := domain.ReadingFormat(cmd.Format)
	if format != domain.ReadingFormatMarkdown && format != domain.ReadingFormatPlain && format != domain.ReadingFormatHTML {
		return errors.ErrInvalidInput
	}

	reading.Title = cmd.Title
	reading.Order = cmd.Order
	reading.Format = format
	reading.Content = &cmd.Content
	if err := s.Readings.Update(ctx, reading); err != nil {
		return err
	}

	event := domain.NewContentUpdatedEvent(domain.ContentTypeReading, reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

func (s *ContentService) updateFile(ctx context.Context, cmd *UpdateContentCommand) error {
	file, ok := s.Files.GetByID(ctx, cmd.ContentID)
	if !ok {
		return errors.ErrNotFound
	}

	module, ok := s.Modules.GetByID(ctx, file.ModuleID)
	if !ok {
		return errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrForbidden
	}

	file.Title = cmd.Title
	file.Order = cmd.Order
	if err := s.Files.Update(ctx, file); err != nil {
		return err
	}

	event := domain.NewContentUpdatedEvent(domain.ContentTypeFile, file.ID, file.ModuleID, module.CourseID, course.InstructorID)
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
	case domain.ContentTypeFile:
		return s.deleteFile(ctx, cmd)
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

func (s *ContentService) deleteFile(ctx context.Context, cmd *DeleteContentCommand) error {
	file, ok := s.Files.GetByID(ctx, cmd.ContentID)
	if !ok {
		return errors.ErrNotFound
	}

	module, ok := s.Modules.GetByID(ctx, file.ModuleID)
	if !ok {
		return errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrForbidden
	}

	if err := s.Files.DeleteByID(ctx, cmd.ContentID); err != nil {
		return err
	}

	_ = s.FileStorage.Delete(ctx, file.StoragePath)

	event := domain.NewContentDeletedEvent(domain.ContentTypeFile, file.ID, file.ModuleID, module.CourseID, course.InstructorID)
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
	case domain.ContentTypeFile:
		return s.publishFile(ctx, cmd)
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

func (s *ContentService) publishFile(ctx context.Context, cmd *PublishContentCommand) error {
	file, ok := s.Files.GetByID(ctx, cmd.ContentID)
	if !ok {
		return errors.ErrNotFound
	}

	module, ok := s.Modules.GetByID(ctx, file.ModuleID)
	if !ok {
		return errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrForbidden
	}
	if file.Status != domain.ContentStatusDraft {
		return errors.ErrInvalidStatusTransition
	}

	file.Status = domain.ContentStatusPublished
	if err := s.Files.Update(ctx, file); err != nil {
		return err
	}

	event := domain.NewContentPublishedEvent(domain.ContentTypeFile, file.ID, file.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type UnpublishContentCommand struct {
	Type      domain.ContentType `json:"type"`
	ContentID int64              `json:"content_id"`
	UserID    int64              `json:"user_id"`
}

func (c *UnpublishContentCommand) Validate(v *validation.Validator) {
	v.Field(string(c.Type), "type").Required()
	v.Field(c.ContentID, "content_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ContentService) Unpublish(ctx context.Context, cmd *UnpublishContentCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	switch cmd.Type {
	case domain.ContentTypeReading:
		return s.unpublishReading(ctx, cmd)
	case domain.ContentTypeFile:
		return s.unpublishFile(ctx, cmd)
	default:
		return errors.ErrInvalidInput
	}
}

func (s *ContentService) unpublishReading(ctx context.Context, cmd *UnpublishContentCommand) error {
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
	if reading.Status != domain.ContentStatusPublished {
		return errors.ErrInvalidStatusTransition
	}

	reading.Status = domain.ContentStatusDraft
	if err := s.Readings.Update(ctx, reading); err != nil {
		return err
	}

	event := domain.NewContentUnpublishedEvent(domain.ContentTypeReading, reading.ID, reading.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

func (s *ContentService) unpublishFile(ctx context.Context, cmd *UnpublishContentCommand) error {
	file, ok := s.Files.GetByID(ctx, cmd.ContentID)
	if !ok {
		return errors.ErrNotFound
	}

	module, ok := s.Modules.GetByID(ctx, file.ModuleID)
	if !ok {
		return errors.ErrNotFound
	}

	course, ok := s.Courses.GetByID(ctx, module.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrForbidden
	}
	if file.Status != domain.ContentStatusPublished {
		return errors.ErrInvalidStatusTransition
	}

	file.Status = domain.ContentStatusDraft
	if err := s.Files.Update(ctx, file); err != nil {
		return err
	}

	event := domain.NewContentUnpublishedEvent(domain.ContentTypeFile, file.ID, file.ModuleID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type ListContentQuery struct {
	ModuleID        int64             `json:"module_id"`
	UserID          int64             `json:"user_id"`
	UserRole        domain.SystemRole `json:"user_role"`
	EnrolledLearner bool              `json:"enrolled_learner"`
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

	if query.UserRole == domain.SystemRoleAdmin {
	} else if course.InstructorID == query.UserID {
	} else if query.EnrolledLearner {
	} else {
		return nil, errors.ErrForbidden
	}

	readings, err := s.Readings.ListByModuleID(ctx, query.ModuleID)
	if err != nil {
		return nil, err
	}

	files, err := s.Files.ListByModuleID(ctx, query.ModuleID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.ContentItem, 0, len(readings)+len(files))

	for i := range readings {
		if query.EnrolledLearner && readings[i].Status != domain.ContentStatusPublished {
			continue
		}
		items = append(items, &readings[i])
	}

	for i := range files {
		if query.EnrolledLearner && files[i].Status != domain.ContentStatusPublished {
			continue
		}
		items = append(items, &files[i])
	}

	sort.SliceStable(items, func(i, j int) bool {
		return getOrder(items[i]) < getOrder(items[j])
	})

	return items, nil
}

func getOrder(item domain.ContentItem) int {
	switch v := item.(type) {
	case *domain.Reading:
		return v.Order
	case *domain.File:
		return v.Order
	default:
		return 0
	}
}

type GetContentQuery struct {
	ContentID       int64             `json:"content_id"`
	ModuleID        int64             `json:"module_id"`
	UserID          int64             `json:"user_id"`
	UserRole        domain.SystemRole `json:"user_role"`
	EnrolledLearner bool              `json:"enrolled_learner"`
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

	if query.UserRole == domain.SystemRoleAdmin {
	} else if course.InstructorID == query.UserID {
	} else if query.EnrolledLearner {
	} else {
		return nil, errors.ErrForbidden
	}

	if reading, ok := s.Readings.GetByID(ctx, query.ContentID); ok {
		if reading.ModuleID != query.ModuleID {
			return nil, errors.ErrNotFound
		}
		if query.EnrolledLearner && reading.Status != domain.ContentStatusPublished {
			return nil, errors.ErrNotFound
		}
		return reading, nil
	}

	if file, ok := s.Files.GetByID(ctx, query.ContentID); ok {
		if file.ModuleID != query.ModuleID {
			return nil, errors.ErrNotFound
		}
		if query.EnrolledLearner && file.Status != domain.ContentStatusPublished {
			return nil, errors.ErrNotFound
		}
		return file, nil
	}

	return nil, errors.ErrNotFound
}

func (s *ContentService) GetFileURL(file *domain.File) string {
	return s.FileStorage.GetPath(file.StoragePath)
}
