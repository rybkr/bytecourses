package services

import (
	"errors"
)

var (
	ErrInvalidInput            = errors.New("invalid input")
	ErrNotFound                = errors.New("not found")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrConflict                = errors.New("conflict")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidToken            = errors.New("invalid or expired token")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)
