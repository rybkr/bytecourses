package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

var (
	_ persistence.ProposalRepository = (*ProposalRepository)(nil)
)

type ProposalRepository struct {
	mu        sync.RWMutex
	proposals map[int64]domain.Proposal
	nextID    int64
}

func NewProposalRepository() *ProposalRepository {
	return &ProposalRepository{
		proposals: make(map[int64]domain.Proposal),
		nextID:    1,
	}
}

func (r *ProposalRepository) Create(ctx context.Context, p *domain.Proposal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p.ID = r.nextID
	r.nextID++
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	r.proposals[p.ID] = *p

	return nil
}

func (r *ProposalRepository) GetByID(ctx context.Context, id int64) (*domain.Proposal, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.proposals[id]
	if !ok {
		return nil, false
	}

	return &p, true
}

func (r *ProposalRepository) ListByAuthorID(ctx context.Context, authorID int64) ([]domain.Proposal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Proposal, 0)
	for _, p := range r.proposals {
		if p.AuthorID == authorID {
			result = append(result, p)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result, nil
}

func (r *ProposalRepository) ListAllSubmitted(ctx context.Context) ([]domain.Proposal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Proposal, 0)
	for _, p := range r.proposals {
		if p.Status == domain.ProposalStatusSubmitted ||
			p.Status == domain.ProposalStatusApproved ||
			p.Status == domain.ProposalStatusRejected ||
			p.Status == domain.ProposalStatusChangesRequested {
			result = append(result, p)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})

	return result, nil
}

func (r *ProposalRepository) Update(ctx context.Context, p *domain.Proposal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.proposals[p.ID]; !ok {
		return nil
	}

	p.UpdatedAt = time.Now()
	r.proposals[p.ID] = *p

	return nil
}

func (r *ProposalRepository) DeleteByID(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.proposals, id)
	return nil
}
