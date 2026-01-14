-- +goose Up
ALTER TABLE courses ADD COLUMN proposal_id BIGINT NULL REFERENCES proposals(id) ON DELETE SET NULL;
CREATE INDEX courses_proposal_id_idx ON courses(proposal_id);

-- +goose Down
DROP INDEX IF EXISTS courses_proposal_id_idx;
ALTER TABLE courses DROP COLUMN IF EXISTS proposal_id;
