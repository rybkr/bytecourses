-- +goose Up
ALTER TABLE proposals
ADD qualifications TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE proposals
DROP COLUMN qualifications;
