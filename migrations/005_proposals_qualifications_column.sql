-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'proposals'
        AND column_name = 'qualifications'
    ) THEN
        ALTER TABLE proposals ADD COLUMN qualifications TEXT NOT NULL DEFAULT '';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE proposals DROP COLUMN IF EXISTS qualifications;
