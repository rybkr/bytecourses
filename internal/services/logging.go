package services

import (
	"bytecourses/internal/domain"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func NewLogger() *Logger {
	level := slog.LevelInfo
	levelStr := os.Getenv("LOG_LEVEL")

	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	}

	format := os.Getenv("LOG_FORMAT")
	if format == "json" {
		return &Logger{
			Logger: slog.New(slog.NewJSONHandler(os.Stdout, opts)),
		}
	} else {
		return &Logger{
			Logger: slog.New(slog.NewTextHandler(os.Stderr, opts)),
		}
	}
}

func sanitizeEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	if len(parts[0]) > 0 {
		return string(parts[0][0]) + "***@" + parts[1]
	}
	return "***@" + parts[1]
}

type AuthLogger struct {
	base *slog.Logger
}

func NewAuthLogger(base *slog.Logger) *AuthLogger {
	return &AuthLogger{
		base: base.With("service", "auth_service"),
	}
}

func (l *AuthLogger) InfoUser(event string, user *domain.User, fields ...any) {
	args := []any{
		"event", event,
		"user_id", user.ID,
		"email", sanitizeEmail(user.Email),
		"role", user.Role,
	}
	args = append(args, fields...)
	l.base.Info("auth event", args...)
}

func (l *AuthLogger) ErrorOp(operation string, email string, err error) {
	l.base.Error("auth operation failed",
		"event", "auth.error",
		"operation", operation,
		"email", sanitizeEmail(email),
		"error", err,
	)
}

func (l *AuthLogger) Info(event string, fields ...any) {
	args := []any{"event", event}
	args = append(args, fields...)
	l.base.Info("auth event", args...)
}

func (l *AuthLogger) Error(msg string, fields ...any) {
	l.base.Error(msg, fields...)
}

func (l *AuthLogger) Warn(msg string, fields ...any) {
	l.base.Warn(msg, fields...)
}

type ProposalLogger struct {
	base *slog.Logger
}

func NewProposalLogger(base *slog.Logger) *ProposalLogger {
	return &ProposalLogger{
		base: base.With("service", "proposal_service"),
	}
}

func (l *ProposalLogger) InfoTransition(action string, proposal *domain.Proposal, user *domain.User, oldStatus, newStatus domain.ProposalStatus) {
	l.base.Info("proposal transition",
		"event", "proposal.transition",
		"action", action,
		"proposal_id", proposal.ID,
		"user_id", user.ID,
		"old_status", oldStatus,
		"new_status", newStatus,
		"title", proposal.Title,
	)
}

func (l *ProposalLogger) InfoOp(operation string, proposal *domain.Proposal, user *domain.User) {
	l.base.Info("proposal operation",
		"event", "proposal.operation",
		"operation", operation,
		"proposal_id", proposal.ID,
		"user_id", user.ID,
		"status", proposal.Status,
		"title", proposal.Title,
	)
}

func (l *ProposalLogger) ErrorOp(operation string, proposal *domain.Proposal, user *domain.User, err error) {
	l.base.Error("proposal operation failed",
		"event", "proposal.operation",
		"operation", operation,
		"proposal_id", proposal.ID,
		"user_id", user.ID,
		"error", err,
	)
}

func (l *ProposalLogger) Info(event string, fields ...any) {
	args := []any{"event", event}
	args = append(args, fields...)
	l.base.Info("proposal event", args...)
}

func (l *ProposalLogger) Error(msg string, fields ...any) {
	l.base.Error(msg, fields...)
}

type CourseLogger struct {
	base *slog.Logger
}

func NewCourseLogger(base *slog.Logger) *CourseLogger {
	return &CourseLogger{
		base: base.With("service", "course_service"),
	}
}

func (l *CourseLogger) Info(event string, fields ...any) {
	args := []any{"event", event}
	args = append(args, fields...)
	l.base.Info("course event", args...)
}

func (l *CourseLogger) Error(msg string, fields ...any) {
	l.base.Error(msg, fields...)
}
