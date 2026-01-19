package test

import (
	"context"
	"testing"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
)

type NewUserRepository func(t *testing.T) persistence.UserRepository

func TestUserRepository(t *testing.T, newUserRepo NewUserRepository) {
	t.Helper()

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		v, ok := users.GetByID(ctx, u.ID)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}
		if v.ID != u.ID || v.Email != u.Email {
			t.Fatalf("users.GetByID: users u and v differ")
		}
	})

	t.Run("GetByEmail", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		v, ok := users.GetByEmail(ctx, "user@example.com")
		if !ok {
			t.Fatalf("users.GetByEmail failed")
		}
		if v.ID != u.ID || v.Email != u.Email {
			t.Fatalf("users.GetByEmail: users u and v differ")
		}
	})

	t.Run("CreateDuplicate", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}
		if err := users.Create(ctx, &u); err == nil {
			t.Fatalf("users.Create allowed duplicate email insert")
		}

		v, ok := users.GetByID(ctx, u.ID)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}
		if v.ID != u.ID || v.Email != u.Email {
			t.Fatalf("users.GetByID: users u and v differ")
		}
	})

	t.Run("CallerModification", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		v, ok := users.GetByID(ctx, u.ID)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}

		v.ID++
		v.Email = "junk@crap.com"
		if u.ID != v.ID-1 || u.Email != "user@example.com" {
			t.Fatalf("UserRepository: external modification affected persisted value")
		}

		w, ok := users.GetByID(ctx, u.ID)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}
		if w.ID != v.ID-1 || w.Email != "user@example.com" {
			t.Fatalf("UserRepository: external modification affected persisted value")
		}
	})

	t.Run("Update", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		id := u.ID
		v, ok := users.GetByID(ctx, id)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}

		v.Email = "new.email@example.com"
		if u.Email != "user@example.com" {
			t.Fatalf("UserRepository: external modification affected persisted value")
		}

		if err := users.Update(ctx, v); err != nil {
			t.Fatalf("users.Update failed: %v", err)
		}

		w, ok := users.GetByID(ctx, id)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}
		if w.Email != "new.email@example.com" {
			t.Fatalf("users.Update failed: user email not updated")
		}
	})

	t.Run("GetByIDNotFound", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		_, ok := users.GetByID(ctx, -1)
		if ok {
			t.Fatalf("users.GetByID: should return false for non-existent user")
		}
	})

	t.Run("GetByEmailNotFound", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		_, ok := users.GetByEmail(ctx, "nonexistent@example.com")
		if ok {
			t.Fatalf("users.GetByEmail: should return false for non-existent email")
		}
	})

	t.Run("UpdateNotFound", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "nonexistent@example.com",
			PasswordHash: make([]byte, 20),
			Role:         domain.UserRoleStudent,
		}

		err := users.Update(ctx, &u)
		if err == nil {
			t.Fatalf("users.Update: should return error for non-existent user")
		}
		if err != errors.ErrNotFound {
			t.Fatalf("users.Update: expected ErrNotFound, got %v", err)
		}
	})

	t.Run("UpdateEmailConflict", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u1 := domain.User{
			Email:        "user1@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u1); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		u2 := domain.User{
			Email:        "user2@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u2); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		v, ok := users.GetByID(ctx, u2.ID)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}

		v.Email = "user1@example.com"
		if err := users.Update(ctx, v); err == nil {
			t.Fatalf("users.Update: should fail when updating email to existing email")
		}
	})

	t.Run("UpdatePasswordHash", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		id := u.ID
		v, ok := users.GetByID(ctx, id)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}

		newHash := make([]byte, 32)
		for i := range newHash {
			newHash[i] = byte(i)
		}
		v.PasswordHash = newHash
		if u.PasswordHash[0] != 0 {
			t.Fatalf("UserRepository: external modification affected persisted value")
		}

		if err := users.Update(ctx, v); err != nil {
			t.Fatalf("users.Update failed: %v", err)
		}

		w, ok := users.GetByID(ctx, id)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}
		if len(w.PasswordHash) != len(newHash) {
			t.Fatalf("users.Update: password hash length not updated")
		}
		for i := range newHash {
			if w.PasswordHash[i] != newHash[i] {
				t.Fatalf("users.Update: password hash not updated correctly")
			}
		}
	})

	t.Run("UpdateRole", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
			Role:         domain.UserRoleStudent,
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		id := u.ID
		v, ok := users.GetByID(ctx, id)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}

		v.Role = domain.UserRoleInstructor
		if u.Role != domain.UserRoleStudent {
			t.Fatalf("UserRepository: external modification affected persisted value")
		}

		if err := users.Update(ctx, v); err != nil {
			t.Fatalf("users.Update failed: %v", err)
		}

		w, ok := users.GetByID(ctx, id)
		if !ok {
			t.Fatalf("users.GetByID failed")
		}
		if w.Role != domain.UserRoleInstructor {
			t.Fatalf("users.Update: role not updated")
		}
	})
}
