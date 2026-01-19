package test

import (
	"context"
	"testing"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

type NewPasswordResetRepository func(t *testing.T) persistence.PasswordResetRepository

func TestPasswordResetRepository(t *testing.T, newPasswordResetRepo NewPasswordResetRepository, newUserRepo NewUserRepository) {
	t.Helper()

	t.Run("CreateResetToken", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		resets := newPasswordResetRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		tokenHash := []byte("test-token-hash")
		expiresAt := time.Now().Add(1 * time.Hour)

		if err := resets.CreateResetToken(ctx, u.ID, tokenHash, expiresAt); err != nil {
			t.Fatalf("resets.CreateResetToken failed: %v", err)
		}
	})

	t.Run("ConsumeResetToken", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		resets := newPasswordResetRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		tokenHash := []byte("test-token-hash")
		expiresAt := time.Now().Add(1 * time.Hour)

		if err := resets.CreateResetToken(ctx, u.ID, tokenHash, expiresAt); err != nil {
			t.Fatalf("resets.CreateResetToken failed: %v", err)
		}

		now := time.Now()
		consumedUserID, ok := resets.ConsumeResetToken(ctx, tokenHash, now)
		if !ok {
			t.Fatalf("resets.ConsumeResetToken failed")
		}
		if consumedUserID != u.ID {
			t.Fatalf("resets.ConsumeResetToken: expected user ID %d, got %d", u.ID, consumedUserID)
		}
	})

	t.Run("ConsumeResetTokenExpired", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		resets := newPasswordResetRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		tokenHash := []byte("expired-token-hash")
		expiresAt := time.Now().Add(-1 * time.Hour)

		if err := resets.CreateResetToken(ctx, u.ID, tokenHash, expiresAt); err != nil {
			t.Fatalf("resets.CreateResetToken failed: %v", err)
		}

		now := time.Now()
		_, ok := resets.ConsumeResetToken(ctx, tokenHash, now)
		if ok {
			t.Fatalf("resets.ConsumeResetToken: should return false for expired token")
		}
	})

	t.Run("ConsumeResetTokenNotFound", func(t *testing.T) {
		ctx := context.Background()
		resets := newPasswordResetRepo(t)

		tokenHash := []byte("non-existent-token-hash")
		now := time.Now()

		_, ok := resets.ConsumeResetToken(ctx, tokenHash, now)
		if ok {
			t.Fatalf("resets.ConsumeResetToken: should return false for non-existent token")
		}
	})

	t.Run("ConsumeResetTokenTwice", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		resets := newPasswordResetRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		tokenHash := []byte("test-token-hash")
		expiresAt := time.Now().Add(1 * time.Hour)

		if err := resets.CreateResetToken(ctx, u.ID, tokenHash, expiresAt); err != nil {
			t.Fatalf("resets.CreateResetToken failed: %v", err)
		}

		now := time.Now()
		consumedUserID, ok := resets.ConsumeResetToken(ctx, tokenHash, now)
		if !ok {
			t.Fatalf("resets.ConsumeResetToken: first consumption failed")
		}
		if consumedUserID != u.ID {
			t.Fatalf("resets.ConsumeResetToken: expected user ID %d, got %d", u.ID, consumedUserID)
		}

		_, ok = resets.ConsumeResetToken(ctx, tokenHash, now)
		if ok {
			t.Fatalf("resets.ConsumeResetToken: should return false when consuming same token twice")
		}
	})

	t.Run("MultipleTokensSameUser", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		resets := newPasswordResetRepo(t)

		u := domain.User{
			Email:        "user@example.com",
			PasswordHash: make([]byte, 20),
		}
		if err := users.Create(ctx, &u); err != nil {
			t.Fatalf("users.Create failed: %v", err)
		}

		tokenHash1 := []byte("token-hash-1")
		tokenHash2 := []byte("token-hash-2")
		expiresAt := time.Now().Add(1 * time.Hour)

		if err := resets.CreateResetToken(ctx, u.ID, tokenHash1, expiresAt); err != nil {
			t.Fatalf("resets.CreateResetToken failed: %v", err)
		}

		if err := resets.CreateResetToken(ctx, u.ID, tokenHash2, expiresAt); err != nil {
			t.Fatalf("resets.CreateResetToken failed: %v", err)
		}

		now := time.Now()

		consumedUserID1, ok := resets.ConsumeResetToken(ctx, tokenHash1, now)
		if !ok {
			t.Fatalf("resets.ConsumeResetToken: first token consumption failed")
		}
		if consumedUserID1 != u.ID {
			t.Fatalf("resets.ConsumeResetToken: expected user ID %d, got %d", u.ID, consumedUserID1)
		}

		consumedUserID2, ok := resets.ConsumeResetToken(ctx, tokenHash2, now)
		if !ok {
			t.Fatalf("resets.ConsumeResetToken: second token consumption failed")
		}
		if consumedUserID2 != u.ID {
			t.Fatalf("resets.ConsumeResetToken: expected user ID %d, got %d", u.ID, consumedUserID2)
		}
	})

	t.Run("MultipleTokensDifferentUsers", func(t *testing.T) {
		ctx := context.Background()
		users := newUserRepo(t)
		resets := newPasswordResetRepo(t)

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

		tokenHash1 := []byte("user1-token-hash")
		tokenHash2 := []byte("user2-token-hash")
		expiresAt := time.Now().Add(1 * time.Hour)

		if err := resets.CreateResetToken(ctx, u1.ID, tokenHash1, expiresAt); err != nil {
			t.Fatalf("resets.CreateResetToken failed: %v", err)
		}

		if err := resets.CreateResetToken(ctx, u2.ID, tokenHash2, expiresAt); err != nil {
			t.Fatalf("resets.CreateResetToken failed: %v", err)
		}

		now := time.Now()

		consumedUserID1, ok := resets.ConsumeResetToken(ctx, tokenHash1, now)
		if !ok {
			t.Fatalf("resets.ConsumeResetToken: user1 token consumption failed")
		}
		if consumedUserID1 != u1.ID {
			t.Fatalf("resets.ConsumeResetToken: expected user ID %d, got %d", u1.ID, consumedUserID1)
		}

		consumedUserID2, ok := resets.ConsumeResetToken(ctx, tokenHash2, now)
		if !ok {
			t.Fatalf("resets.ConsumeResetToken: user2 token consumption failed")
		}
		if consumedUserID2 != u2.ID {
			t.Fatalf("resets.ConsumeResetToken: expected user ID %d, got %d", u2.ID, consumedUserID2)
		}

		if consumedUserID1 == consumedUserID2 {
			t.Fatalf("resets.ConsumeResetToken: tokens from different users should be isolated")
		}
	})
}
