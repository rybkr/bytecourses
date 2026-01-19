package auth

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if err := CheckPassword(hash, password); err != nil {
		t.Fatalf("CheckPassword failed for correct password: %v", err)
	}
}

func TestCheckPassword(t *testing.T) {
	password := "correctpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if err := CheckPassword(hash, password); err != nil {
		t.Fatalf("CheckPassword failed for correct password: %v", err)
	}

	if err := CheckPassword(hash, "wrongpassword"); err == nil {
		t.Fatalf("CheckPassword should fail for wrong password")
	}
}

func TestHashPasswordWithCost(t *testing.T) {
	password := "testpassword"
	cost := bcrypt.MinCost

	hash, err := HashPasswordWithCost(password, cost)
	if err != nil {
		t.Fatalf("HashPasswordWithCost failed: %v", err)
	}

	if err := CheckPassword(hash, password); err != nil {
		t.Fatalf("CheckPassword failed: %v", err)
	}
}

func TestSetBcryptCost(t *testing.T) {
	originalCost := GetBcryptCost()
	defer SetBcryptCost(originalCost)

	newCost := bcrypt.MinCost
	SetBcryptCost(newCost)

	if GetBcryptCost() != newCost {
		t.Fatalf("SetBcryptCost/GetBcryptCost: expected %d, got %d", newCost, GetBcryptCost())
	}
}

func TestInMemorySessionStore_Create(t *testing.T) {
	store := NewInMemorySessionStore(1 * time.Hour)

	userID := int64(1)
	sessionID, err := store.Create(userID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sessionID == "" {
		t.Fatalf("Create: session ID is empty")
	}

	retrievedUserID, ok := store.Get(sessionID)
	if !ok {
		t.Fatalf("Get: session not found after creation")
	}
	if retrievedUserID != userID {
		t.Fatalf("Get: expected user ID %d, got %d", userID, retrievedUserID)
	}
}

func TestInMemorySessionStore_Get(t *testing.T) {
	store := NewInMemorySessionStore(1 * time.Hour)

	userID := int64(123)
	sessionID, err := store.Create(userID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrievedUserID, ok := store.Get(sessionID)
	if !ok {
		t.Fatalf("Get: session not found")
	}
	if retrievedUserID != userID {
		t.Fatalf("Get: expected user ID %d, got %d", userID, retrievedUserID)
	}

	_, ok = store.Get("nonexistent")
	if ok {
		t.Fatalf("Get: should return false for non-existent session")
	}
}

func TestInMemorySessionStore_Delete(t *testing.T) {
	store := NewInMemorySessionStore(1 * time.Hour)

	userID := int64(1)
	sessionID, err := store.Create(userID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := store.Delete(sessionID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, ok := store.Get(sessionID)
	if ok {
		t.Fatalf("Get: session should not exist after deletion")
	}
}

func TestInMemorySessionStore_Expiration(t *testing.T) {
	store := NewInMemorySessionStore(50 * time.Millisecond)

	userID := int64(1)
	sessionID, err := store.Create(userID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, ok := store.Get(sessionID)
	if !ok {
		t.Fatalf("Get: session should be valid immediately")
	}

	time.Sleep(100 * time.Millisecond)

	_, ok = store.Get(sessionID)
	if ok {
		t.Fatalf("Get: session should be expired")
	}
}
