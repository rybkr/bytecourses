-- +goose Up
CREATE TYPE proposal_status AS ENUM (
	'draft',
	'submitted',
	'withdrawn',
	'approved',
	'rejected',
	'changes_requested'
);

CREATE TABLE proposals (
    id                    BIGSERIAL PRIMARY KEY,
    author_id             BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title                 TEXT NOT NULL default '',
    summary               TEXT NOT NULL default '',
    target_audience       TEXT NOT NULL default '',
    learning_objectives   TEXT NOT NULL default '',
    outline               TEXT NOT NULL default '',
    assumed_prerequisites TEXT NOT NULL DEFAULT '',
    status                proposal_status NOT NULL default 'draft',
    reviewer_id           BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    review_notes          TEXT NOT NULL DEFAULT '',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX proposals_author_id_idx ON proposals(author_id);
CREATE INDEX proposals_status_idx ON proposals(status);

-- +goose Down
DROP TABLE IF EXISTS proposals;
DROP TYPE IF EXISTS proposal_status;
