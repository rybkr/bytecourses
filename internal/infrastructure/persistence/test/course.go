package test

import (
	"context"
	"testing"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

type NewCourseRepository func(t *testing.T) persistence.CourseRepository

func TestCourseRepository(t *testing.T, newCourseRepo NewCourseRepository, newUserRepo NewUserRepository, newProposalRepo NewProposalRepository) {
	t.Helper()

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:      "Test Course",
			Summary:    "A test course",
			InstructorID: u.ID,
			Status:     domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}
		if c.ID == 0 {
			t.Fatalf("courses.Create: ID not set")
		}
		if c.CreatedAt.IsZero() {
			t.Fatalf("courses.Create: CreatedAt not set")
		}
		if c.UpdatedAt.IsZero() {
			t.Fatalf("courses.Create: UpdatedAt not set")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

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

		v, ok := courses.GetByID(ctx, c.ID)
		if !ok {
			t.Fatalf("courses.GetByID failed")
		}
		if v.ID != c.ID || v.Title != c.Title || v.InstructorID != c.InstructorID {
			t.Fatalf("courses.GetByID: courses differ")
		}
	})

	t.Run("GetByIDNotFound", func(t *testing.T) {
		ctx := context.Background()
		courses := newCourseRepo(t)

		_, ok := courses.GetByID(ctx, -1)
		if ok {
			t.Fatalf("courses.GetByID: should return false for non-existent course")
		}
	})

	t.Run("GetByProposalID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)
		courses := newCourseRepo(t)

		u := domain.User{
			Email:        "author@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p := domain.Proposal{
			Title:    "Test Proposal",
			Summary:  "A test proposal",
			AuthorID: u.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Test Course",
			Summary:      "A test course",
			InstructorID: u.ID,
			ProposalID:   &p.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		v, ok := courses.GetByProposalID(ctx, p.ID)
		if !ok {
			t.Fatalf("courses.GetByProposalID failed")
		}
		if v.ID != c.ID {
			t.Fatalf("courses.GetByProposalID: wrong course returned")
		}
		if v.ProposalID == nil || *v.ProposalID != p.ID {
			t.Fatalf("courses.GetByProposalID: proposal ID mismatch")
		}
	})

	t.Run("GetByProposalIDNotFound", func(t *testing.T) {
		ctx := context.Background()
		courses := newCourseRepo(t)

		_, ok := courses.GetByProposalID(ctx, -1)
		if ok {
			t.Fatalf("courses.GetByProposalID: should return false for non-existent proposal ID")
		}
	})

	t.Run("GetByProposalIDNil", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

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
			ProposalID:   nil,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		_, ok := courses.GetByProposalID(ctx, 0)
		if ok {
			t.Fatalf("courses.GetByProposalID: should return false when searching for nil proposal ID")
		}
	})

	t.Run("ListAllLive", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

		u1 := domain.User{
			Email:        "instructor1@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u1); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		u2 := domain.User{
			Email:        "instructor2@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u2); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c1 := domain.Course{
			Title:        "Draft Course",
			Summary:      "Draft",
			InstructorID: u1.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c1); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		c2 := domain.Course{
			Title:        "Live Course 1",
			Summary:      "Live",
			InstructorID: u1.ID,
			Status:       domain.CourseStatusLive,
		}
		if err := courses.Create(ctx, &c2); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		c3 := domain.Course{
			Title:        "Live Course 2",
			Summary:      "Live",
			InstructorID: u2.ID,
			Status:       domain.CourseStatusLive,
		}
		if err := courses.Create(ctx, &c3); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		list, err := courses.ListAllLive(ctx)
		if err != nil {
			t.Fatalf("courses.ListAllLive failed: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("courses.ListAllLive: expected 2 courses, got %d", len(list))
		}
		for _, course := range list {
			if course.Status != domain.CourseStatusLive {
				t.Fatalf("courses.ListAllLive: found course with non-live status")
			}
		}
	})

	t.Run("ListAllLiveEmpty", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Draft Course",
			Summary:      "Draft",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		list, err := courses.ListAllLive(ctx)
		if err != nil {
			t.Fatalf("courses.ListAllLive failed: %v", err)
		}
		if list == nil {
			t.Fatalf("courses.ListAllLive: should return empty slice, not nil")
		}
		if len(list) != 0 {
			t.Fatalf("courses.ListAllLive: expected empty slice, got %d items", len(list))
		}
	})

	t.Run("Update", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Original Title",
			Summary:      "Original Summary",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		id := c.ID
		originalUpdatedAt := c.UpdatedAt

		v, ok := courses.GetByID(ctx, id)
		if !ok {
			t.Fatalf("courses.GetByID failed")
		}

		v.Title = "Updated Title"
		v.Summary = "Updated Summary"
		v.Status = domain.CourseStatusLive
		if c.Title != "Original Title" {
			t.Fatalf("CourseRepository: external modification affected persisted value")
		}

		if err := courses.Update(ctx, v); err != nil {
			t.Fatalf("courses.Update failed: %v", err)
		}

		w, ok := courses.GetByID(ctx, id)
		if !ok {
			t.Fatalf("courses.GetByID failed")
		}
		if w.Title != "Updated Title" {
			t.Fatalf("courses.Update: title not updated")
		}
		if w.Summary != "Updated Summary" {
			t.Fatalf("courses.Update: summary not updated")
		}
		if w.Status != domain.CourseStatusLive {
			t.Fatalf("courses.Update: status not updated")
		}
		if !w.UpdatedAt.After(originalUpdatedAt) {
			t.Fatalf("courses.Update: UpdatedAt not updated")
		}
	})

	t.Run("UpdateNotFound", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			ID:           -1,
			Title:        "Non-existent",
			Summary:      "Does not exist",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}

		err := courses.Update(ctx, &c)
		if err != nil {
			t.Fatalf("courses.Update: should handle non-existent course gracefully")
		}
	})

	t.Run("CallerModification", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c := domain.Course{
			Title:        "Test Course",
			Summary:      "Test Summary",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		v, ok := courses.GetByID(ctx, c.ID)
		if !ok {
			t.Fatalf("courses.GetByID failed")
		}

		originalInstructorID := v.InstructorID
		v.ID++
		v.Title = "Modified Title"
		v.InstructorID = -1 
		if c.ID != v.ID-1 || c.Title != "Test Course" || c.InstructorID != originalInstructorID {
			t.Fatalf("CourseRepository: external modification affected persisted value")
		}

		w, ok := courses.GetByID(ctx, c.ID)
		if !ok {
			t.Fatalf("courses.GetByID failed")
		}
		if w.ID != v.ID-1 || w.Title != "Test Course" || w.InstructorID != originalInstructorID {
			t.Fatalf("CourseRepository: external modification affected persisted value")
		}
	})

	t.Run("StatusFiltering", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		courses := newCourseRepo(t)

		u := domain.User{
			Email:        "instructor@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		c1 := domain.Course{
			Title:        "Draft Course 1",
			Summary:      "Draft",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c1); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		c2 := domain.Course{
			Title:        "Draft Course 2",
			Summary:      "Draft",
			InstructorID: u.ID,
			Status:       domain.CourseStatusDraft,
		}
		if err := courses.Create(ctx, &c2); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		c3 := domain.Course{
			Title:        "Live Course",
			Summary:      "Live",
			InstructorID: u.ID,
			Status:       domain.CourseStatusLive,
		}
		if err := courses.Create(ctx, &c3); err != nil {
			t.Fatalf("courses.Create failed: %v", err)
		}

		list, err := courses.ListAllLive(ctx)
		if err != nil {
			t.Fatalf("courses.ListAllLive failed: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("courses.ListAllLive: expected 1 live course, got %d", len(list))
		}
		if list[0].Status != domain.CourseStatusLive {
			t.Fatalf("courses.ListAllLive: returned course is not live")
		}
		if list[0].ID != c3.ID {
			t.Fatalf("courses.ListAllLive: returned wrong course")
		}
	})
}
