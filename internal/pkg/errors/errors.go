package errors

import (
	"errors"
	"fmt"
	"net/http"
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

type AppError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(statusCode int, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func WrapAppError(err error, statusCode int, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUnauthorized),
		errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrInvalidLogin):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrInvalidToken):
		return http.StatusBadRequest
	case errors.Is(err, ErrInvalidStatusTransition):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func GetUserMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return "Resource not found"
	case errors.Is(err, ErrUnauthorized):
		return "Unauthorized"
	case errors.Is(err, ErrForbidden):
		return "Forbidden"
	case errors.Is(err, ErrConflict):
		return "Resource already exists"
	case errors.Is(err, ErrInvalidInput):
		return "Invalid request"
	case errors.Is(err, ErrInvalidCredentials):
		return "Invalid credentials"
	case errors.Is(err, ErrInvalidToken):
		return "Invalid or expired token"
	case errors.Is(err, ErrInvalidStatusTransition):
		return "Invalid status transition"
	case errors.Is(err, ErrInvalidLogin):
		return "Invalid email or password"
	default:
		return "An error occurred. Please try again later."
	}
}
