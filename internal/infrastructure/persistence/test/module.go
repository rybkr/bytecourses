package test

import (
	"context"
	"testing"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

type NewModuleRepository func(t *testing.T) persistence.ModuleRepository

func TestModuleRepository(t *testing.T, newModuleRepo NewModuleRepository, newCourseRepo NewCourseRepository, newUserRepo NewUserRepository) {
	t.Helper()

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)

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
			Title:        "Test Module",
			Description:  "A test module",
			Order:        1,
			Status:       domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}
		if m.ID == 0 {
			t.Fatalf("modules.Create: ID not set")
		}
		if m.CreatedAt.IsZero() {
			t.Fatalf("modules.Create: CreatedAt not set")
		}
		if m.UpdatedAt.IsZero() {
			t.Fatalf("modules.Create: UpdatedAt not set")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)

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
			Title:        "Test Module",
			Description:  "A test module",
			Order:        1,
			Status:       domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		v, ok := modules.GetByID(ctx, m.ID)
		if !ok {
			t.Fatalf("modules.GetByID failed")
		}
		if v.ID != m.ID || v.Title != m.Title || v.CourseID != m.CourseID {
			t.Fatalf("modules.GetByID: modules differ")
		}
	})

	t.Run("GetByIDNotFound", func(t *testing.T) {
		ctx := context.Background()
		modules := newModuleRepo(t)

		_, ok := modules.GetByID(ctx, -1)
		if ok {
			t.Fatalf("modules.GetByID: should return false for non-existent module")
		}
	})

	t.Run("ListByCourseID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c1 := domain.Course{
			Title:        "Course 1",
			Summary:      "Course 1",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c1); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		c2 := domain.Course{
			Title:        "Course 2",
			Summary:      "Course 2",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c2); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		m1 := domain.Module{
			CourseID:    c1.ID,
			Title:        "Module 1",
			Description:  "Module 1",
			Order:        2,
			Status:       domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m1); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		m2 := domain.Module{
			CourseID:    c1.ID,
			Title:        "Module 2",
			Description:  "Module 2",
			Order:        1,
			Status:       domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m2); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		m3 := domain.Module{
			CourseID:    c2.ID,
			Title:        "Module 3",
			Description:  "Module 3",
			Order:        1,
			Status:       domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m3); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		list, err := modules.ListByCourseID(ctx, c1.ID)
		if err != nil {
			t.Fatalf("modules.ListByCourseID failed: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("modules.ListByCourseID: expected 2 modules, got %d", len(list))
		}
		if list[0].ID != m2.ID || list[1].ID != m1.ID {
			t.Fatalf("modules.ListByCourseID: modules not in correct order")
		}
	})

	t.Run("ListByCourseIDEmpty", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)

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

		list, err := modules.ListByCourseID(ctx, c.ID)
		if err != nil {
			t.Fatalf("modules.ListByCourseID failed: %v", err)
		}
		if list == nil {
			t.Fatalf("modules.ListByCourseID: should return empty slice, not nil")
		}
		if len(list) != 0 {
			t.Fatalf("modules.ListByCourseID: expected empty slice, got %d items", len(list))
		}
	})

	t.Run("Update", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)

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
			Title:        "Original Title",
			Description:  "Original Description",
			Order:        1,
			Status:       domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		id := m.ID
		originalUpdatedAt := m.UpdatedAt

		v, ok := modules.GetByID(ctx, id)
		if !ok {
			t.Fatalf("modules.GetByID failed")
		}

		v.Title = "Updated Title"
		v.Description = "Updated Description"
		v.Order = 2
		v.Status = domain.ModuleStatusPublished
		if m.Title != "Original Title" {
			t.Fatalf("ModuleRepository: external modification affected persisted value")
		}

		if err := modules.Update(ctx, v); err != nil {
			t.Fatalf("modules.Update failed: %v", err)
		}

		w, ok := modules.GetByID(ctx, id)
		if !ok {
			t.Fatalf("modules.GetByID failed")
		}
		if w.Title != "Updated Title" {
			t.Fatalf("modules.Update: title not updated")
		}
		if w.Description != "Updated Description" {
			t.Fatalf("modules.Update: description not updated")
		}
		if w.Order != 2 {
			t.Fatalf("modules.Update: order not updated")
		}
		if w.Status != domain.ModuleStatusPublished {
			t.Fatalf("modules.Update: status not updated")
		}
		if !w.UpdatedAt.After(originalUpdatedAt) {
			t.Fatalf("modules.Update: UpdatedAt not updated")
		}
	})

	t.Run("DeleteByID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)
		modules := newModuleRepo(t)

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
			Title:        "Test Module",
			Description:  "A test module",
			Order:        1,
			Status:       domain.ModuleStatusDraft,
		}
		if err := modules.Create(ctx, &m); err != nil {
			t.Fatalf("modules.Create failed: %v", err)
		}

		if err := modules.DeleteByID(ctx, m.ID); err != nil {
			t.Fatalf("modules.DeleteByID failed: %v", err)
		}

		_, ok := modules.GetByID(ctx, m.ID)
		if ok {
			t.Fatalf("modules.DeleteByID: module still exists after deletion")
		}
	})
}
