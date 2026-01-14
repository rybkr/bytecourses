package sqlstore

import (
	"bytecourses/internal/store"
	"bytecourses/internal/store/storetest"
	"context"
	"os"
	"testing"
)

func TestUserStore(t *testing.T) {
	storetest.TestUserStore(t, func(t *testing.T) store.UserStore {
		s := openTestStore(t)
		resetTestDB(t, s)
		return s
	})
}

func TestProposalStore(t *testing.T) {
	storetest.TestProposalStore(t, func(t *testing.T) (store.UserStore, store.ProposalStore) {
		s := openTestStore(t)
		resetTestDB(t, s)
		return s, s
	})
}

func TestPasswordResetStore(t *testing.T) {
	storetest.TestPasswordResetStore(t, func(t *testing.T) (store.UserStore, store.PasswordResetStore) {
		s := openTestStore(t)
		resetTestDB(t, s)
		return s, s
	})
}

func TestCourseStore(t *testing.T) {
	storetest.TestCourseStore(t, func(t *testing.T) (store.UserStore, store.CourseStore) {
		s := openTestStore(t)
		resetTestDB(t, s)
		return s, s
	})
}

func openTestStore(t *testing.T) *Store {
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

func resetTestDB(t *testing.T, s *Store) {
	t.Helper()

	_, err := s.db.ExecContext(context.Background(), `
		TRUNCATE TABLE proposals RESTART IDENTITY CASCADE;
		TRUNCATE TABLE users RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		t.Fatalf("reset db: %v", err)
	}
}
