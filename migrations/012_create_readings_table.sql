-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'reading_format') THEN
        CREATE TYPE reading_format AS ENUM (
            'markdown',
            'plain',
            'html'
        );
    ELSE
        IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'plain' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'reading_format')) THEN
            ALTER TYPE reading_format ADD VALUE 'plain';
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'html' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'reading_format')) THEN
            ALTER TYPE reading_format ADD VALUE 'html';
        END IF;
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS readings (
    content_item_id BIGINT PRIMARY KEY REFERENCES content(id) ON DELETE CASCADE,
    format          reading_format NOT NULL,
    content         TEXT
);

-- +goose Down
DROP TABLE IF EXISTS readings;
DROP TYPE IF EXISTS reading_format;
