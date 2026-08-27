-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
ALTER TYPE power_rarity ADD VALUE IF NOT EXISTS 'MYTHICAL';
-- +goose StatementEnd

-- +goose Down
-- Postgres cannot drop a value from an enum type. This migration is
-- deliberately irreversible - a Down here would either have to leave
-- MYTHICAL in place (a no-op, misleading) or rebuild the whole enum from
-- scratch (recreate the type, rewrite every dependent column), which is far
-- riskier than just documenting the limitation.
