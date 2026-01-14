-- +goose Up
CREATE TYPE course_status AS ENUM (
    'draft',
    'live'
);

CREATE TABLE COURSES (
    id            BIGSERIAL PRIMARY KEY,
    title         TEXT NOT NULL default '',
    summary       TEXT NOT NULL default '',
    instructor_id BIGINT NOT NULL REFERENCES users(id), --! ON DELETE CASCADE
    status        course_status NOT NULL DEFAULT 'draft',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX courses_instructor_id_idx ON courses(instructor_id);
CREATE INDEX courses_status_idx ON courses(status);

-- +goose Down

DROP TABLE IF EXISTS courses;
DROP TYPE IF EXISTS course_status;
