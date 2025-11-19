CREATE TABLE courses (
    id SERIAL PRIMARY KEY,
    instructor_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_courses_status ON courses(status);

INSERT INTO courses (instructor_id, title, description, status, created_at, updated_at)
SELECT instructor_id, title, description, status, created_at, updated_at
FROM applications
WHERE status IN ('draft', 'pending');

INSERT INTO courses (instructor_id, title, description, status, created_at, updated_at)
SELECT instructor_id, title, description, status, created_at, updated_at
FROM applications
WHERE status = 'rejected';

INSERT INTO courses (instructor_id, title, description, status, created_at, updated_at)
SELECT instructor_id, title, description, 'approved', created_at, updated_at
FROM courses
WHERE id NOT IN (SELECT id FROM courses WHERE status IS NOT NULL);

DROP TABLE applications;
