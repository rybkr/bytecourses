package auth

import (
	"golang.org/x/crypto/bcrypt"
)

var bcryptCostFactor = bcrypt.DefaultCost

func SetBcryptCost(cost int) {
	if cost >= bcrypt.MinCost && cost <= bcrypt.MaxCost {
		bcryptCostFactor = cost
	}
}

func HashPassword(password string) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCostFactor)
	return bytes, err
}

func VerifyPassword(hash []byte, password string) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}
