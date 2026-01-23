package postgres

import (
	"context"
	"os"
	"sync"
	"testing"

	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/infrastructure/persistence/test"
)

var (
	dbCache    = make(map[string]*DB)
	resetCache = make(map[string]bool)
	dbMutex    sync.Mutex
)

func TestUserRepository(t *testing.T) {
	test.TestUserRepository(t, func(t *testing.T) persistence.UserRepository {
		db := getOrOpenTestDB(t)
		return NewUserRepository(db)
	})
}

func TestProposalRepository(t *testing.T) {
	test.TestProposalRepository(t, func(t *testing.T) persistence.ProposalRepository {
		db := getOrOpenTestDB(t)
		return NewProposalRepository(db)
	}, func(t *testing.T) persistence.UserRepository {
		db := getOrOpenTestDB(t)
		return NewUserRepository(db)
	})
}

func TestCourseRepository(t *testing.T) {
	test.TestCourseRepository(t, func(t *testing.T) persistence.CourseRepository {
		db := getOrOpenTestDB(t)
		return NewCourseRepository(db)
	}, func(t *testing.T) persistence.UserRepository {
		db := getOrOpenTestDB(t)
		return NewUserRepository(db)
	}, func(t *testing.T) persistence.ProposalRepository {
		db := getOrOpenTestDB(t)
		return NewProposalRepository(db)
	})
}

func TestPasswordResetRepository(t *testing.T) {
	test.TestPasswordResetRepository(t, func(t *testing.T) persistence.PasswordResetRepository {
		db := getOrOpenTestDB(t)
		return NewPasswordResetRepository(db)
	}, func(t *testing.T) persistence.UserRepository {
		db := getOrOpenTestDB(t)
		return NewUserRepository(db)
	})
}

func TestModuleRepository(t *testing.T) {
	test.TestModuleRepository(t, func(t *testing.T) persistence.ModuleRepository {
		db := getOrOpenTestDB(t)
		return NewModuleRepository(db)
	}, func(t *testing.T) persistence.CourseRepository {
		db := getOrOpenTestDB(t)
		return NewCourseRepository(db)
	}, func(t *testing.T) persistence.UserRepository {
		db := getOrOpenTestDB(t)
		return NewUserRepository(db)
	})
}

func TestReadingRepository(t *testing.T) {
	test.TestReadingRepository(t, func(t *testing.T) persistence.ReadingRepository {
		db := getOrOpenTestDB(t)
		return NewReadingRepository(db)
	}, func(t *testing.T) persistence.ModuleRepository {
		db := getOrOpenTestDB(t)
		return NewModuleRepository(db)
	}, func(t *testing.T) persistence.CourseRepository {
		db := getOrOpenTestDB(t)
		return NewCourseRepository(db)
	}, func(t *testing.T) persistence.UserRepository {
		db := getOrOpenTestDB(t)
		return NewUserRepository(db)
	})
}

func getOrOpenTestDB(t *testing.T) *DB {
	t.Helper()

	testName := t.Name()
	parentTestName := testName
	for i := 0; i < len(testName); i++ {
		if testName[i] == '/' {
			parentTestName = testName[:i]
			break
		}
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	db, ok := dbCache[parentTestName]
	if !ok {
		db = openTestDB(t)
		dbCache[parentTestName] = db

		t.Cleanup(func() {
			dbMutex.Lock()
			delete(dbCache, parentTestName)
			for key := range resetCache {
				if len(key) > len(parentTestName) && key[:len(parentTestName)+1] == parentTestName+"/" {
					delete(resetCache, key)
				}
			}
			dbMutex.Unlock()
		})
	}

	if !resetCache[testName] {
		resetTestDB(t, db)
		resetCache[testName] = true
	}

	return db
}

func openTestDB(t *testing.T) *DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	s, err := Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open sql store: %v", err)
	}

	t.Cleanup(func() { _ = s.Close() })
	return s
}

func resetTestDB(t *testing.T, db *DB) {
	t.Helper()

	_, err := db.db.ExecContext(context.Background(), `
		TRUNCATE TABLE readings RESTART IDENTITY CASCADE;
		TRUNCATE TABLE content_items RESTART IDENTITY CASCADE;
		TRUNCATE TABLE modules RESTART IDENTITY CASCADE;
		TRUNCATE TABLE password_reset_tokens RESTART IDENTITY CASCADE;
		TRUNCATE TABLE courses RESTART IDENTITY CASCADE;
		TRUNCATE TABLE proposals RESTART IDENTITY CASCADE;
		TRUNCATE TABLE users RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		t.Fatalf("reset db: %v", err)
	}
}
