-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'file' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'content_type')) THEN
        ALTER TYPE content_type ADD VALUE 'file';
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS files (
    content_item_id BIGINT PRIMARY KEY REFERENCES content(id) ON DELETE CASCADE,
    file_name       TEXT NOT NULL,
    file_size       BIGINT NOT NULL,
    mime_type       TEXT NOT NULL,
    storage_path    TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS files;
