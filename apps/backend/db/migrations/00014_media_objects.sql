-- +goose Up
-- media_objects is the content-addressed index the immutable media proxy
-- (GET /api/v1/media/{group}/{variant}.webp) reads: group_id is derived
-- from the main rendition's bytes (see dto.MediaURLBuilder), so the same
-- source image always resolves to the same group regardless of when/how
-- often it was re-uploaded, and any pixel change gets a brand new group -
-- no invalidation needed anywhere. storage_key points at the exact same
-- object the existing presign path already uploaded; this table is purely
-- an additional index onto it, not a second copy of the bytes.
-- +goose StatementBegin
CREATE TABLE media_objects (
    group_id     text NOT NULL,
    variant      text NOT NULL,
    storage_key  text NOT NULL,
    content_type text NOT NULL,
    bytes        bigint NOT NULL,
    scope        text NOT NULL DEFAULT 'public',
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, variant)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX media_objects_storage_key_idx ON media_objects (storage_key);
-- +goose StatementEnd

-- picture_media_id/avatar_media_id are the group_id media_objects rows for
-- an entity's renditions are keyed by. Empty means "not backfilled yet" -
-- the DTO layer falls back to the existing presign path in that case (see
-- PictureURLResolver), so a deploy of this migration + the media route is
-- safe before mediabackfill has run.
-- +goose StatementBegin
ALTER TABLE powers ADD COLUMN picture_media_id text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE stages ADD COLUMN picture_media_id text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users ADD COLUMN avatar_media_id text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE powers DROP COLUMN picture_media_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE stages DROP COLUMN picture_media_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users DROP COLUMN avatar_media_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE media_objects;
-- +goose StatementEnd
