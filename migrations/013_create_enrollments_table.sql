-- +goose Up
CREATE TABLE IF NOT EXISTS enrollments (
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id   BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, course_id)
);

CREATE INDEX IF NOT EXISTS enrollments_user_id_idx ON enrollments(user_id);
CREATE INDEX IF NOT EXISTS enrollments_course_id_idx ON enrollments(course_id);

-- +goose Down
DROP INDEX IF EXISTS enrollments_course_id_idx;
DROP INDEX IF EXISTS enrollments_user_id_idx;
DROP TABLE IF EXISTS enrollments;
