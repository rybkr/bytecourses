-- +goose Up
ALTER TABLE courses ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE courses DROP COLUMN IF EXISTS updated_at;
