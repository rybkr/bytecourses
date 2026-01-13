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

func (s *ProposalStore) CreateProposal(ctx context.Context, p *domain.Proposal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p.ID = s.nextID
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt

	s.proposalsByID[p.ID] = *p
	s.nextID++

	return nil
}

func (s *ProposalStore) GetProposalByID(ctx context.Context, id int64) (*domain.Proposal, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if p, ok := s.proposalsByID[id]; ok {
		return &p, true
	}
	return nil, false
}

func (s *ProposalStore) ListProposalsByAuthorID(ctx context.Context, uid int64) ([]domain.Proposal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Proposal, 0)
	for _, p := range s.proposalsByID {
		if p.AuthorID == uid {
			out = append(out, p)
		}
	}

	return out, nil
}

func (s *ProposalStore) ListAllSubmittedProposals(ctx context.Context) ([]domain.Proposal, error) {
	out := make([]domain.Proposal, 0)
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.proposalsByID {
		if p.WasSubmitted() {
			out = append(out, p)
		}
	}

	return out, nil
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

func (s *ProposalStore) DeleteProposalByID(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.proposalsByID[id]; !exists {
		return errors.New("proposal does not exist")
	}

	delete(s.proposalsByID, id)
	return nil
}
