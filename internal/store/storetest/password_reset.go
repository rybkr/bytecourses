package storetest

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"crypto/sha256"
	"testing"
	"time"
)

func TestPasswordResetStore(t *testing.T, newStore func(t *testing.T) (store.UserStore, store.PasswordResetStore)) {
	t.Helper()

	newActor := func(ctx context.Context, t *testing.T, users store.UserStore, email string) domain.User {
		t.Helper()
		u := domain.User{
			Name:         "User",
			Email:        email,
			PasswordHash: []byte("x"),
			Role:         domain.UserRoleStudent,
		}
		if err := users.CreateUser(ctx, &u); err != nil {
			t.Fatalf("CreateUser (seed author) failed: %v", err)
		}
		return u
	}

	t.Run("CreateAndConsume", func(t *testing.T) {
		ctx := context.Background()
		users, s := newStore(t)
		actor := newActor(ctx, t, users, "u1@example.com")

		hash := sha256.Sum256([]byte("token-1"))
		expires := time.Now().Add(30 * time.Minute)

		if err := s.CreateResetToken(ctx, actor.ID, hash[:], expires); err != nil {
			t.Fatalf("CreateResetToken: %v", err)
		}

		gotUID, ok := s.ConsumeResetToken(ctx, hash[:], time.Now())
		if !ok || gotUID != actor.ID {
			t.Fatalf("ConsumeResetToken = (%d,%v), want (%d,true)", gotUID, ok, actor.ID)
		}

		_, ok = s.ConsumeResetToken(ctx, hash[:], time.Now())
		if ok {
			t.Fatalf("expected token to be unusable after first consume")
		}
	})

	t.Run("UnknownToken", func(t *testing.T) {
		ctx := context.Background()
		_, s := newStore(t)

		hash := sha256.Sum256([]byte("nope"))
		_, ok := s.ConsumeResetToken(ctx, hash[:], time.Unix(1_700_000_000, 0))
		if ok {
			t.Fatalf("expected ok=false for unknown token")
		}
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		ctx := context.Background()
		users, s := newStore(t)
		actor := newActor(ctx, t, users, "u1@example.com")

		hash := sha256.Sum256([]byte("token-exp"))
		expires := time.Now().Add(-time.Minute)

		if err := s.CreateResetToken(ctx, actor.ID, hash[:], expires); err != nil {
			t.Fatalf("CreateResetToken: %v", err)
		}

		_, ok := s.ConsumeResetToken(ctx, hash[:], expires)
		if ok {
			t.Fatalf("expected ok=false for expired token")
		}
	})
}
