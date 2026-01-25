-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'reading_format') THEN
        CREATE TYPE reading_format AS ENUM (
            'markdown'
        );
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
