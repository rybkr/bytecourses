package storetest

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"testing"
)

type NewUserStore func(t *testing.T) store.UserStore
type NewProposalStore func(t *testing.T) store.ProposalStore
type NewStores func(t *testing.T) (store.UserStore, store.ProposalStore)

func TestUserStore(t *testing.T, newStore NewUserStore) {
	t.Helper()

	t.Run("CreateAndGet", func(t *testing.T) {
		ctx := context.Background()
		s := newStore(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := s.CreateUser(ctx, &u); err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		v, ok := s.GetUserByID(ctx, u.ID)
		if !ok {
			t.Fatalf("GetUserByID failed")
		}
		if v.ID != u.ID || u.Email != v.Email {
			t.Fatalf("GetUserByID: users u and v differ")
		}

		v, ok = s.GetUserByEmail(ctx, u.Email)
		if !ok {
			t.Fatalf("GetUserByEmail failed")
		}
		if v.ID != u.ID || u.Email != v.Email {
			t.Fatalf("GetUserByEmail: users u and v differ")
		}

		w := domain.User{Email: "user@example.com"}
		if err := s.CreateUser(ctx, &w); err == nil {
			t.Fatalf("UserStore: allowed to insert duplicate emails")
		}
	})

	t.Run("CallerModification", func(t *testing.T) {
		ctx := context.Background()
		s := newStore(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := s.CreateUser(ctx, &u); err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		v, ok := s.GetUserByID(ctx, u.ID)
		if !ok {
			t.Fatalf("GetUserByID failed")
		}

		v.ID++
		if u.ID != v.ID-1 {
			t.Fatalf("UserStore: external pointer modification affected original value")
		}

		w, ok := s.GetUserByID(ctx, u.ID)
		if !ok {
			t.Fatalf("GetUserByID failed")
		}
		if w.ID != v.ID-1 {
			t.Fatalf("UserStore: external pointer modification affected stored value")
		}
	})

	t.Run("UpdateUser", func(t *testing.T) {
		ctx := context.Background()
		s := newStore(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := s.CreateUser(ctx, &u); err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		uid := u.ID
		v, ok := s.GetUserByID(ctx, uid)
		if !ok {
			t.Fatalf("GetUserByID failed")
		}

		v.Email = "new.email@example.com"
		if u.Email != "user@example.com" {
			t.Fatalf("UserStore: external pointer modification affected original value")
		}

		if err := s.UpdateUser(ctx, v); err != nil {
			t.Fatalf("UpdateUser failed: %v", err)
		}

		w, ok := s.GetUserByID(ctx, uid)
		if !ok {
			t.Fatalf("GetUserByID failed")
		}
		if w.Email != "new.email@example.com" {
			t.Fatalf("UserStore: failed to update user")
		}
	})

	t.Run("GetNonexistentUser", func(t *testing.T) {
		ctx := context.Background()
		s := newStore(t)

		if _, ok := s.GetUserByID(ctx, 1); ok {
			t.Fatalf("GetUserByID returned nonexistent user")
		}
		if _, ok := s.GetUserByEmail(ctx, "user@example.com"); ok {
			t.Fatalf("GetUserByEmail returned nonexistent user")
		}
	})

	t.Run("UpdateNonexistentUser", func(t *testing.T) {
		ctx := context.Background()
		s := newStore(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		u.Email = "new.email@example.com"
		if err := s.UpdateUser(ctx, &u); err == nil {
			t.Fatal("UpdateUser accepted nonexistent user")
		}
	})
}

func TestProposalStore(t *testing.T, newStores NewStores) {
	t.Helper()

	newAuthor := func(ctx context.Context, t *testing.T, users store.UserStore, email string) domain.User {
		t.Helper()
		u := domain.User{
			Name:         "Author",
			Email:        email,
			PasswordHash: []byte("x"),
			Role:         domain.UserRoleStudent,
		}
		if err := users.CreateUser(ctx, &u); err != nil {
			t.Fatalf("CreateUser (seed author) failed: %v", err)
		}
		return u
	}

	t.Run("CreateAndGet", func(t *testing.T) {
		ctx := context.Background()
		users, proposals := newStores(t)
		author := newAuthor(ctx, t, users, "a1@example.com")

		p := domain.Proposal{
			Title:    "Title",
			Summary:  "Summary",
			AuthorID: author.ID,
		}
		if err := proposals.CreateProposal(ctx, &p); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}

		q, ok := proposals.GetProposalByID(ctx, p.ID)
		if !ok {
			t.Fatalf("GetProposalByID failed")
		}
		if q.ID != p.ID {
			t.Fatalf("GetProposalByID: proposals p and q differ")
		}
	})

	t.Run("CallerModification", func(t *testing.T) {
		ctx := context.Background()
		users, proposals := newStores(t)
		author := newAuthor(ctx, t, users, "a2@example.com")

		p := domain.Proposal{
			Title:    "Title",
			Summary:  "Summary",
			AuthorID: author.ID,
		}
		if err := proposals.CreateProposal(ctx, &p); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}

		q, ok := proposals.GetProposalByID(ctx, p.ID)
		if !ok {
			t.Fatalf("GetProposalByID failed")
		}

		q.ID++
		if p.ID != q.ID-1 {
			t.Fatalf("ProposalStore: external pointer modification affected original value")
		}

		r, ok := proposals.GetProposalByID(ctx, p.ID)
		if !ok {
			t.Fatalf("GetProposalByID failed")
		}
		if r.ID != q.ID-1 {
			t.Fatalf("ProposalStore: external pointer modification affected stored value")
		}
	})

	t.Run("UpdateProposal", func(t *testing.T) {
		ctx := context.Background()
		users, proposals := newStores(t)
		author := newAuthor(ctx, t, users, "a3@example.com")

		p := domain.Proposal{
			Title:    "Title",
			Summary:  "Summary",
			AuthorID: author.ID,
		}
		if err := proposals.CreateProposal(ctx, &p); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}

		q := domain.Proposal{
			Title:    "New Title",
			Summary:  "New summary",
			AuthorID: author.ID,
		}
		q.ID = p.ID
		if err := proposals.UpdateProposal(ctx, &q); err != nil {
			t.Fatalf("UpdateProposal failed: %v", err)
		}

		r, ok := proposals.GetProposalByID(ctx, p.ID)
		if !ok {
			t.Fatalf("GetProposalByID failed")
		}
		if r.Title != "New Title" || r.Summary != "New summary" {
			t.Fatalf("ProposalStore: failed to update proposal")
		}

		s := domain.Proposal{Title: "R", AuthorID: author.ID}
		if err := proposals.UpdateProposal(ctx, &s); err == nil {
			t.Fatalf("ProposalStore: was allowed to update nonexistent proposal")
		}
	})

	t.Run("ListProposalsByAuthorID", func(t *testing.T) {
		ctx := context.Background()
		users, proposals := newStores(t)

		a1 := newAuthor(ctx, t, users, "a4@example.com")
		a2 := newAuthor(ctx, t, users, "a5@example.com")

		p := domain.Proposal{Title: "P", AuthorID: a1.ID}
		q := domain.Proposal{Title: "Q", AuthorID: a2.ID}
		r := domain.Proposal{Title: "R", AuthorID: a1.ID}

		if err := proposals.CreateProposal(ctx, &p); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}
		if err := proposals.CreateProposal(ctx, &q); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}
		if err := proposals.CreateProposal(ctx, &r); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}

		p1, err := proposals.ListProposalsByAuthorID(ctx, a1.ID)
		if err != nil {
			t.Fatalf("ListProposalsByAuthorID failed: %v", err)
		}
		if len(p1) != 2 || (p1[0].Title != "P" && p1[0].Title != "R") || (p1[1].Title != "P" && p1[1].Title != "R") || p1[0].Title == p1[1].Title {
			t.Fatalf("ProposalStore: failed to list proposals by author ID")
		}

		p2, err := proposals.ListProposalsByAuthorID(ctx, a2.ID)
		if err != nil {
			t.Fatalf("ListProposalsByAuthorID failed: %v", err)
		}
		if len(p2) != 1 || p2[0].Title != "Q" {
			t.Fatalf("ProposalStore: failed to list proposals by author ID")
		}
	})

	t.Run("ListAllSubmittedProposals", func(t *testing.T) {
		ctx := context.Background()
		users, proposals := newStores(t)

		a1 := newAuthor(ctx, t, users, "a6@example.com")
		a2 := newAuthor(ctx, t, users, "a7@example.com")

		p := domain.Proposal{Title: "P", AuthorID: a1.ID, Status: "draft"}
		q := domain.Proposal{Title: "Q", AuthorID: a2.ID, Status: "submitted"}

		if err := proposals.CreateProposal(ctx, &p); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}
		if err := proposals.CreateProposal(ctx, &q); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}

		ps, err := proposals.ListAllSubmittedProposals(ctx)
		if err != nil {
			t.Fatalf("ListAllSubmittedProposals failed: %v", err)
		}
		if len(ps) != 1 || ps[0].Title != "Q" {
			t.Fatalf("ProposalStore: failed to list all submitted proposals")
		}
	})

	t.Run("DeleteProposalByID", func(t *testing.T) {
		ctx := context.Background()
		users, proposals := newStores(t)
		author := newAuthor(ctx, t, users, "a8@example.com")

		p := domain.Proposal{Title: "P", AuthorID: author.ID}
		q := domain.Proposal{Title: "Q", AuthorID: author.ID}
		r := domain.Proposal{Title: "R", AuthorID: author.ID}

		if err := proposals.CreateProposal(ctx, &p); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}
		if err := proposals.CreateProposal(ctx, &q); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}
		if err := proposals.CreateProposal(ctx, &r); err != nil {
			t.Fatalf("CreateProposal failed: %v", err)
		}

		ps, err := proposals.ListProposalsByAuthorID(ctx, author.ID)
		if err != nil {
			t.Fatalf("ListProposalsByAuthorID failed: %v", err)
		}
		if len(ps) != 3 {
			t.Fatalf("ProposalStore: failed to list proposals by author ID")
		}

		if err = proposals.DeleteProposalByID(ctx, p.ID); err != nil {
			t.Fatalf("DeleteProposalByID failed: %v", err)
		}
		if err = proposals.DeleteProposalByID(ctx, r.ID); err != nil {
			t.Fatalf("DeleteProposalByID failed: %v", err)
		}

		ps, err = proposals.ListProposalsByAuthorID(ctx, author.ID)
		if err != nil {
			t.Fatalf("ListProposalsByAuthorID failed: %v", err)
		}
		if len(ps) != 1 || ps[0].Title != "Q" {
			t.Fatalf("ProposalStore: failed to delete proposals from store")
		}

		if err = proposals.DeleteProposalByID(ctx, p.ID); err == nil {
			t.Fatalf("ProposalStore: was allowed to delete a nonexistent proposal")
		}
	})

	t.Run("GetNonexistentProposal", func(t *testing.T) {
		ctx := context.Background()
		_, proposals := newStores(t)

		if _, ok := proposals.GetProposalByID(ctx, 1); ok {
			t.Fatalf("GetProposalByID returned nonexistent proposal")
		}
	})
}
