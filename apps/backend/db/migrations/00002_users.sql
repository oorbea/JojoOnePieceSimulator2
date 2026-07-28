-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('REGULAR', 'ADMIN');
-- +goose StatementEnd

-- Users authenticate exclusively via Google. `google_sub` is the stable
-- identity ('sub' claim of the verified ID token) and is what logins are
-- keyed on; `email` is kept unique too since it is also used to link a
-- pre-existing account and to bootstrap admins via ADMIN_EMAILS.
-- +goose StatementBegin
CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    google_sub    text NOT NULL UNIQUE CHECK (google_sub <> ''),
    email         text NOT NULL UNIQUE CHECK (email <> ''),
    username      text NOT NULL UNIQUE CHECK (username <> ''),
    complete_name text NOT NULL DEFAULT '',
    picture       text NOT NULL DEFAULT '',
    role          user_role NOT NULL DEFAULT 'REGULAR',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
