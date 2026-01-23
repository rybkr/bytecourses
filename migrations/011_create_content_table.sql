-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'content_status') THEN
        CREATE TYPE content_status AS ENUM (
            'draft',
            'published'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'content_type') THEN
        CREATE TYPE content_type AS ENUM (
            'reading'
        );
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS content (
    id            BIGSERIAL PRIMARY KEY,
    module_id     BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    content_type  content_type NOT NULL,
    title         TEXT NOT NULL DEFAULT '',
    order_index   INT NOT NULL,
    status        content_status NOT NULL DEFAULT 'draft',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS content_module_id_idx ON content(module_id);
CREATE INDEX IF NOT EXISTS content_content_type_idx ON content(content_type);
CREATE INDEX IF NOT EXISTS content_status_idx ON content(status);
CREATE INDEX IF NOT EXISTS content_module_order_idx ON content(module_id, order_index);

-- +goose Down
DROP INDEX IF EXISTS content_module_order_idx;
DROP INDEX IF EXISTS content_status_idx;
DROP INDEX IF EXISTS content_content_type_idx;
DROP INDEX IF EXISTS content_module_id_idx;
DROP TABLE IF EXISTS content;
DROP TYPE IF EXISTS content_type;
DROP TYPE IF EXISTS content_status;
