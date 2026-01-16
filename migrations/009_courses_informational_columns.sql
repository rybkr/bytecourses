-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'courses'
        AND column_name = 'target_audience'
    ) THEN
        ALTER TABLE courses ADD COLUMN target_audience TEXT NOT NULL DEFAULT '';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'courses'
        AND column_name = 'learning_objectives'
    ) THEN
        ALTER TABLE courses ADD COLUMN learning_objectives TEXT NOT NULL DEFAULT '';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'courses'
        AND column_name = 'assumed_prerequisites'
    ) THEN
        ALTER TABLE courses ADD COLUMN assumed_prerequisites TEXT NOT NULL DEFAULT '';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE courses DROP COLUMN IF EXISTS assumed_prerequisites;
ALTER TABLE courses DROP COLUMN IF EXISTS learning_objectives;
ALTER TABLE courses DROP COLUMN IF EXISTS target_audience;
