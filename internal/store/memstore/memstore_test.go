package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"testing"
)

func TestUserStore_InsertAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	u := domain.User{
		Email:        "user@example.com",
		PasswordHash: make([]byte, 20),
	}
	if err := store.InsertUser(ctx, &u); err != nil {
		t.Fatalf("InsertUser failed: %v", err)
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
}

func TestUserStore_CallerModification(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	u := domain.User{
		Email:        "user@example.com",
		PasswordHash: make([]byte, 20),
	}
	if err := store.InsertUser(ctx, &u); err != nil {
		t.Fatalf("InsertUser failed: %v", err)
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
	if err := store.InsertUser(ctx, &u); err != nil {
		t.Fatalf("InsertUser failed: %v", err)
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

	if err := store.UpdateUser(ctx, &v); err != nil {
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

func TestProposalStore_InsertAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewProposalStore()

	p := domain.Proposal{
		Title:    "Title",
		Summary:  "Summary",
		AuthorID: 1,
	}
	if err := store.InsertProposal(ctx, &p); err != nil {
		t.Fatalf("InsertProposal failed: %v", err)
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
	if err := store.InsertProposal(ctx, &p); err != nil {
		t.Fatalf("InsertProposal failed: %v", err)
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
	if err := store.InsertProposal(ctx, &p); err != nil {
		t.Fatalf("InsertProposal failed: %v", err)
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
}
