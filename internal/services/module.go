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
	_ Command = (*CreateModuleCommand)(nil)
	_ Command = (*UpdateModuleCommand)(nil)
	_ Command = (*DeleteModuleCommand)(nil)
	_ Command = (*PublishModuleCommand)(nil)
)

var (
	_ Query = (*ListModulesQuery)(nil)
)

type ModuleService struct {
	Modules persistence.ModuleRepository
	Courses persistence.CourseRepository
	Events  events.EventBus
}

func NewModuleService(
	modules persistence.ModuleRepository,
	courses persistence.CourseRepository,
	eventBus events.EventBus,
) *ModuleService {
	return &ModuleService{
		Modules: modules,
		Courses: courses,
		Events:  eventBus,
	}
}

type CreateModuleCommand struct {
	CourseID    int64  `json:"course_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	UserID      int64  `json:"user_id"`
}

func (c *CreateModuleCommand) Validate(v *validation.Validator) {
	v.Field(c.CourseID, "course_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(1).MaxLength(255).IsTrimmed()
	v.Field(c.Description, "description").MaxLength(2048).IsTrimmed()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ModuleService) Create(ctx context.Context, cmd *CreateModuleCommand) (*domain.Module, error) {
	if err := validation.Validate(cmd); err != nil {
		return nil, err
	}

	course, ok := s.Courses.GetByID(ctx, cmd.CourseID)
	if !ok {
		return nil, errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return nil, errors.ErrNotFound
	}

	module := domain.Module{
		CourseID:    cmd.CourseID,
		Title:       cmd.Title,
		Description: cmd.Description,
		Order:       cmd.Order,
		Status:      domain.ModuleStatusDraft,
	}
	if err := s.Modules.Create(ctx, &module); err != nil {
		return nil, err
	}

	event := domain.NewModuleCreatedEvent(module.ID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return &module, nil
}

type UpdateModuleCommand struct {
	ModuleID    int64  `json:"module_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	UserID      int64  `json:"user_id"`
}

func (c *UpdateModuleCommand) Validate(v *validation.Validator) {
	v.Field(c.ModuleID, "module_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(1).MaxLength(255).IsTrimmed()
	v.Field(c.Description, "description").MaxLength(2048).IsTrimmed()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ModuleService) Update(ctx context.Context, cmd *UpdateModuleCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	module, ok := s.Modules.GetByID(ctx, cmd.ModuleID)
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

	module.Title = cmd.Title
	module.Description = cmd.Description
	module.Order = cmd.Order
	if err := s.Modules.Update(ctx, module); err != nil {
		return err
	}

	event := domain.NewModuleUpdatedEvent(module.ID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type DeleteModuleCommand struct {
	ModuleID int64 `json:"module_id"`
	UserID   int64 `json:"user_id"`
}

func (c *DeleteModuleCommand) Validate(v *validation.Validator) {
	v.Field(c.ModuleID, "module_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ModuleService) Delete(ctx context.Context, cmd *DeleteModuleCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	module, ok := s.Modules.GetByID(ctx, cmd.ModuleID)
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

	if err := s.Modules.DeleteByID(ctx, cmd.ModuleID); err != nil {
		return err
	}

	event := domain.NewModuleDeletedEvent(module.ID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type PublishModuleCommand struct {
	ModuleID int64 `json:"module_id"`
	UserID   int64 `json:"user_id"`
}

func (c *PublishModuleCommand) Validate(v *validation.Validator) {
	v.Field(c.ModuleID, "module_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ModuleService) Publish(ctx context.Context, cmd *PublishModuleCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	module, ok := s.Modules.GetByID(ctx, cmd.ModuleID)
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
	if module.Status != domain.ModuleStatusDraft {
		return errors.ErrInvalidStatusTransition
	}

	module.Status = domain.ModuleStatusPublished
	if err := s.Modules.Update(ctx, module); err != nil {
		return err
	}

	event := domain.NewModulePublishedEvent(module.ID, module.CourseID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type ListModulesQuery struct {
	CourseID int64           `json:"course_id"`
	UserID   int64           `json:"user_id"`
	UserRole domain.UserRole `json:"user_role"`
}

func (s *ModuleService) List(ctx context.Context, query *ListModulesQuery) ([]domain.Module, error) {
	course, ok := s.Courses.GetByID(ctx, query.CourseID)
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

	modules, err := s.Modules.ListByCourseID(ctx, query.CourseID)
	if err != nil {
		return nil, err
	}

	return modules, nil
}
