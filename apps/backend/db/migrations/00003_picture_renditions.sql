-- +goose Up
-- +goose StatementBegin
CREATE TYPE picture_status AS ENUM ('NONE', 'PENDING', 'READY', 'FAILED');
-- +goose StatementEnd

-- picture_thumb holds the thumbnail rendition's object-storage key, kept
-- alongside picture (the main rendition's key) since both are produced
-- together by the async compression worker. picture_status tracks where a
-- Power's picture is in that pipeline: a freshly uploaded picture goes to
-- PENDING immediately (picture/picture_thumb still point at the previous
-- renditions, if any) and only moves to READY once the worker has replaced
-- them, or to FAILED if it could not.
-- +goose StatementBegin
ALTER TABLE powers
    ADD COLUMN picture_thumb  text NOT NULL DEFAULT '',
    ADD COLUMN picture_status picture_status NOT NULL DEFAULT 'NONE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE powers
    DROP COLUMN picture_thumb,
    DROP COLUMN picture_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS picture_status;
-- +goose StatementEnd
