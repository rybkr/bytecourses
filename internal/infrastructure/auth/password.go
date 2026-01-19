package auth

import (
	"golang.org/x/crypto/bcrypt"
)

var bcryptCost = bcrypt.DefaultCost

func SetBcryptCost(cost int) {
	bcryptCost = cost
}

func GetBcryptCost() int {
	return bcryptCost
}

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
}

func HashPasswordWithCost(password string, cost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), cost)
}

func CheckPassword(hash []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}
