CREATE TABLE assignments (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date TIMESTAMP,
    points INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE submissions (
    id SERIAL PRIMARY KEY,
    assignment_id INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'submitted',
    UNIQUE(assignment_id, student_id)
);

CREATE TABLE grades (
    id SERIAL PRIMARY KEY,
    submission_id INTEGER NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    instructor_id INTEGER NOT NULL REFERENCES users(id),
    score INTEGER,
    feedback TEXT,
    graded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assignments_course_id ON assignments(course_id);
CREATE INDEX idx_submissions_assignment_id ON submissions(assignment_id);
CREATE INDEX idx_submissions_student_id ON submissions(student_id);
CREATE INDEX idx_grades_submission_id ON grades(submission_id);

