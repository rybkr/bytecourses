package persistence

import (
	"context"
	"time"

	"bytecourses/internal/domain"
)

type Repository[T any] interface {
	Create(context.Context, *T) error
	GetByID(context.Context, int64) (*T, bool)
	Update(context.Context, *T) error
}

type UserRepository interface {
	Repository[domain.User]
	GetByEmail(context.Context, string) (*domain.User, bool)
	DeleteByID(context.Context, int64) error
}

type ProposalRepository interface {
	Repository[domain.Proposal]
	ListByAuthorID(context.Context, int64) ([]domain.Proposal, error)
	ListAllSubmitted(context.Context) ([]domain.Proposal, error)
	DeleteByID(context.Context, int64) error
}

type CourseRepository interface {
	Repository[domain.Course]
	ListAllLive(ctx context.Context) ([]domain.Course, error)
	GetByProposalID(ctx context.Context, proposalID int64) (*domain.Course, bool)
}

type PasswordResetRepository interface {
	CreateResetToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error
	ConsumeResetToken(ctx context.Context, tokenHash []byte, now time.Time) (userID int64, ok bool)
}

type ModuleRepository interface {
	Repository[domain.Module]
	ListByCourseID(ctx context.Context, courseID int64) ([]domain.Module, error)
	DeleteByID(ctx context.Context, id int64) error
}

type ReadingRepository interface {
	Repository[domain.Reading]
	ListByModuleID(ctx context.Context, moduleID int64) ([]domain.Reading, error)
	DeleteByID(ctx context.Context, id int64) error
}

type DB interface {
	Ping(context.Context) error
	Close() error
	Stats() *DBStats
}

type DBStats struct {
	MaxOpenConnections int   `json:"max_open_connections"`
	OpenConnections    int   `json:"open_connections"`
	InUse              int   `json:"in_use"`
	Idle               int   `json:"idle"`
	WaitCount          int64 `json:"wait_count"`
	WaitDurationMS     int64 `json:"wait_duration_ms"`
	MaxIdleClosed      int64 `json:"max_idle_closed"`
	MaxIdleTimeClosed  int64 `json:"max_idle_time_closed"`
	MaxLifetimeClosed  int64 `json:"max_lifetime_closed"`
}
