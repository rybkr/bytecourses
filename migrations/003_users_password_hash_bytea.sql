-- +goose Up
ALTER TABLE users
    ALTER COLUMN password_hash TYPE BYTEA
    USING password_hash::bytea;

-- +goose Down
ALTER TABLE users
    ALTER COLUMN password_hash TYPE TEXT
    USING encode(password_hash, 'escape');
