package auth

import (
	"testing"
	"time"
)

func TestCreateAndGet(t *testing.T) {
	store := NewInMemorySessionStore(1 * time.Hour)

	userID := int64(42)
	sessionID, err := store.Create(userID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sessionID == "" {
		t.Fatalf("Create: session ID is empty")
	}

	gotUserID, ok := store.Get(sessionID)
	if !ok {
		t.Fatalf("Get: session not found")
	}
	if gotUserID != userID {
		t.Fatalf("Get: expected user ID %d, got %d", userID, gotUserID)
	}
}

func TestGetNonExistent(t *testing.T) {
	store := NewInMemorySessionStore(1 * time.Hour)

	if _, ok := store.Get("nonexistent"); ok {
		t.Fatalf("Get: expected false for non-existent session ID")
	}
}

func TestDelete(t *testing.T) {
	store := NewInMemorySessionStore(1 * time.Hour)

	userID := int64(1)
	sessionID, err := store.Create(userID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := store.Delete(sessionID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, ok := store.Get(sessionID); ok {
		t.Fatalf("Get: session should not exist after deletion")
	}

	if err := store.Delete(sessionID); err != nil {
		t.Fatalf("Delete: expected no error when deleting non-existent session, got %v", err)
	}
}

func TestExpiration(t *testing.T) {
	ttl := 1 * time.Millisecond
	store := NewInMemorySessionStore(ttl)

	userID := int64(7)
	sessionID, err := store.Create(userID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if _, ok := store.Get(sessionID); !ok {
		t.Fatalf("Get: session should be valid immediately after creation")
	}

	time.Sleep(3 * time.Millisecond)

	if _, ok := store.Get(sessionID); ok {
		t.Fatalf("Get: session should be expired")
	}
}
