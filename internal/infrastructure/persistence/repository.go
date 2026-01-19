package persistence

import (
	"context"
	"time"

	"bytecourses/internal/domain"
)

type UserRepository interface {
	Create(context.Context, *domain.User) error
	GetByID(context.Context, int64) (*domain.User, bool)
	GetByEmail(context.Context, string) (*domain.User, bool)
	Update(context.Context, *domain.User) error
}

type ProposalRepository interface {
	Create(context.Context, *domain.Proposal) error
	GetByID(context.Context, int64) (*domain.Proposal, bool)
	ListByAuthorID(context.Context, int64) ([]domain.Proposal, error)
	ListAllSubmitted(context.Context) ([]domain.Proposal, error)
	Update(context.Context, *domain.Proposal) error
	DeleteByID(context.Context, int64) error
}

type CourseRepository interface {
	Create(ctx context.Context, c *domain.Course) error
	GetByID(ctx context.Context, id int64) (*domain.Course, bool)
	GetByProposalID(ctx context.Context, proposalID int64) (*domain.Course, bool)
	ListAllLive(ctx context.Context) ([]domain.Course, error)
	Update(ctx context.Context, c *domain.Course) error
}

type PasswordResetRepository interface {
	CreateResetToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error
	ConsumeResetToken(ctx context.Context, tokenHash []byte, now time.Time) (userID int64, ok bool)
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
