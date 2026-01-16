-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'courses'
        AND column_name = 'proposal_id'
    ) THEN
        ALTER TABLE courses ADD COLUMN proposal_id BIGINT NULL REFERENCES proposals(id) ON DELETE SET NULL;
    END IF;
END $$;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS courses_proposal_id_idx ON courses(proposal_id);

-- +goose Down
DROP INDEX IF EXISTS courses_proposal_id_idx;
ALTER TABLE courses DROP COLUMN IF EXISTS proposal_id;
