package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"testing"
)

func TestUserStore_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	u := domain.User{
		Email:        "user@example.com",
		PasswordHash: make([]byte, 20),
	}
	if err := store.CreateUser(ctx, &u); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	v, ok := store.GetUserByID(ctx, u.ID)
	if !ok {
		t.Fatalf("GetUserByID failed")
	}
	if v.ID != u.ID || u.Email != v.Email {
		t.Fatalf("GetUserByID: users u and v differ")
	}

	v, ok = store.GetUserByEmail(ctx, u.Email)
	if !ok {
		t.Fatalf("GetUserByEmail failed")
	}
	if v.ID != u.ID || u.Email != v.Email {
		t.Fatalf("GetUserByEmail: users u and v differ")
	}

	w := domain.User{
		Email: "user@example.com",
	}
    if err := store.CreateUser(ctx, &w); err == nil {
        t.Fatalf("UserStore: allowed to insert duplicate emails")
    }
}

func TestUserStore_CallerModification(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	u := domain.User{
		Email:        "user@example.com",
		PasswordHash: make([]byte, 20),
	}
	if err := store.CreateUser(ctx, &u); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	v, ok := store.GetUserByID(ctx, u.ID)
	if !ok {
		t.Fatalf("GetUserByID failed")
	}

	v.ID++
	if u.ID != v.ID-1 {
		t.Fatalf("UserStore: external pointer modification affected original value")
	}

	w, ok := store.GetUserByID(ctx, u.ID)
	if !ok {
		t.Fatalf("GetUserByID failed")
	}
	if w.ID != v.ID-1 {
		t.Fatalf("UserStore: external pointer modification affected stored value")
	}
}

func TestUserStore_UpdateUser(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	u := domain.User{
		Email:        "user@example.com",
		PasswordHash: make([]byte, 20),
	}
	if err := store.CreateUser(ctx, &u); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	uid := u.ID
	v, ok := store.GetUserByID(ctx, uid)
	if !ok {
		t.Fatalf("GetUserByID failed")
	}

	v.Email = "new.email@example.com"
	if u.Email != "user@example.com" {
		t.Fatalf("UserStore: external pointer modification affected original value")
	}

	if err := store.UpdateUser(ctx, v); err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	w, ok := store.GetUserByID(ctx, uid)
	if !ok {
		t.Fatalf("GetUserByID failed")
	}
	if w.Email != "new.email@example.com" {
		t.Fatalf("UserStore: failed to update user")
	}
}

func TestUserStore_GetNonexistentUser(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	if _, ok := store.GetUserByID(ctx, 1); ok {
		t.Fatalf("GetUserByID retuned nonexistent user")
	}
	if _, ok := store.GetUserByEmail(ctx, "user@example.com"); ok {
		t.Fatalf("GetUserByID retuned nonexistent user")
	}
}

func TestUserStore_UpdateNonexistentUser(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	u := domain.User{
		Email:        "user@example.com",
		PasswordHash: make([]byte, 20),
	}
	u.Email = "new.email@example.com"
	if err := store.UpdateUser(ctx, &u); err == nil {
		t.Fatal("UpdateUser accepted nonexistent user")
	}
}

func TestProposalStore_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewProposalStore()

	p := domain.Proposal{
		Title:    "Title",
		Summary:  "Summary",
		AuthorID: 1,
	}
	if err := store.CreateProposal(ctx, &p); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}

	q, ok := store.GetProposalByID(ctx, p.ID)
	if !ok {
		t.Fatalf("GetProposalByID failed")
	}
	if q.ID != p.ID {
		t.Fatalf("GetUserByID: proposals p and q differ")
	}
}

func TestProposalStore_CallerModification(t *testing.T) {
	ctx := context.Background()
	store := NewProposalStore()

	p := domain.Proposal{
		Title:    "Title",
		Summary:  "Summary",
		AuthorID: 1,
	}
	if err := store.CreateProposal(ctx, &p); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}

	q, ok := store.GetProposalByID(ctx, p.ID)
	if !ok {
		t.Fatalf("GetProposalByID failed")
	}

	q.ID++
	if p.ID != q.ID-1 {
		t.Fatalf("ProposalStore: external pointer modification affected original value")
	}

	r, ok := store.GetProposalByID(ctx, p.ID)
	if !ok {
		t.Fatalf("GetProposalByID failed")
	}
	if r.ID != q.ID-1 {
		t.Fatalf("ProposalStore: external pointer modification affected stored value")
	}
}

func TestProposalStore_UpdateProposal(t *testing.T) {
	ctx := context.Background()
	store := NewProposalStore()

	p := domain.Proposal{
		Title:    "Title",
		Summary:  "Summary",
		AuthorID: 1,
	}
	if err := store.CreateProposal(ctx, &p); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}

	q := domain.Proposal{
		Title:    "New Title",
		Summary:  "New summary",
		AuthorID: 1,
	}
	q.ID = p.ID
	if err := store.UpdateProposal(ctx, &q); err != nil {
		t.Fatalf("UpdateProposal failed: %v", err)
	}

	r, ok := store.GetProposalByID(ctx, p.ID)
	if !ok {
		t.Fatalf("GetProposalByID failed")
	}
	if r.Title != "New Title" || r.Summary != "New summary" {
		t.Fatalf("ProposalStore: failed to update proposal")
	}

	s := domain.Proposal{
		Title:    "R",
	}
    if err := store.UpdateProposal(ctx, &s); err == nil {
        t.Fatalf("ProposalStore: was allowed to update nonexistent proposal")
    }
}

func TestProposalStore_ListProposalsByAuthorID(t *testing.T) {
	ctx := context.Background()
	store := NewProposalStore()

	p := domain.Proposal{
		Title:    "P",
		AuthorID: 1,
	}
	q := domain.Proposal{
		Title:    "Q",
		AuthorID: 2,
	}
	r := domain.Proposal{
		Title:    "R",
		AuthorID: 1,
	}

	if err := store.CreateProposal(ctx, &p); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}
	if err := store.CreateProposal(ctx, &q); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}
	if err := store.CreateProposal(ctx, &r); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}

	p1, err := store.ListProposalsByAuthorID(ctx, 1)
	if err != nil {
		t.Fatalf("ListProposalsByAuthorID failed: %v", err)
	}
	if len(p1) != 2 || (p1[0].Title != "P" && p1[0].Title != "R") || (p1[1].Title != "P" && p1[1].Title != "R") || p1[0].Title == p1[1].Title {
		t.Fatalf("ProposalStore: failed to list proposals by author ID")
	}

	p2, err := store.ListProposalsByAuthorID(ctx, 2)
	if err != nil {
		t.Fatalf("ListProposalsByAuthorID failed: %v", err)
	}
	if len(p2) != 1 || p2[0].Title != "Q" {
		t.Fatalf("ProposalStore: failed to list proposals by author ID")
	}
}

func TestProposalStore_ListAllSubmittedProposals(t *testing.T) {
	ctx := context.Background()
	store := NewProposalStore()

	p := domain.Proposal{
		Title:    "P",
		AuthorID: 1,
		Status:   "draft",
	}
	q := domain.Proposal{
		Title:    "Q",
		AuthorID: 2,
		Status:   "submitted",
	}

	if err := store.CreateProposal(ctx, &p); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}
	if err := store.CreateProposal(ctx, &q); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}

	ps, err := store.ListAllSubmittedProposals(ctx)
	if err != nil {
		t.Fatalf("ListAllSubmittedProposals failed: %v", err)
	}
	if len(ps) != 1 || ps[0].Title != "Q" {
		t.Fatalf("ProposalStore: failed to list all submitted proposals")
	}
}

func TestProposalStore_DeleteProposalByID(t *testing.T) {
	ctx := context.Background()
	store := NewProposalStore()

	p := domain.Proposal{
		Title:    "P",
		AuthorID: 1,
	}
	q := domain.Proposal{
		Title:    "Q",
		AuthorID: 1,
	}
	r := domain.Proposal{
		Title:    "R",
		AuthorID: 1,
	}

	if err := store.CreateProposal(ctx, &p); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}
	if err := store.CreateProposal(ctx, &q); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}
	if err := store.CreateProposal(ctx, &r); err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}

	ps, err := store.ListProposalsByAuthorID(ctx, 1)
	if err != nil {
		t.Fatalf("ListProposalsByAuthorID failed: %v", err)
	}
	if len(ps) != 3 {
		t.Fatalf("ProposalStore: failed to list proposals by author ID")
	}

	if err = store.DeleteProposalByID(ctx, p.ID); err != nil {
		t.Fatalf("DeleteProposalByID failed: %v", err)
	}
	if err = store.DeleteProposalByID(ctx, r.ID); err != nil {
		t.Fatalf("DeleteProposalByID failed: %v", err)
	}

	ps, err = store.ListProposalsByAuthorID(ctx, 1)
	if err != nil {
		t.Fatalf("ListProposalsByAuthorID failed: %v", err)
	}
	if len(ps) != 1 || ps[0].Title != "Q" {
		t.Fatalf("ProposalStore: failed to delete proposals from store")
	}

    if err = store.DeleteProposalByID(ctx, p.ID); err == nil {
        t.Fatalf("ProposalStore: was allowed to delete a nonexistent proposal")
    }
}

func TestProposalStore_GetNonexistentProposal(t *testing.T) {
	ctx := context.Background()
	store := NewProposalStore()

	if _, ok := store.GetProposalByID(ctx, 1); ok {
		t.Fatalf("GetProposalByID retuned nonexistent proposal")
	}
}
