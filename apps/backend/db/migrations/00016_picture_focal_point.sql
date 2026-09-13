-- +goose Up
-- Lets an admin pick which part of a picture stays visible when a client
-- renders it with contentFit:'cover' (catalogue cards, the in-game stage
-- banner, avatars) instead of always cropping to center. No server-side
-- crop and no re-transcode: the worker/vips pipeline is untouched, this is
-- purely a normalized point clients read through picture-source.ts. Default
-- 0.5/0.5 (dead center) reproduces today's behaviour for every existing row
-- with no backfill needed.
-- +goose StatementBegin
ALTER TABLE powers
    ADD COLUMN focal_x double precision NOT NULL DEFAULT 0.5 CHECK (focal_x >= 0 AND focal_x <= 1),
    ADD COLUMN focal_y double precision NOT NULL DEFAULT 0.5 CHECK (focal_y >= 0 AND focal_y <= 1);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE stages
    ADD COLUMN focal_x double precision NOT NULL DEFAULT 0.5 CHECK (focal_x >= 0 AND focal_x <= 1),
    ADD COLUMN focal_y double precision NOT NULL DEFAULT 0.5 CHECK (focal_y >= 0 AND focal_y <= 1);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE characters
    ADD COLUMN focal_x double precision NOT NULL DEFAULT 0.5 CHECK (focal_x >= 0 AND focal_x <= 1),
    ADD COLUMN focal_y double precision NOT NULL DEFAULT 0.5 CHECK (focal_y >= 0 AND focal_y <= 1);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN avatar_focal_x double precision NOT NULL DEFAULT 0.5 CHECK (avatar_focal_x >= 0 AND avatar_focal_x <= 1),
    ADD COLUMN avatar_focal_y double precision NOT NULL DEFAULT 0.5 CHECK (avatar_focal_y >= 0 AND avatar_focal_y <= 1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE powers
    DROP COLUMN focal_x,
    DROP COLUMN focal_y;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE stages
    DROP COLUMN focal_x,
    DROP COLUMN focal_y;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE characters
    DROP COLUMN focal_x,
    DROP COLUMN focal_y;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN avatar_focal_x,
    DROP COLUMN avatar_focal_y;
-- +goose StatementEnd
