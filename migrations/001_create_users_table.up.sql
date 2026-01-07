CREATE TYPE user_role AS ENUM (
    'student',
    'instructor',
    'admin'
);

CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role          user_role DEFAULT 'student',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
)
