package sqlstore

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"database/sql"
	"github.com/jackc/pgconn"
	"time"
)

func (s *Store) CreateProposal(ctx context.Context, p *domain.Proposal) error {
	now := time.Now().UTC()
	status := p.Status
	if status == "" {
		status = domain.ProposalStatusDraft
	}

	if err := s.db.QueryRowContext(ctx, `
		INSERT INTO proposals (
			author_id,
			title, summary, qualifications,
			target_audience, learning_objectives, outline, assumed_prerequisites,
			status,
			reviewer_id, review_notes,
			created_at, updated_at
		) VALUES (
			$1,
			$2, $3, $4,
			$5, $6, $7, $8,
			$9,
			$10, $11,
			$12, $13
		)
		RETURNING id
	`,
		p.AuthorID,
		p.Title, p.Summary, p.Qualifications,
		p.TargetAudience, p.LearningObjectives, p.Outline, p.AssumedPrerequisites,
		string(status),
		nullInt64Ptr(p.ReviewerID), p.ReviewNotes,
		now, now,
	).Scan(&p.ID); err != nil {
		return err
	}

	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

func (s *Store) GetProposalByID(ctx context.Context, id int64) (*domain.Proposal, bool) {
	var p domain.Proposal
	var status string
	var reviewer sql.NullInt64

	if err := s.db.QueryRowContext(ctx, `
		SELECT id, author_id,
		       title, summary, qualifications,
		       target_audience, learning_objectives, outline, assumed_prerequisites,
		       status,
		       reviewer_id, review_notes,
		       created_at, updated_at
		  FROM proposals
		 WHERE id = $1
	`, id).Scan(
		&p.ID, &p.AuthorID,
		&p.Title, &p.Summary, &p.Qualifications,
		&p.TargetAudience, &p.LearningObjectives, &p.Outline, &p.AssumedPrerequisites,
		&status,
		&reviewer, &p.ReviewNotes,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, false
	}

	p.Status = domain.ProposalStatus(status)
	p.ReviewerID = ptrFromNullInt64(reviewer)
	return &p, true
}

func (s *Store) ListProposalsByAuthorID(ctx context.Context, authorID int64) ([]domain.Proposal, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, author_id,
		       title, summary, qualifications,
		       target_audience, learning_objectives, outline, assumed_prerequisites,
		       status,
		       reviewer_id, review_notes,
		       created_at, updated_at
		  FROM proposals
		 WHERE author_id = $1
		 ORDER BY updated_at DESC, id DESC
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Proposal, 0, 8)
	for rows.Next() {
		var p domain.Proposal
		var status string
		var reviewer sql.NullInt64

		if err := rows.Scan(
			&p.ID, &p.AuthorID,
			&p.Title, &p.Summary, &p.Qualifications,
			&p.TargetAudience, &p.LearningObjectives, &p.Outline, &p.AssumedPrerequisites,
			&status,
			&reviewer, &p.ReviewNotes,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		p.Status = domain.ProposalStatus(status)
		p.ReviewerID = ptrFromNullInt64(reviewer)
		out = append(out, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) ListAllSubmittedProposals(ctx context.Context) ([]domain.Proposal, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, author_id,
		       title, summary, qualifications,
		       target_audience, learning_objectives, outline, assumed_prerequisites,
		       status,
		       reviewer_id, review_notes,
		       created_at, updated_at
		  FROM proposals
		 WHERE status IN ('submitted', 'approved', 'rejected', 'changes_requested')
		 ORDER BY updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Proposal, 0, 16)
	for rows.Next() {
		var p domain.Proposal
		var status string
		var reviewer sql.NullInt64

		if err := rows.Scan(
			&p.ID, &p.AuthorID,
			&p.Title, &p.Summary, &p.Qualifications,
			&p.TargetAudience, &p.LearningObjectives, &p.Outline, &p.AssumedPrerequisites,
			&status,
			&reviewer, &p.ReviewNotes,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		p.Status = domain.ProposalStatus(status)
		p.ReviewerID = ptrFromNullInt64(reviewer)
		out = append(out, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) UpdateProposal(ctx context.Context, p *domain.Proposal) error {
	now := time.Now().UTC()
	status := p.Status
	if status == "" {
		status = domain.ProposalStatusDraft
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE proposals
		   SET title = $2,
		       summary = $3,
		       qualifications = $4,
		       target_audience = $5,
		       learning_objectives = $6,
		       outline = $7,
		       assumed_prerequisites = $8,
		       status = $9,
		       reviewer_id = $10,
		       review_notes = $11,
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
		string(status),
		nullInt64Ptr(p.ReviewerID),
		p.ReviewNotes,
		now,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			return err
		}
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}

	p.UpdatedAt = now
	return nil
}

func (s *Store) DeleteProposalByID(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM proposals WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

func nullInt64Ptr(p *int64) any {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: *p,
		Valid: true,
	}
}

func ptrFromNullInt64(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}
