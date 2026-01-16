-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users'
        AND column_name = 'password_hash'
        AND data_type = 'text'
    ) THEN
        ALTER TABLE users
            ALTER COLUMN password_hash TYPE BYTEA
            USING password_hash::bytea;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users'
        AND column_name = 'password_hash'
        AND data_type = 'bytea'
    ) THEN
        ALTER TABLE users
            ALTER COLUMN password_hash TYPE TEXT
            USING encode(password_hash, 'escape');
    END IF;
END $$;
-- +goose StatementEnd
