package storetest

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"testing"
)

type NewStoresCourse func(t *testing.T) (store.UserStore, store.CourseStore)

func TestCourseStore(t *testing.T, newStores NewStoresCourse) {
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

	t.Run("CreateAndGet", func(t *testing.T) {
		ctx := context.Background()
		users, courses := newStores(t)

		instructor := newInstructor(ctx, t, users, "i1@example.com")

		c := domain.Course{
			Title:        "Title",
			Summary:      "Summary",
			InstructorID: instructor.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.CreateCourse(ctx, &c); err != nil {
			t.Fatalf("CreateCourse failed: %v", err)
		}

		q, ok := courses.GetCourseByID(ctx, c.ID)
		if !ok {
			t.Fatalf("GetCourseByID failed")
		}
		if q.ID != c.ID {
			t.Fatalf("GetCourseByID: courses c and q differ")
		}
		if q.Title != c.Title || q.Summary != c.Summary || q.InstructorID != c.InstructorID {
			t.Fatalf("GetCourseByID: course fields don't match")
		}
	})

	t.Run("CallerModification", func(t *testing.T) {
		ctx := context.Background()
		users, courses := newStores(t)
		instructor := newInstructor(ctx, t, users, "i2@example.com")

		c := domain.Course{
			Title:        "Title",
			Summary:      "Summary",
			InstructorID: instructor.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.CreateCourse(ctx, &c); err != nil {
			t.Fatalf("CreateCourse failed: %v", err)
		}

		q, ok := courses.GetCourseByID(ctx, c.ID)
		if !ok {
			t.Fatalf("GetCourseByID failed")
		}

		q.ID++
		if c.ID != q.ID-1 {
			t.Fatalf("CourseStore: external pointer modification affected original value")
		}

		r, ok := courses.GetCourseByID(ctx, c.ID)
		if !ok {
			t.Fatalf("GetCourseByID failed")
		}
		if r.ID != q.ID-1 {
			t.Fatalf("CourseStore: external pointer modification affected stored value")
		}
	})

	t.Run("ListAllLiveCourses", func(t *testing.T) {
		ctx := context.Background()
		users, courses := newStores(t)

		i1 := newInstructor(ctx, t, users, "i3@example.com")
		i2 := newInstructor(ctx, t, users, "i4@example.com")

		c1 := domain.Course{Title: "Draft Course", Summary: "Summary", InstructorID: i1.ID, Status: domain.CourseStatusDraft}
		c2 := domain.Course{Title: "Live Course 1", Summary: "Summary", InstructorID: i1.ID, Status: domain.CourseStatusLive}
		c3 := domain.Course{Title: "Live Course 2", Summary: "Summary", InstructorID: i2.ID, Status: domain.CourseStatusLive}

		if err := courses.CreateCourse(ctx, &c1); err != nil {
			t.Fatalf("CreateCourse failed: %v", err)
		}
		if err := courses.CreateCourse(ctx, &c2); err != nil {
			t.Fatalf("CreateCourse failed: %v", err)
		}
		if err := courses.CreateCourse(ctx, &c3); err != nil {
			t.Fatalf("CreateCourse failed: %v", err)
		}

		liveCourses, err := courses.ListAllLiveCourses(ctx)
		if err != nil {
			t.Fatalf("ListAllLiveCourses failed: %v", err)
		}
		if len(liveCourses) != 2 {
			t.Fatalf("ListAllLiveCourses: expected 2 live courses, got %d", len(liveCourses))
		}
		liveTitles := make(map[string]bool)
		for _, course := range liveCourses {
			if course.Status != domain.CourseStatusLive {
				t.Fatalf("ListAllLiveCourses: returned non-live course: %s", course.Status)
			}
			liveTitles[course.Title] = true
		}
		if !liveTitles["Live Course 1"] || !liveTitles["Live Course 2"] {
			t.Fatalf("ListAllLiveCourses: did not return expected live courses")
		}
	})

	t.Run("GetNonexistentCourse", func(t *testing.T) {
		ctx := context.Background()
		_, courses := newStores(t)

		if _, ok := courses.GetCourseByID(ctx, 1); ok {
			t.Fatalf("GetCourseByID returned nonexistent course")
		}
	})

	t.Run("UpdateCourse", func(t *testing.T) {
		ctx := context.Background()
		users, courses := newStores(t)
		instructor := newInstructor(ctx, t, users, "i5@example.com")

		c := domain.Course{
			Title:        "Title",
			Summary:      "Summary",
			InstructorID: instructor.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.CreateCourse(ctx, &c); err != nil {
			t.Fatalf("CreateCourse failed: %v", err)
		}

		q := domain.Course{
			Title:        "New Title",
			Summary:      "New Summary",
			InstructorID: instructor.ID,
			Status:       domain.CourseStatusDraft,
		}
		q.ID = c.ID
		if err := courses.UpdateCourse(ctx, &q); err != nil {
			t.Fatalf("UpdateCourse failed: %v", err)
		}

		r, ok := courses.GetCourseByID(ctx, c.ID)
		if !ok {
			t.Fatalf("GetCourseByID failed")
		}
		if r.Title != "New Title" || r.Summary != "New Summary" {
			t.Fatalf("CourseStore: failed to update course")
		}

		s := domain.Course{Title: "R", InstructorID: instructor.ID}
		if err := courses.UpdateCourse(ctx, &s); err == nil {
			t.Fatalf("CourseStore: was allowed to update nonexistent course")
		}
	})
}
