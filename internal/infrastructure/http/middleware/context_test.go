package middleware

import (
	"bytecourses/internal/domain"
	"context"
	"testing"
)

func TestWithUser(t *testing.T) {
	user := &domain.User{
		ID:    1,
		Email: "test@example.com",
	}
	ctx := WithUser(context.Background(), user)

	retrievedUser, ok := UserFromContext(ctx)
	if !ok {
		t.Fatalf("UserFromContext: should find user after WithUser")
	}
	if retrievedUser.ID != user.ID {
		t.Fatalf("UserFromContext: expected user ID %d, got %d", user.ID, retrievedUser.ID)
	}
	if retrievedUser.Email != user.Email {
		t.Fatalf("UserFromContext: expected email %s, got %s", user.Email, retrievedUser.Email)
	}
}

func TestUserFromContext(t *testing.T) {
	user := &domain.User{
		ID:    42,
		Email: "user@example.com",
	}
	ctx := WithUser(context.Background(), user)

	retrievedUser, ok := UserFromContext(ctx)
	if !ok {
		t.Fatalf("UserFromContext: should find user")
	}
	if retrievedUser.ID != user.ID {
		t.Fatalf("UserFromContext: expected user ID %d, got %d", user.ID, retrievedUser.ID)
	}
}

func TestUserFromContext_NotFound(t *testing.T) {
	_, ok := UserFromContext(context.Background())
	if ok {
		t.Fatalf("UserFromContext: should return false when user not in context")
	}
}

func TestWithSession(t *testing.T) {
	sessionID := "session123"
	ctx := WithSession(context.Background(), sessionID)

	retrievedSession, ok := SessionFromContext(ctx)
	if !ok {
		t.Fatalf("SessionFromContext: should find session after WithSession")
	}
	if retrievedSession != sessionID {
		t.Fatalf("SessionFromContext: expected session ID %s, got %s", sessionID, retrievedSession)
	}
}

func TestSessionFromContext(t *testing.T) {
	sessionID := "abc123"
	ctx := WithSession(context.Background(), sessionID)

	retrievedSession, ok := SessionFromContext(ctx)
	if !ok {
		t.Fatalf("SessionFromContext: should find session")
	}
	if retrievedSession != sessionID {
		t.Fatalf("SessionFromContext: expected session ID %s, got %s", sessionID, retrievedSession)
	}
}

func TestSessionFromContext_NotFound(t *testing.T) {
	_, ok := SessionFromContext(context.Background())
	if ok {
		t.Fatalf("SessionFromContext: should return false when session not in context")
	}
}
