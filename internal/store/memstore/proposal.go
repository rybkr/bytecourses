package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"errors"
	"sync"
	"time"
)

type ProposalStore struct {
	mu            sync.RWMutex
	proposalsByID map[int64]domain.Proposal
	nextID        int64
}

func NewProposalStore() *ProposalStore {
	return &ProposalStore{
		proposalsByID: make(map[int64]domain.Proposal),
		nextID:        1,
	}
}

func (s *ProposalStore) InsertProposal(ctx context.Context, p *domain.Proposal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p.ID = s.nextID
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}

	s.proposalsByID[p.ID] = *p
	s.nextID++

	return nil
}

func (s *ProposalStore) GetProposalByID(ctx context.Context, id int64) (domain.Proposal, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.proposalsByID[id]
	return p, ok
}

func (s *ProposalStore) GetProposalsByUserID(ctx context.Context, userID int64) []domain.Proposal {
	out := make([]domain.Proposal, 0)
	for _, p := range s.proposalsByID {
		if p.AuthorID == userID {
			out = append(out, p)
		}
	}
	return out
}

func (s *ProposalStore) UpdateProposal(ctx context.Context, p *domain.Proposal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.proposalsByID[p.ID]; !exists {
		return errors.New("proposal does not exist")
	}

	s.proposalsByID[p.ID] = *p

	return nil
}
