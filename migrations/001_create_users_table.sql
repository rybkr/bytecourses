-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE user_role AS ENUM (
    'student',
    'instructor',
    'admin'
);

CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         CITEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role          user_role NOT NULL DEFAULT 'student',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
DROP EXTENSION IF EXISTS citext;
