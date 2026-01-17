package services

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"log/slog"
)

type ContentService struct {
	content store.ContentStore
	logger  *ContentLogger
}

func NewContentService(content store.ContentStore, logger *slog.Logger) *ContentService {
	return &ContentService{
		content: content,
		logger:  NewContentLogger(logger),
	}
}

type CreateLectureRequest struct {
	Title string `json:"title"`
}

func (r *CreateLectureRequest) IsValid() bool {
	return r.Title != ""
}

type UpdateLectureRequest struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}

type ReorderContentRequest struct {
	ItemIDs []int64 `json:"item_ids"`
}

func (r *ReorderContentRequest) IsValid() bool {
	return len(r.ItemIDs) > 0
}

func (s *ContentService) CreateLecture(ctx context.Context, module *domain.Module, course *domain.Course, user *domain.User, request *CreateLectureRequest) (*domain.ContentItem, error) {
	if !course.IsTaughtBy(user) {
		return nil, ErrForbidden
	}
	if module.CourseID != course.ID {
		return nil, ErrNotFound
	}
	if !request.IsValid() {
		return nil, ErrInvalidInput
	}

	item := &domain.ContentItem{
		ModuleID: module.ID,
		Title:    request.Title,
		Type:     domain.ContentTypeLecture,
		Status:   domain.ContentStatusDraft,
	}

	if err := s.content.CreateContentItem(ctx, item); err != nil {
		s.logger.Error("content item creation failed",
			"event", "content.creation",
			"module_id", module.ID,
			"course_id", course.ID,
			"user_id", user.ID,
			"title", request.Title,
			"error", err,
		)
		return nil, err
	}

	lecture := &domain.Lecture{
		ContentItemID: item.ID,
		Content:       "",
	}
	if err := s.content.UpsertLecture(ctx, lecture); err != nil {
		s.logger.Error("lecture creation failed",
			"event", "lecture.creation",
			"content_item_id", item.ID,
			"error", err,
		)
		return nil, err
	}

	s.logger.Info("content.created",
		"content_item_id", item.ID,
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
		"title", item.Title,
		"type", item.Type,
	)

	return item, nil
}

func (s *ContentService) GetLecture(ctx context.Context, item *domain.ContentItem, module *domain.Module, course *domain.Course, user *domain.User) (*domain.ContentItem, *domain.Lecture, error) {
	if !course.IsViewableBy(user) {
		return nil, nil, ErrNotFound
	}
	if module.CourseID != course.ID {
		return nil, nil, ErrNotFound
	}
	if item.ModuleID != module.ID {
		return nil, nil, ErrNotFound
	}

	// Non-instructors can only see published content
	if !course.IsTaughtBy(user) && item.Status != domain.ContentStatusPublished {
		return nil, nil, ErrNotFound
	}

	lecture, ok := s.content.GetLecture(ctx, item.ID)
	if !ok {
		lecture = &domain.Lecture{
			ContentItemID: item.ID,
			Content:       "",
		}
	}

	return item, lecture, nil
}

func (s *ContentService) UpdateLecture(ctx context.Context, item *domain.ContentItem, module *domain.Module, course *domain.Course, user *domain.User, request *UpdateLectureRequest) error {
	if !course.IsTaughtBy(user) {
		return ErrForbidden
	}
	if module.CourseID != course.ID {
		return ErrNotFound
	}
	if item.ModuleID != module.ID {
		return ErrNotFound
	}

	if request.Title != nil {
		item.Title = *request.Title
	}

	if err := s.content.UpdateContentItem(ctx, item); err != nil {
		s.logger.Error("content item update failed",
			"event", "content.update",
			"content_item_id", item.ID,
			"error", err,
		)
		return err
	}

	if request.Content != nil {
		lecture := &domain.Lecture{
			ContentItemID: item.ID,
			Content:       *request.Content,
		}
		if err := s.content.UpsertLecture(ctx, lecture); err != nil {
			s.logger.Error("lecture update failed",
				"event", "lecture.update",
				"content_item_id", item.ID,
				"error", err,
			)
			return err
		}
	}

	s.logger.Info("content.updated",
		"content_item_id", item.ID,
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
	)

	return nil
}

func (s *ContentService) DeleteContent(ctx context.Context, item *domain.ContentItem, module *domain.Module, course *domain.Course, user *domain.User) error {
	if !course.IsTaughtBy(user) {
		return ErrForbidden
	}
	if module.CourseID != course.ID {
		return ErrNotFound
	}
	if item.ModuleID != module.ID {
		return ErrNotFound
	}

	if err := s.content.DeleteContentItemByID(ctx, item.ID); err != nil {
		s.logger.Error("content deletion failed",
			"event", "content.deletion",
			"content_item_id", item.ID,
			"module_id", module.ID,
			"course_id", course.ID,
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	s.logger.Info("content.deleted",
		"content_item_id", item.ID,
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
	)

	return nil
}

func (s *ContentService) ListContent(ctx context.Context, module *domain.Module, course *domain.Course, user *domain.User) ([]domain.ContentItem, map[int64]*domain.Lecture, error) {
	if !course.IsViewableBy(user) {
		return nil, nil, ErrNotFound
	}
	if module.CourseID != course.ID {
		return nil, nil, ErrNotFound
	}

	items, lectures, err := s.content.ListContentItemsWithLecturesByModuleID(ctx, module.ID)
	if err != nil {
		return nil, nil, err
	}

	// Filter unpublished content for non-instructors
	if !course.IsTaughtBy(user) {
		filtered := make([]domain.ContentItem, 0, len(items))
		filteredLectures := make(map[int64]*domain.Lecture)
		for _, item := range items {
			if item.Status == domain.ContentStatusPublished {
				filtered = append(filtered, item)
				if lec, ok := lectures[item.ID]; ok {
					filteredLectures[item.ID] = lec
				}
			}
		}
		return filtered, filteredLectures, nil
	}

	return items, lectures, nil
}

func (s *ContentService) ReorderContent(ctx context.Context, module *domain.Module, course *domain.Course, user *domain.User, request *ReorderContentRequest) error {
	if !course.IsTaughtBy(user) {
		return ErrForbidden
	}
	if module.CourseID != course.ID {
		return ErrNotFound
	}
	if !request.IsValid() {
		return ErrInvalidInput
	}

	err := s.content.ReorderContentItems(ctx, module.ID, request.ItemIDs)
	if err != nil {
		s.logger.Error("content reorder failed",
			"event", "content.reorder",
			"module_id", module.ID,
			"course_id", course.ID,
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	s.logger.Info("content.reordered",
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
		"item_ids", request.ItemIDs,
	)

	return nil
}

func (s *ContentService) PublishContent(ctx context.Context, item *domain.ContentItem, module *domain.Module, course *domain.Course, user *domain.User) error {
	if !course.IsTaughtBy(user) {
		return ErrForbidden
	}
	if module.CourseID != course.ID {
		return ErrNotFound
	}
	if item.ModuleID != module.ID {
		return ErrNotFound
	}

	item.Status = domain.ContentStatusPublished

	if err := s.content.UpdateContentItem(ctx, item); err != nil {
		s.logger.Error("content publish failed",
			"event", "content.publish",
			"content_item_id", item.ID,
			"error", err,
		)
		return err
	}

	s.logger.Info("content.published",
		"content_item_id", item.ID,
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
	)

	return nil
}

func (s *ContentService) UnpublishContent(ctx context.Context, item *domain.ContentItem, module *domain.Module, course *domain.Course, user *domain.User) error {
	if !course.IsTaughtBy(user) {
		return ErrForbidden
	}
	if module.CourseID != course.ID {
		return ErrNotFound
	}
	if item.ModuleID != module.ID {
		return ErrNotFound
	}

	item.Status = domain.ContentStatusDraft

	if err := s.content.UpdateContentItem(ctx, item); err != nil {
		s.logger.Error("content unpublish failed",
			"event", "content.unpublish",
			"content_item_id", item.ID,
			"error", err,
		)
		return err
	}

	s.logger.Info("content.unpublished",
		"content_item_id", item.ID,
		"module_id", module.ID,
		"course_id", course.ID,
		"user_id", user.ID,
	)

	return nil
}

type ContentLogger struct {
	base *slog.Logger
}

func NewContentLogger(base *slog.Logger) *ContentLogger {
	return &ContentLogger{
		base: base.With("service", "content_service"),
	}
}

func (l *ContentLogger) Info(event string, fields ...any) {
	args := []any{"event", event}
	args = append(args, fields...)
	l.base.Info("content event", args...)
}

func (l *ContentLogger) Error(msg string, fields ...any) {
	l.base.Error(msg, fields...)
}
