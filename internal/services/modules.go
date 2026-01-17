package services

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"log/slog"
)

type ModuleService struct {
	modules store.ModuleStore
	logger  *ModuleLogger
}

func NewModuleService(modules store.ModuleStore, logger *slog.Logger) *ModuleService {
	return &ModuleService{
		modules: modules,
		logger:  NewModuleLogger(logger),
	}
}

type CreateModuleRequest struct {
	Title string `json:"title"`
}

func (r *CreateModuleRequest) IsValid() bool {
	return r.Title != ""
}

type UpdateModuleRequest struct {
	Title string `json:"title"`
}

type ReorderModulesRequest struct {
	ModuleIDs []int64 `json:"module_ids"`
}

func (r *ReorderModulesRequest) IsValid() bool {
	return len(r.ModuleIDs) > 0
}

func (s *ModuleService) CreateModule(ctx context.Context, course *domain.Course, user *domain.User, request *CreateModuleRequest) (*domain.Module, error) {
	if !course.IsTaughtBy(user) {
		return nil, ErrForbidden
	}
	if !request.IsValid() {
		return nil, ErrInvalidInput
	}

	module := &domain.Module{
		CourseID: course.ID,
		Title:    request.Title,
	}

	if err := s.modules.CreateModule(ctx, module); err != nil {
		s.logger.Error("module creation failed",
			"event", "module.creation",
			"course_id", course.ID,
			"user_id", user.ID,
			"title", request.Title,
			"error", err,
		)
		return nil, err
	}

	s.logger.Info("module.created",
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
		"title", module.Title,
	)

	return module, nil
}

func (s *ModuleService) GetModule(ctx context.Context, module *domain.Module, course *domain.Course, user *domain.User) (*domain.Module, error) {
	if !course.IsViewableBy(user) {
		return nil, ErrNotFound
	}
	if module.CourseID != course.ID {
		return nil, ErrNotFound
	}
	return module, nil
}

func (s *ModuleService) ListModules(ctx context.Context, course *domain.Course, user *domain.User) ([]domain.Module, error) {
	if !course.IsViewableBy(user) {
		return nil, ErrNotFound
	}
	return s.modules.ListModulesByCourseID(ctx, course.ID)
}

func (s *ModuleService) UpdateModule(ctx context.Context, module *domain.Module, course *domain.Course, user *domain.User, request *UpdateModuleRequest) error {
	if !course.IsTaughtBy(user) {
		return ErrForbidden
	}
	if module.CourseID != course.ID {
		return ErrNotFound
	}

	module.Title = request.Title

	err := s.modules.UpdateModule(ctx, module)
	if err != nil {
		s.logger.Error("module update failed",
			"event", "module.update",
			"module_id", module.ID,
			"course_id", course.ID,
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	s.logger.Info("module.updated",
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
	)

	return nil
}

func (s *ModuleService) DeleteModule(ctx context.Context, module *domain.Module, course *domain.Course, user *domain.User) error {
	if !course.IsTaughtBy(user) {
		return ErrForbidden
	}
	if module.CourseID != course.ID {
		return ErrNotFound
	}

	err := s.modules.DeleteModuleByID(ctx, module.ID)
	if err != nil {
		s.logger.Error("module deletion failed",
			"event", "module.deletion",
			"module_id", module.ID,
			"course_id", course.ID,
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	s.logger.Info("module.deleted",
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
	)

	return nil
}

func (s *ModuleService) ReorderModules(ctx context.Context, course *domain.Course, user *domain.User, request *ReorderModulesRequest) error {
	if !course.IsTaughtBy(user) {
		return ErrForbidden
	}
	if !request.IsValid() {
		return ErrInvalidInput
	}

	err := s.modules.ReorderModules(ctx, course.ID, request.ModuleIDs)
	if err != nil {
		s.logger.Error("module reorder failed",
			"event", "module.reorder",
			"course_id", course.ID,
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	s.logger.Info("module.reordered",
		"course_id", course.ID,
		"user_id", user.ID,
		"module_ids", request.ModuleIDs,
	)

	return nil
}

type ModuleLogger struct {
	base *slog.Logger
}

func NewModuleLogger(base *slog.Logger) *ModuleLogger {
	return &ModuleLogger{
		base: base.With("service", "module_service"),
	}
}

func (l *ModuleLogger) Info(event string, fields ...any) {
	args := []any{"event", event}
	args = append(args, fields...)
	l.base.Info("module event", args...)
}

func (l *ModuleLogger) Error(msg string, fields ...any) {
	l.base.Error(msg, fields...)
}
