package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"testing"
)

func TestUserStore_InsertAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	u := domain.NewUser("user@example.com", make([]byte, 20))
	if err := store.InsertUser(ctx, u); err != nil {
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

	u := domain.NewUser("user@example.com", make([]byte, 20))
	if err := store.InsertUser(ctx, u); err != nil {
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

	u := domain.NewUser("user@example.com", make([]byte, 20))
	if err := store.InsertUser(ctx, u); err != nil {
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

func TestUserStore_UpdateNonexistentUser(t *testing.T) {
	ctx := context.Background()
	store := NewUserStore()

	u := domain.NewUser("user@example.com", make([]byte, 20))
	u.Email = "new.email@example.com"
	if err := store.UpdateUser(ctx, u); err == nil {
		t.Fatal("UpdateUser accepted nonexistent user")
	}
}

func TestProposalStore_InsertAndGet(t *testing.T) {}
