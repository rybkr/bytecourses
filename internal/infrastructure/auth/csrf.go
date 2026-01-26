package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
)

const (
	CSRFTokenLength    = 32
	CSRFTokenHexLength = CSRFTokenLength * 2
)

func GenerateCSRFToken() (string, error) {
	tokenBytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

func ValidateCSRFToken(token1, token2 string) bool {
	if len(token1) != CSRFTokenHexLength || len(token2) != CSRFTokenHexLength {
		return false
	}

	token1Bytes := []byte(token1)
	token2Bytes := []byte(token2)
	return subtle.ConstantTimeCompare(token1Bytes, token2Bytes) == 1
}
