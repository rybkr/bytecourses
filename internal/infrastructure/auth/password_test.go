package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndCheckPassword(t *testing.T) {
	originalCost := GetBcryptCost()
	defer SetBcryptCost(originalCost)
	SetBcryptCost(bcrypt.MinCost)

	password := "testpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if len(hash) == 0 {
		t.Fatalf("HashPassword: hash is empty")
	}

	if err := CheckPassword(hash, password); err != nil {
		t.Fatalf("CheckPassword failed for correct password: %v", err)
	}
}

func TestCheckPasswordWrong(t *testing.T) {
	originalCost := GetBcryptCost()
	defer SetBcryptCost(originalCost)
	SetBcryptCost(bcrypt.MinCost)

	password := "correctpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
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
		t.Fatalf("CheckPassword failed for hash produced by HashPasswordWithCost: %v", err)
	}
}

func TestSetBcryptCost(t *testing.T) {
	originalCost := GetBcryptCost()
	defer SetBcryptCost(originalCost)

	newCost := bcrypt.MinCost
	SetBcryptCost(newCost)

	if got := GetBcryptCost(); got != newCost {
		t.Fatalf("GetBcryptCost: expected %d, got %d", newCost, got)
	}
}
