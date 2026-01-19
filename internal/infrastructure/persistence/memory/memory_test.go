package memory

import (
	"testing"

	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/infrastructure/persistence/test"
)

func TestUserRepository(t *testing.T) {
    test.TestUserRepository(t, func(t *testing.T) persistence.UserRepository {
        return NewUserRepository()
    })
}

func TestProposalRepository(t *testing.T) {
	test.TestProposalRepository(t, func(t *testing.T) persistence.ProposalRepository {
		return NewProposalRepository()
	}, func(t *testing.T) persistence.UserRepository {
		return NewUserRepository()
	})
}

func TestCourseRepository(t *testing.T) {
	test.TestCourseRepository(t, func(t *testing.T) persistence.CourseRepository {
		return NewCourseRepository()
	}, func(t *testing.T) persistence.UserRepository {
		return NewUserRepository()
	}, func(t *testing.T) persistence.ProposalRepository {
		return NewProposalRepository()
	})
}

func TestPasswordResetRepository(t *testing.T) {
	test.TestPasswordResetRepository(t, func(t *testing.T) persistence.PasswordResetRepository {
		return NewPasswordResetRepository()
	}, func(t *testing.T) persistence.UserRepository {
		return NewUserRepository()
	})
}
