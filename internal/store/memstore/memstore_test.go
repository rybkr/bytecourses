package memstore

import (
	"bytecourses/internal/store"
	"bytecourses/internal/store/storetest"
	"testing"
)

func TestUserStore(t *testing.T) {
	storetest.TestUserStore(t, func(t *testing.T) store.UserStore {
		return NewUserStore()
	})
}

func TestProposalStore(t *testing.T) {
	storetest.TestProposalStore(t, func(t *testing.T) (store.UserStore, store.ProposalStore) {
		return NewUserStore(), NewProposalStore()
	})
}

func TestPasswordResetStore(t *testing.T) {
	storetest.TestPasswordResetStore(t, func(t *testing.T) (store.UserStore, store.PasswordResetStore) {
		return NewUserStore(), NewPasswordResetStore()
	})
}
