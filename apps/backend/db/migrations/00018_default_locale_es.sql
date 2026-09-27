-- +goose Up
-- Switches the mandatory/default content locale from en-GB to es-ES (see
-- enums.DefaultLocale and ObsidianVault/i18n-multi-language.md) - the
-- application layer now requires and reads es-ES as the base translation
-- for powers/characters, and es-ES is the default UI language for a new
-- user.
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN language SET DEFAULT 'es-ES';
-- +goose StatementEnd

-- Backfill safety net: 00017_seed_catalog_*.sql already filled in es-ES for
-- every power/character it knew about at the time it was generated, but a
-- row created or edited in prod between that dump and this migration
-- running might still only have en-GB. Rather than block the whole catalog
-- read on that gap, copy en-GB -> es-ES wherever es-ES is still missing -
-- an admin can correct the copy afterwards through the normal edit flow.
-- +goose StatementBegin
INSERT INTO power_translations (power_id, locale, description, skills)
SELECT power_id, 'es-ES', description, skills
FROM power_translations src
WHERE src.locale = 'en-GB'
  AND NOT EXISTS (
    SELECT 1 FROM power_translations existing
    WHERE existing.power_id = src.power_id AND existing.locale = 'es-ES'
  );
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO character_translations (character_id, locale, description)
SELECT character_id, 'es-ES', description
FROM character_translations src
WHERE src.locale = 'en-GB'
  AND NOT EXISTS (
    SELECT 1 FROM character_translations existing
    WHERE existing.character_id = src.character_id AND existing.locale = 'es-ES'
  );
-- +goose StatementEnd

-- +goose Down
-- Only the column default is reverted - the backfilled es-ES translation
-- rows stay. Deleting them could remove a translation an admin has since
-- edited directly (indistinguishable at rollback time from the copy this
-- migration made), same reasoning as 00017's no-op Down.
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN language SET DEFAULT 'en-GB';
-- +goose StatementEnd
