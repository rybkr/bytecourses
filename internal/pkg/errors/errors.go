package errors

import (
	"errors"
)

var (
	ErrNotFound                = errors.New("not found")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrConflict                = errors.New("conflict")
	ErrInvalidInput            = errors.New("invalid input")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidToken            = errors.New("invalid or expired token")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrInvalidLogin            = errors.New("invalid login")
)
