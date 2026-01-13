package store

import (
	"bytecourses/internal/domain"
	"context"
	"time"
)

type UserStore interface {
	CreateUser(context.Context, *domain.User) error
	GetUserByID(context.Context, int64) (*domain.User, bool)
	GetUserByEmail(context.Context, string) (*domain.User, bool)
	UpdateUser(context.Context, *domain.User) error
}

type ProposalStore interface {
	CreateProposal(context.Context, *domain.Proposal) error
	GetProposalByID(context.Context, int64) (*domain.Proposal, bool)
	ListProposalsByAuthorID(context.Context, int64) ([]domain.Proposal, error)
	ListAllSubmittedProposals(context.Context) ([]domain.Proposal, error)
	UpdateProposal(context.Context, *domain.Proposal) error
	DeleteProposalByID(context.Context, int64) error
}

type PasswordResetStore interface {
	CreateResetToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error
	ConsumeResetToken(ctx context.Context, tokenHash []byte, now time.Time) (userID int64, ok bool)
}

type CourseStore interface {
	CreateCourse(ctx context.Context, c *domain.Course) error
	GetCourseByID(ctx context.Context, id int64) (*domain.Course, bool)
	ListAllLiveCourses(ctx context.Context) ([]domain.Course, error)
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

type DB interface {
	Ping(context.Context) error
	Stats() *DBStats
}
