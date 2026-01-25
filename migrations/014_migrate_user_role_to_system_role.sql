-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'system_role') THEN
        CREATE TYPE system_role AS ENUM (
            'user',
            'admin'
        );
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE users ADD COLUMN IF NOT EXISTS system_role_temp system_role;

UPDATE users 
SET system_role_temp = CASE 
    WHEN role::text = 'admin' THEN 'admin'::system_role
    ELSE 'user'::system_role
END;

ALTER TABLE users 
    ALTER COLUMN system_role_temp SET NOT NULL,
    ALTER COLUMN system_role_temp SET DEFAULT 'user'::system_role;
ALTER TABLE users DROP COLUMN role;
ALTER TABLE users RENAME COLUMN system_role_temp TO role;

DROP TYPE IF EXISTS user_role;

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM (
            'student',
            'instructor',
            'admin'
        );
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE users ADD COLUMN role_old user_role;

UPDATE users 
SET role_old = CASE 
    WHEN role::text = 'admin' THEN 'admin'::user_role
    ELSE 'student'::user_role
END;

ALTER TABLE users 
    ALTER COLUMN role_old SET NOT NULL,
    ALTER COLUMN role_old SET DEFAULT 'student'::user_role;
ALTER TABLE users DROP COLUMN role;
ALTER TABLE users RENAME COLUMN role_old TO role;

DROP TYPE IF EXISTS system_role;
