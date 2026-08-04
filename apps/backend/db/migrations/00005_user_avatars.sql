-- +goose Up
-- google_picture is the avatar synced from the Google account on every
-- login - read-only from the API's point of view. avatar_key/avatar_thumb_key
-- are the user's own uploaded avatar (an R2 object key, produced by the same
-- background compression worker Stands/DevilFruits use), which the login
-- sync must never touch, so an uploaded avatar survives future logins.
-- avatar_status tracks that pipeline the same way picture_status does for
-- Powers.
-- +goose StatementBegin
ALTER TABLE users RENAME COLUMN picture TO google_picture;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN avatar_key       text NOT NULL DEFAULT '',
    ADD COLUMN avatar_thumb_key text NOT NULL DEFAULT '',
    ADD COLUMN avatar_status    picture_status NOT NULL DEFAULT 'NONE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN avatar_key,
    DROP COLUMN avatar_thumb_key,
    DROP COLUMN avatar_status;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users RENAME COLUMN google_picture TO picture;
-- +goose StatementEnd
