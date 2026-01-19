package postgres

import (
	"context"
	"database/sql"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

var (
	_ persistence.ProposalRepository = (*ProposalRepository)(nil)
)

type ProposalRepository struct {
	db *sql.DB
}

func NewProposalRepository(db *DB) *ProposalRepository {
	return &ProposalRepository{
		db: db.DB(),
	}
}

func (r *ProposalRepository) Create(ctx context.Context, p *domain.Proposal) error {
	now := time.Now().UTC()

	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO proposals (
			title, summary, qualifications, target_audience,
			learning_objectives, outline, assumed_prerequisites,
			author_id, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		p.Title,
		p.Summary,
		p.Qualifications,
		p.TargetAudience,
		p.LearningObjectives,
		p.Outline,
		p.AssumedPrerequisites,
		p.AuthorID,
		string(p.Status),
		now,
		now,
	).Scan(&p.ID); err != nil {
		return err
	}

	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

func (r *ProposalRepository) GetByID(ctx context.Context, id int64) (*domain.Proposal, bool) {
	var p domain.Proposal
	var status string

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, title, summary, qualifications, target_audience,
		       learning_objectives, outline, assumed_prerequisites,
		       author_id, reviewer_id, review_notes, status, created_at, updated_at
		FROM proposals
		WHERE id = $1
	`, id).Scan(
		&p.ID,
		&p.Title,
		&p.Summary,
		&p.Qualifications,
		&p.TargetAudience,
		&p.LearningObjectives,
		&p.Outline,
		&p.AssumedPrerequisites,
		&p.AuthorID,
		&p.ReviewerID,
		&p.ReviewNotes,
		&status,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return nil, false
	}

	p.Status = domain.ProposalStatus(status)
	return &p, true
}

func (r *ProposalRepository) ListByAuthorID(ctx context.Context, authorID int64) ([]domain.Proposal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, summary, qualifications, target_audience,
		       learning_objectives, outline, assumed_prerequisites,
		       author_id, reviewer_id, review_notes, status, created_at, updated_at
		FROM proposals
		WHERE author_id = $1
		ORDER BY created_at DESC
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProposals(rows)
}

func (r *ProposalRepository) ListAllSubmitted(ctx context.Context) ([]domain.Proposal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, summary, qualifications, target_audience,
		       learning_objectives, outline, assumed_prerequisites,
		       author_id, reviewer_id, review_notes, status, created_at, updated_at
		FROM proposals
		WHERE status IN ('submitted', 'approved', 'rejected', 'changes_requested')
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProposals(rows)
}

func (r *ProposalRepository) Update(ctx context.Context, p *domain.Proposal) error {
	p.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		UPDATE proposals
		SET title = $2,
		    summary = $3,
		    qualifications = $4,
		    target_audience = $5,
		    learning_objectives = $6,
		    outline = $7,
		    assumed_prerequisites = $8,
		    reviewer_id = $9,
		    review_notes = $10,
		    status = $11,
		    updated_at = $12
		WHERE id = $1
	`,
		p.ID,
		p.Title,
		p.Summary,
		p.Qualifications,
		p.TargetAudience,
		p.LearningObjectives,
		p.Outline,
		p.AssumedPrerequisites,
		p.ReviewerID,
		p.ReviewNotes,
		string(p.Status),
		p.UpdatedAt,
	)
	return err
}

func (r *ProposalRepository) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM proposals WHERE id = $1`, id)
	return err
}

func scanProposals(rows *sql.Rows) ([]domain.Proposal, error) {
	proposals := make([]domain.Proposal, 0)

	for rows.Next() {
		var p domain.Proposal
		var status string

		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Summary,
			&p.Qualifications,
			&p.TargetAudience,
			&p.LearningObjectives,
			&p.Outline,
			&p.AssumedPrerequisites,
			&p.AuthorID,
			&p.ReviewerID,
			&p.ReviewNotes,
			&status,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		p.Status = domain.ProposalStatus(status)
		proposals = append(proposals, p)
	}

	return proposals, rows.Err()
}
