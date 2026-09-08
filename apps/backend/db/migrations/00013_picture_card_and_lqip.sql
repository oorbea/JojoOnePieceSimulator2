-- +goose Up
-- picture_card holds the new "card" rendition's object-storage key (128px,
-- the size most catalogue grid cells actually render at - see
-- ObsidianVault/entrega-imagenes-red-lenta-2026-09-07.md), alongside the
-- existing picture/picture_thumb columns. picture_lqip stores a complete
-- data: URI (tiny WebP, ~16px) for an instant low-quality placeholder while
-- the real rendition loads - both are produced by the same async
-- compression worker as picture_thumb, so they default empty until the
-- worker has run at least once for a given row.
-- +goose StatementBegin
ALTER TABLE powers
    ADD COLUMN picture_card text NOT NULL DEFAULT '',
    ADD COLUMN picture_lqip text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE stages
    ADD COLUMN picture_card text NOT NULL DEFAULT '',
    ADD COLUMN picture_lqip text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN avatar_card_key text NOT NULL DEFAULT '',
    ADD COLUMN avatar_lqip     text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE powers
    DROP COLUMN picture_card,
    DROP COLUMN picture_lqip;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE stages
    DROP COLUMN picture_card,
    DROP COLUMN picture_lqip;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN avatar_card_key,
    DROP COLUMN avatar_lqip;
-- +goose StatementEnd
