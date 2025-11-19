CREATE TABLE applications (
    id SERIAL PRIMARY KEY,
    instructor_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    learning_objectives TEXT,
    prerequisites TEXT,
    course_format VARCHAR(50),
    category_tags VARCHAR(200),
    skill_level VARCHAR(20),
    course_duration VARCHAR(100),
    instructor_qualifications TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    rejected_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_applications_status ON applications(status);
CREATE INDEX idx_applications_instructor_id ON applications(instructor_id);
CREATE INDEX idx_applications_rejected_at ON applications(rejected_at);

INSERT INTO applications (instructor_id, title, description, status, created_at, updated_at)
SELECT instructor_id, title, description, status, created_at, updated_at
FROM courses
WHERE status IN ('draft', 'pending');

INSERT INTO applications (instructor_id, title, description, status, rejected_at, created_at, updated_at)
SELECT instructor_id, title, description, status, updated_at, created_at, updated_at
FROM courses
WHERE status = 'rejected';

CREATE TEMP TABLE approved_courses_temp AS
SELECT id, instructor_id, title, description, created_at, updated_at
FROM courses
WHERE status = 'approved';

DROP TABLE courses;

CREATE TABLE courses (
    id SERIAL PRIMARY KEY,
    instructor_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_courses_instructor_id ON courses(instructor_id);

INSERT INTO courses (id, instructor_id, title, description, created_at, updated_at)
SELECT id, instructor_id, title, description, created_at, updated_at
FROM approved_courses_temp;

SELECT setval('courses_id_seq', COALESCE((SELECT MAX(id) FROM courses), 1), true);
