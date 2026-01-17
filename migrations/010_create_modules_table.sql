-- +goose Up
CREATE TABLE modules (
    id         BIGSERIAL PRIMARY KEY,
    course_id  BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title      TEXT NOT NULL DEFAULT '',
    position   INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX modules_course_id_idx ON modules(course_id);

-- +goose Down
DROP TABLE IF EXISTS modules;
