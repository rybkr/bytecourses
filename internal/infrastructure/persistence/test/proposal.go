package test

import (
	"context"
	"testing"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

type NewProposalRepository func(t *testing.T) persistence.ProposalRepository

func TestProposalRepository(t *testing.T, newProposalRepo NewProposalRepository, newUserRepo NewUserRepository) {
	t.Helper()

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

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
		if p.ID == 0 {
			t.Fatalf("proposals.Create: ID not set")
		}
		if p.CreatedAt.IsZero() {
			t.Fatalf("proposals.Create: CreatedAt not set")
		}
		if p.UpdatedAt.IsZero() {
			t.Fatalf("proposals.Create: UpdatedAt not set")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

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

		v, ok := proposals.GetByID(ctx, p.ID)
		if !ok {
			t.Fatalf("proposals.GetByID failed")
		}
		if v.ID != p.ID || v.Title != p.Title || v.AuthorID != p.AuthorID {
			t.Fatalf("proposals.GetByID: proposals differ")
		}
	})

	t.Run("GetByIDNotFound", func(t *testing.T) {
		ctx := context.Background()
		proposals := newProposalRepo(t)

		_, ok := proposals.GetByID(ctx, -1)
		if ok {
			t.Fatalf("proposals.GetByID: should return false for non-existent proposal")
		}
	})

	t.Run("ListByAuthorID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

		u1 := domain.User{
			Email:        "author1@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u1); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		u2 := domain.User{
			Email:        "author2@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u2); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p1 := domain.Proposal{
			Title:    "Proposal 1",
			Summary:  "First proposal",
			AuthorID: u1.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p1); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		p2 := domain.Proposal{
			Title:    "Proposal 2",
			Summary:  "Second proposal",
			AuthorID: u1.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p2); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		p3 := domain.Proposal{
			Title:    "Proposal 3",
			Summary:  "Third proposal",
			AuthorID: u2.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p3); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		list, err := proposals.ListByAuthorID(ctx, u1.ID)
		if err != nil {
			t.Fatalf("proposals.ListByAuthorID failed: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("proposals.ListByAuthorID: expected 2 proposals, got %d", len(list))
		}
		for _, prop := range list {
			if prop.AuthorID != u1.ID {
				t.Fatalf("proposals.ListByAuthorID: found proposal with wrong author ID")
			}
		}
	})

	t.Run("ListByAuthorIDEmpty", func(t *testing.T) {
		ctx := context.Background()
		proposals := newProposalRepo(t)

		list, err := proposals.ListByAuthorID(ctx, -1)
		if err != nil {
			t.Fatalf("proposals.ListByAuthorID failed: %v", err)
		}
		if list == nil {
			t.Fatalf("proposals.ListByAuthorID: should return empty slice, not nil")
		}
		if len(list) != 0 {
			t.Fatalf("proposals.ListByAuthorID: expected empty slice, got %d items", len(list))
		}
	})

	t.Run("ListAllSubmitted", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

		u1 := domain.User{
			Email:        "author1@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u1); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		u2 := domain.User{
			Email:        "author2@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u2); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		u3 := domain.User{
			Email:        "author3@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u3); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p1 := domain.Proposal{
			Title:    "Draft Proposal",
			Summary:  "Draft",
			AuthorID: u1.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p1); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		p2 := domain.Proposal{
			Title:    "Submitted Proposal",
			Summary:  "Submitted",
			AuthorID: u1.ID,
			Status:   domain.ProposalStatusSubmitted,
		}
		if err := proposals.Create(ctx, &p2); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		p3 := domain.Proposal{
			Title:    "Approved Proposal",
			Summary:  "Approved",
			AuthorID: u2.ID,
			Status:   domain.ProposalStatusApproved,
		}
		if err := proposals.Create(ctx, &p3); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		p4 := domain.Proposal{
			Title:    "Rejected Proposal",
			Summary:  "Rejected",
			AuthorID: u2.ID,
			Status:   domain.ProposalStatusRejected,
		}
		if err := proposals.Create(ctx, &p4); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		p5 := domain.Proposal{
			Title:    "Changes Requested Proposal",
			Summary:  "Changes",
			AuthorID: u3.ID,
			Status:   domain.ProposalStatusChangesRequested,
		}
		if err := proposals.Create(ctx, &p5); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		p6 := domain.Proposal{
			Title:    "Withdrawn Proposal",
			Summary:  "Withdrawn",
			AuthorID: u3.ID,
			Status:   domain.ProposalStatusWithdrawn,
		}
		if err := proposals.Create(ctx, &p6); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		list, err := proposals.ListAllSubmitted(ctx)
		if err != nil {
			t.Fatalf("proposals.ListAllSubmitted failed: %v", err)
		}
		if len(list) != 4 {
			t.Fatalf("proposals.ListAllSubmitted: expected 4 proposals, got %d", len(list))
		}
		for _, prop := range list {
			if prop.Status != domain.ProposalStatusSubmitted &&
				prop.Status != domain.ProposalStatusApproved &&
				prop.Status != domain.ProposalStatusRejected &&
				prop.Status != domain.ProposalStatusChangesRequested {
				t.Fatalf("proposals.ListAllSubmitted: found proposal with non-submitted status")
			}
		}
	})

	t.Run("ListAllSubmittedEmpty", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

		u := domain.User{
			Email:        "author@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p := domain.Proposal{
			Title:    "Draft Proposal",
			Summary:  "Draft",
			AuthorID: u.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		list, err := proposals.ListAllSubmitted(ctx)
		if err != nil {
			t.Fatalf("proposals.ListAllSubmitted failed: %v", err)
		}
		if list == nil {
			t.Fatalf("proposals.ListAllSubmitted: should return empty slice, not nil")
		}
		if len(list) != 0 {
			t.Fatalf("proposals.ListAllSubmitted: expected empty slice, got %d items", len(list))
		}
	})

	t.Run("Update", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

		u := domain.User{
			Email:        "author@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p := domain.Proposal{
			Title:    "Original Title",
			Summary:  "Original Summary",
			AuthorID: u.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		id := p.ID
		originalUpdatedAt := p.UpdatedAt

		v, ok := proposals.GetByID(ctx, id)
		if !ok {
			t.Fatalf("proposals.GetByID failed")
		}

		v.Title = "Updated Title"
		v.Summary = "Updated Summary"
		v.Status = domain.ProposalStatusSubmitted
		if p.Title != "Original Title" {
			t.Fatalf("ProposalRepository: external modification affected persisted value")
		}

		if err := proposals.Update(ctx, v); err != nil {
			t.Fatalf("proposals.Update failed: %v", err)
		}

		w, ok := proposals.GetByID(ctx, id)
		if !ok {
			t.Fatalf("proposals.GetByID failed")
		}
		if w.Title != "Updated Title" {
			t.Fatalf("proposals.Update: title not updated")
		}
		if w.Summary != "Updated Summary" {
			t.Fatalf("proposals.Update: summary not updated")
		}
		if w.Status != domain.ProposalStatusSubmitted {
			t.Fatalf("proposals.Update: status not updated")
		}
		if !w.UpdatedAt.After(originalUpdatedAt) {
			t.Fatalf("proposals.Update: UpdatedAt not updated")
		}
	})

	t.Run("UpdateNotFound", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

		u := domain.User{
			Email:        "author@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p := domain.Proposal{
			ID:       -1,
			Title:    "Non-existent",
			Summary:  "Does not exist",
			AuthorID: u.ID,
			Status:   domain.ProposalStatusDraft,
		}

		err := proposals.Update(ctx, &p)
		if err != nil {
			t.Fatalf("proposals.Update: should handle non-existent proposal gracefully")
		}
	})

	t.Run("DeleteByID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

		u := domain.User{
			Email:        "author@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p := domain.Proposal{
			Title:    "To Delete",
			Summary:  "Will be deleted",
			AuthorID: u.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		id := p.ID
		if err := proposals.DeleteByID(ctx, id); err != nil {
			t.Fatalf("proposals.DeleteByID failed: %v", err)
		}

		_, ok := proposals.GetByID(ctx, id)
		if ok {
			t.Fatalf("proposals.DeleteByID: proposal still exists after deletion")
		}
	})

	t.Run("DeleteByIDNotFound", func(t *testing.T) {
		ctx := context.Background()
		proposals := newProposalRepo(t)

		err := proposals.DeleteByID(ctx, -1)
		if err != nil {
			t.Fatalf("proposals.DeleteByID: should handle non-existent proposal gracefully")
		}
	})

	t.Run("CallerModification", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

		u := domain.User{
			Email:        "author@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p := domain.Proposal{
			Title:    "Test Proposal",
			Summary:  "Test Summary",
			AuthorID: u.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		v, ok := proposals.GetByID(ctx, p.ID)
		if !ok {
			t.Fatalf("proposals.GetByID failed")
		}

		originalAuthorID := v.AuthorID
		v.ID++
		v.Title = "Modified Title"
		v.AuthorID = -1
		if p.ID != v.ID-1 || p.Title != "Test Proposal" || p.AuthorID != originalAuthorID {
			t.Fatalf("ProposalRepository: external modification affected persisted value")
		}

		w, ok := proposals.GetByID(ctx, p.ID)
		if !ok {
			t.Fatalf("proposals.GetByID failed")
		}
		if w.ID != v.ID-1 || w.Title != "Test Proposal" || w.AuthorID != originalAuthorID {
			t.Fatalf("ProposalRepository: external modification affected persisted value")
		}
	})

	t.Run("MultipleAuthors", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		proposals := newProposalRepo(t)

		u1 := domain.User{
			Email:        "author1@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u1); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		u2 := domain.User{
			Email:        "author2@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u2); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		p1 := domain.Proposal{
			Title:    "Author 1 Proposal",
			Summary:  "By author 1",
			AuthorID: u1.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p1); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		p2 := domain.Proposal{
			Title:    "Author 2 Proposal",
			Summary:  "By author 2",
			AuthorID: u2.ID,
			Status:   domain.ProposalStatusDraft,
		}
		if err := proposals.Create(ctx, &p2); err != nil {
			t.Fatalf("proposals.Create failed: %v", err)
		}

		list1, err := proposals.ListByAuthorID(ctx, u1.ID)
		if err != nil {
			t.Fatalf("proposals.ListByAuthorID failed: %v", err)
		}
		if len(list1) != 1 {
			t.Fatalf("proposals.ListByAuthorID: expected 1 proposal for author 1, got %d", len(list1))
		}

		list2, err := proposals.ListByAuthorID(ctx, u2.ID)
		if err != nil {
			t.Fatalf("proposals.ListByAuthorID failed: %v", err)
		}
		if len(list2) != 1 {
			t.Fatalf("proposals.ListByAuthorID: expected 1 proposal for author 2, got %d", len(list2))
		}

		if list1[0].ID == list2[0].ID {
			t.Fatalf("proposals.ListByAuthorID: proposals from different authors should be isolated")
		}
	})
}
