package middleware

import (
	"context"

	"bytecourses/internal/domain"
)

type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionContextKey contextKey = "session"
)

func WithUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(userContextKey).(*domain.User)
	return user, ok
}

func WithSession(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionContextKey, sessionID)
}

func SessionFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionContextKey).(string)
	return sessionID, ok
}
