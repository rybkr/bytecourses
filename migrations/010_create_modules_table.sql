-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'module_status') THEN
        CREATE TYPE module_status AS ENUM (
            'draft',
            'published'
        );
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS modules (
    id            BIGSERIAL PRIMARY KEY,
    course_id     BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title         TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL DEFAULT '',
    order_index   INT NOT NULL,
    status        module_status NOT NULL DEFAULT 'draft',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS modules_course_id_idx ON modules(course_id);
CREATE INDEX IF NOT EXISTS modules_status_idx ON modules(status);
CREATE INDEX IF NOT EXISTS modules_course_order_idx ON modules(course_id, order_index);

-- +goose Down
DROP INDEX IF EXISTS modules_course_order_idx;
DROP INDEX IF EXISTS modules_status_idx;
DROP INDEX IF EXISTS modules_course_id_idx;
DROP TABLE IF EXISTS modules;
DROP TYPE IF EXISTS module_status;
