package test

import (
	"context"
	"testing"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

type NewReadingRepository func(t *testing.T) persistence.ReadingRepository

func TestReadingRepository(t *testing.T, newReadingRepo NewReadingRepository, newModuleRepo NewModuleRepository, newCourseRepo NewCourseRepository, newUserRepo NewUserRepository) {
	t.Helper()

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)
		readings := newReadingRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Test Course",
			Summary:      "A test course",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		m := domain.Module{
			CourseID:    c.ID,
			Title:       "Test Module",
			Description: "A test module",
			Order:       1,
			Status:      domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		content := "# Test Reading\n\nThis is test content."
		r := domain.Reading{
			BaseContentItem: domain.BaseContentItem{
				ModuleID: m.ID,
				Title:    "Test Reading",
				Order:    1,
				Status:   domain.ContentStatusDraft,
			},
			Format:  domain.ReadingFormatMarkdown,
			Content: &content,
		}
		if err := readings.Create(ctx, &r); err != nil {
			t.Fatalf("readings.Create failed: %v", err)
		}
		if r.ID == 0 {
			t.Fatalf("readings.Create: ID not set")
		}
		if r.CreatedAt.IsZero() {
			t.Fatalf("readings.Create: CreatedAt not set")
		}
		if r.UpdatedAt.IsZero() {
			t.Fatalf("readings.Create: UpdatedAt not set")
		}
		if r.Content == nil || *r.Content != content {
			t.Fatalf("readings.Create: content not set correctly")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)
		readings := newReadingRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Test Course",
			Summary:      "A test course",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		m := domain.Module{
			CourseID:    c.ID,
			Title:       "Test Module",
			Description: "A test module",
			Order:       1,
			Status:      domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		content := "# Test Reading\n\nThis is test content."
		r := domain.Reading{
			BaseContentItem: domain.BaseContentItem{
				ModuleID: m.ID,
				Title:    "Test Reading",
				Order:    1,
				Status:   domain.ContentStatusDraft,
			},
			Format:  domain.ReadingFormatMarkdown,
			Content: &content,
		}
		if err := readings.Create(ctx, &r); err != nil {
			t.Fatalf("readings.Create failed: %v", err)
		}

		v, ok := readings.GetByID(ctx, r.ID)
		if !ok {
			t.Fatalf("readings.GetByID failed")
		}
		if v.ID != r.ID || v.Title != r.Title || v.ModuleID != r.ModuleID {
			t.Fatalf("readings.GetByID: readings differ")
		}
		if v.Content == nil || *v.Content != content {
			t.Fatalf("readings.GetByID: content not retrieved correctly")
		}
		if v.Format != domain.ReadingFormatMarkdown {
			t.Fatalf("readings.GetByID: format not retrieved correctly")
		}
	})

	t.Run("GetByIDNotFound", func(t *testing.T) {
		ctx := context.Background()
		readings := newReadingRepo(t)

		_, ok := readings.GetByID(ctx, -1)
		if ok {
			t.Fatalf("readings.GetByID: should return false for non-existent reading")
		}
	})

	t.Run("ListByModuleID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)
		readings := newReadingRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Test Course",
			Summary:      "A test course",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		m1 := domain.Module{
			CourseID:    c.ID,
			Title:       "Module 1",
			Description: "Module 1",
			Order:       1,
			Status:      domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m1); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		m2 := domain.Module{
			CourseID:    c.ID,
			Title:       "Module 2",
			Description: "Module 2",
			Order:       2,
			Status:      domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m2); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		content1 := "# Reading 1"
		r1 := domain.Reading{
			BaseContentItem: domain.BaseContentItem{
				ModuleID: m1.ID,
				Title:    "Reading 1",
				Order:    2,
				Status:   domain.ContentStatusDraft,
			},
			Format:  domain.ReadingFormatMarkdown,
			Content: &content1,
		}
		if err := readings.Create(ctx, &r1); err != nil {
			t.Fatalf("readings.Create failed: %v", err)
		}

		content2 := "# Reading 2"
		r2 := domain.Reading{
			BaseContentItem: domain.BaseContentItem{
				ModuleID: m1.ID,
				Title:    "Reading 2",
				Order:    1,
				Status:   domain.ContentStatusDraft,
			},
			Format:  domain.ReadingFormatMarkdown,
			Content: &content2,
		}
		if err := readings.Create(ctx, &r2); err != nil {
			t.Fatalf("readings.Create failed: %v", err)
		}

		content3 := "# Reading 3"
		r3 := domain.Reading{
			BaseContentItem: domain.BaseContentItem{
				ModuleID: m2.ID,
				Title:    "Reading 3",
				Order:    1,
				Status:   domain.ContentStatusDraft,
			},
			Format:  domain.ReadingFormatMarkdown,
			Content: &content3,
		}
		if err := readings.Create(ctx, &r3); err != nil {
			t.Fatalf("readings.Create failed: %v", err)
		}

		list, err := readings.ListByModuleID(ctx, m1.ID)
		if err != nil {
			t.Fatalf("readings.ListByModuleID failed: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("readings.ListByModuleID: expected 2 readings, got %d", len(list))
		}
		if list[0].ID != r2.ID || list[1].ID != r1.ID {
			t.Fatalf("readings.ListByModuleID: readings not in correct order")
		}
	})

	t.Run("ListByModuleIDEmpty", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)
		readings := newReadingRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Test Course",
			Summary:      "A test course",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		m := domain.Module{
			CourseID:    c.ID,
			Title:       "Test Module",
			Description: "A test module",
			Order:       1,
			Status:      domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		list, err := readings.ListByModuleID(ctx, m.ID)
		if err != nil {
			t.Fatalf("readings.ListByModuleID failed: %v", err)
		}
		if list == nil {
			t.Fatalf("readings.ListByModuleID: should return empty slice, not nil")
		}
		if len(list) != 0 {
			t.Fatalf("readings.ListByModuleID: expected empty slice, got %d items", len(list))
		}
	})

	t.Run("Update", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)
		readings := newReadingRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Test Course",
			Summary:      "A test course",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		m := domain.Module{
			CourseID:    c.ID,
			Title:       "Test Module",
			Description: "A test module",
			Order:       1,
			Status:      domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		originalContent := "# Original"
		r := domain.Reading{
			BaseContentItem: domain.BaseContentItem{
				ModuleID: m.ID,
				Title:    "Original Title",
				Order:    1,
				Status:   domain.ContentStatusDraft,
			},
			Format:  domain.ReadingFormatMarkdown,
			Content: &originalContent,
		}
		if err := readings.Create(ctx, &r); err != nil {
			t.Fatalf("readings.Create failed: %v", err)
		}

		id := r.ID
		originalUpdatedAt := r.UpdatedAt

		v, ok := readings.GetByID(ctx, id)
		if !ok {
			t.Fatalf("readings.GetByID failed")
		}

		updatedContent := "# Updated"
		v.Title = "Updated Title"
		v.Order = 2
		v.Status = domain.ContentStatusPublished
		v.Content = &updatedContent
		if r.Title != "Original Title" {
			t.Fatalf("ReadingRepository: external modification affected persisted value")
		}

		if err := readings.Update(ctx, v); err != nil {
			t.Fatalf("readings.Update failed: %v", err)
		}

		w, ok := readings.GetByID(ctx, id)
		if !ok {
			t.Fatalf("readings.GetByID failed")
		}
		if w.Title != "Updated Title" {
			t.Fatalf("readings.Update: title not updated")
		}
		if w.Order != 2 {
			t.Fatalf("readings.Update: order not updated")
		}
		if w.Status != domain.ContentStatusPublished {
			t.Fatalf("readings.Update: status not updated")
		}
		if w.Content == nil || *w.Content != updatedContent {
			t.Fatalf("readings.Update: content not updated")
		}
		if !w.UpdatedAt.After(originalUpdatedAt) {
			t.Fatalf("readings.Update: UpdatedAt not updated")
		}
	})

	t.Run("DeleteByID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)
		readings := newReadingRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Test Course",
			Summary:      "A test course",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		m := domain.Module{
			CourseID:    c.ID,
			Title:       "Test Module",
			Description: "A test module",
			Order:       1,
			Status:      domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		content := "# Test Reading"
		r := domain.Reading{
			BaseContentItem: domain.BaseContentItem{
				ModuleID: m.ID,
				Title:    "Test Reading",
				Order:    1,
				Status:   domain.ContentStatusDraft,
			},
			Format:  domain.ReadingFormatMarkdown,
			Content: &content,
		}
		if err := readings.Create(ctx, &r); err != nil {
			t.Fatalf("readings.Create failed: %v", err)
		}

		if err := readings.DeleteByID(ctx, r.ID); err != nil {
			t.Fatalf("readings.DeleteByID failed: %v", err)
		}

		_, ok := readings.GetByID(ctx, r.ID)
		if ok {
			t.Fatalf("readings.DeleteByID: reading still exists after deletion")
		}
	})
}
