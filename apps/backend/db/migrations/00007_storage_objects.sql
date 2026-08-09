-- +goose Up
-- storage_objects is the single source of truth for which object-storage
-- provider (r2/b2/supabase - see the fallback storage chain) a given picture
-- key lives on, and how many bytes it occupies. powers.picture/picture_thumb
-- and users.avatar_key/avatar_thumb_key keep storing bare keys, unchanged -
-- this table is consulted (or, for keys predating it, defaulted to the
-- first configured tier) whenever a key needs presigning or deleting.
-- +goose StatementBegin
CREATE TABLE storage_objects (
    key        text   PRIMARY KEY,
    provider   text   NOT NULL,
    bytes      bigint NOT NULL CHECK (bytes >= 0),
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX storage_objects_provider_idx ON storage_objects (provider);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE storage_objects;
-- +goose StatementEnd
