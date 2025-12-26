package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const bcryptCostFactor int = 10

func HashPassword(password string) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCostFactor)
	return bytes, err
}

func VerifyPassword(hash []byte, password string) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}
