package storetest

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"testing"
)

type NewStoresModule func(t *testing.T) (store.UserStore, store.CourseStore, store.ModuleStore)

func TestModuleStore(t *testing.T, newStores NewStoresModule) {
	t.Helper()

	newInstructor := func(ctx context.Context, t *testing.T, users store.UserStore, email string) domain.User {
		t.Helper()
		u := domain.User{
			Name:         "Instructor",
			Email:        email,
			PasswordHash: []byte("x"),
			Role:         domain.UserRoleInstructor,
		}
		if err := users.CreateUser(ctx, &u); err != nil {
			t.Fatalf("CreateUser (seed instructor) failed: %v", err)
		}
		return u
	}

	newCourse := func(ctx context.Context, t *testing.T, courses store.CourseStore, instructorID int64, title string) domain.Course {
		t.Helper()
		c := domain.Course{
			Title:        title,
			Summary:      "Summary",
			InstructorID: instructorID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.CreateCourse(ctx, &c); err != nil {
			t.Fatalf("CreateCourse (seed course) failed: %v", err)
		}
		return c
	}

	t.Run("CreateAndGet", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i1@example.com")
		course := newCourse(ctx, t, courses, instructor.ID, "Course 1")

		m := domain.Module{
			CourseID: course.ID,
			Title:    "Module 1",
		}
		if err := modules.CreateModule(ctx, &m); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}
		if m.ID == 0 {
			t.Fatalf("CreateModule did not assign ID")
		}
		if m.Position != 1 {
			t.Fatalf("CreateModule did not assign position: got %d, want 1", m.Position)
		}

		q, ok := modules.GetModuleByID(ctx, m.ID)
		if !ok {
			t.Fatalf("GetModuleByID failed")
		}
		if q.ID != m.ID {
			t.Fatalf("GetModuleByID: modules m and q differ")
		}
		if q.Title != m.Title || q.CourseID != m.CourseID {
			t.Fatalf("GetModuleByID: module fields don't match")
		}
	})

	t.Run("CallerModification", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i2@example.com")
		course := newCourse(ctx, t, courses, instructor.ID, "Course 2")

		m := domain.Module{
			CourseID: course.ID,
			Title:    "Module 1",
		}
		if err := modules.CreateModule(ctx, &m); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}

		q, ok := modules.GetModuleByID(ctx, m.ID)
		if !ok {
			t.Fatalf("GetModuleByID failed")
		}

		q.ID++
		if m.ID != q.ID-1 {
			t.Fatalf("ModuleStore: external pointer modification affected original value")
		}

		r, ok := modules.GetModuleByID(ctx, m.ID)
		if !ok {
			t.Fatalf("GetModuleByID failed")
		}
		if r.ID != q.ID-1 {
			t.Fatalf("ModuleStore: external pointer modification affected stored value")
		}
	})

	t.Run("UpdateModule", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i3@example.com")
		course := newCourse(ctx, t, courses, instructor.ID, "Course 3")

		m := domain.Module{
			CourseID: course.ID,
			Title:    "Original Title",
		}
		if err := modules.CreateModule(ctx, &m); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}

		m.Title = "Updated Title"
		if err := modules.UpdateModule(ctx, &m); err != nil {
			t.Fatalf("UpdateModule failed: %v", err)
		}

		q, ok := modules.GetModuleByID(ctx, m.ID)
		if !ok {
			t.Fatalf("GetModuleByID failed")
		}
		if q.Title != "Updated Title" {
			t.Fatalf("ModuleStore: failed to update module title")
		}

		// Update nonexistent module
		nonexistent := domain.Module{Title: "R", CourseID: course.ID}
		if err := modules.UpdateModule(ctx, &nonexistent); err == nil {
			t.Fatalf("ModuleStore: was allowed to update nonexistent module")
		}
	})

	t.Run("DeleteModule", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i4@example.com")
		course := newCourse(ctx, t, courses, instructor.ID, "Course 4")

		m := domain.Module{
			CourseID: course.ID,
			Title:    "Module to Delete",
		}
		if err := modules.CreateModule(ctx, &m); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}

		if err := modules.DeleteModuleByID(ctx, m.ID); err != nil {
			t.Fatalf("DeleteModuleByID failed: %v", err)
		}

		if _, ok := modules.GetModuleByID(ctx, m.ID); ok {
			t.Fatalf("GetModuleByID returned deleted module")
		}

		// Delete nonexistent module
		if err := modules.DeleteModuleByID(ctx, m.ID); err == nil {
			t.Fatalf("ModuleStore: was allowed to delete nonexistent module")
		}
	})

	t.Run("ListModulesByCourseID", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i5@example.com")
		course1 := newCourse(ctx, t, courses, instructor.ID, "Course 5a")
		course2 := newCourse(ctx, t, courses, instructor.ID, "Course 5b")

		m1 := domain.Module{CourseID: course1.ID, Title: "Module 1"}
		m2 := domain.Module{CourseID: course1.ID, Title: "Module 2"}
		m3 := domain.Module{CourseID: course2.ID, Title: "Module 3"}

		if err := modules.CreateModule(ctx, &m1); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}
		if err := modules.CreateModule(ctx, &m2); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}
		if err := modules.CreateModule(ctx, &m3); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}

		list1, err := modules.ListModulesByCourseID(ctx, course1.ID)
		if err != nil {
			t.Fatalf("ListModulesByCourseID failed: %v", err)
		}
		if len(list1) != 2 {
			t.Fatalf("ListModulesByCourseID: expected 2 modules, got %d", len(list1))
		}
		// Should be sorted by position
		if list1[0].Position > list1[1].Position {
			t.Fatalf("ListModulesByCourseID: modules not sorted by position")
		}

		list2, err := modules.ListModulesByCourseID(ctx, course2.ID)
		if err != nil {
			t.Fatalf("ListModulesByCourseID failed: %v", err)
		}
		if len(list2) != 1 || list2[0].Title != "Module 3" {
			t.Fatalf("ListModulesByCourseID: expected 1 module with title 'Module 3'")
		}
	})

	t.Run("AutoIncrementPosition", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i6@example.com")
		course := newCourse(ctx, t, courses, instructor.ID, "Course 6")

		m1 := domain.Module{CourseID: course.ID, Title: "Module 1"}
		m2 := domain.Module{CourseID: course.ID, Title: "Module 2"}
		m3 := domain.Module{CourseID: course.ID, Title: "Module 3"}

		if err := modules.CreateModule(ctx, &m1); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}
		if err := modules.CreateModule(ctx, &m2); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}
		if err := modules.CreateModule(ctx, &m3); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}

		if m1.Position != 1 || m2.Position != 2 || m3.Position != 3 {
			t.Fatalf("ModuleStore: positions not auto-incremented: got %d, %d, %d", m1.Position, m2.Position, m3.Position)
		}
	})

	t.Run("ReorderModules", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i7@example.com")
		course := newCourse(ctx, t, courses, instructor.ID, "Course 7")

		m1 := domain.Module{CourseID: course.ID, Title: "Module 1"}
		m2 := domain.Module{CourseID: course.ID, Title: "Module 2"}
		m3 := domain.Module{CourseID: course.ID, Title: "Module 3"}

		if err := modules.CreateModule(ctx, &m1); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}
		if err := modules.CreateModule(ctx, &m2); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}
		if err := modules.CreateModule(ctx, &m3); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}

		// Reorder: 3, 1, 2
		if err := modules.ReorderModules(ctx, course.ID, []int64{m3.ID, m1.ID, m2.ID}); err != nil {
			t.Fatalf("ReorderModules failed: %v", err)
		}

		list, err := modules.ListModulesByCourseID(ctx, course.ID)
		if err != nil {
			t.Fatalf("ListModulesByCourseID failed: %v", err)
		}
		if len(list) != 3 {
			t.Fatalf("ListModulesByCourseID: expected 3 modules, got %d", len(list))
		}
		if list[0].Title != "Module 3" || list[1].Title != "Module 1" || list[2].Title != "Module 2" {
			t.Fatalf("ReorderModules: modules not in expected order: %s, %s, %s", list[0].Title, list[1].Title, list[2].Title)
		}
		if list[0].Position != 1 || list[1].Position != 2 || list[2].Position != 3 {
			t.Fatalf("ReorderModules: positions not updated correctly: %d, %d, %d", list[0].Position, list[1].Position, list[2].Position)
		}
	})

	t.Run("ReorderModulesRejectsWrongCourse", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i8@example.com")
		course1 := newCourse(ctx, t, courses, instructor.ID, "Course 8a")
		course2 := newCourse(ctx, t, courses, instructor.ID, "Course 8b")

		m1 := domain.Module{CourseID: course1.ID, Title: "Module 1"}
		m2 := domain.Module{CourseID: course2.ID, Title: "Module 2"}

		if err := modules.CreateModule(ctx, &m1); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}
		if err := modules.CreateModule(ctx, &m2); err != nil {
			t.Fatalf("CreateModule failed: %v", err)
		}

		// Try to reorder with a module from a different course
		if err := modules.ReorderModules(ctx, course1.ID, []int64{m1.ID, m2.ID}); err == nil {
			t.Fatalf("ReorderModules: should reject module from different course")
		}
	})

	t.Run("GetNonexistentModule", func(t *testing.T) {
		ctx := context.Background()
		_, _, modules := newStores(t)

		if _, ok := modules.GetModuleByID(ctx, 1); ok {
			t.Fatalf("GetModuleByID returned nonexistent module")
		}
	})

	t.Run("ListModulesEmptyCourse", func(t *testing.T) {
		ctx := context.Background()
		users, courses, modules := newStores(t)

		instructor := newInstructor(ctx, t, users, "i9@example.com")
		course := newCourse(ctx, t, courses, instructor.ID, "Course 9")

		list, err := modules.ListModulesByCourseID(ctx, course.ID)
		if err != nil {
			t.Fatalf("ListModulesByCourseID failed: %v", err)
		}
		if len(list) != 0 {
			t.Fatalf("ListModulesByCourseID: expected 0 modules, got %d", len(list))
		}
	})
}
