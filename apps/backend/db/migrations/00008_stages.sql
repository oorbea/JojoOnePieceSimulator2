-- +goose Up
-- manga discriminates a Stage's source material - mirrors enums.Manga's
-- String() form exactly (JOJO/ONE_PIECE), same convention as every other
-- domain enum type in this schema.
-- +goose StatementBegin
CREATE TYPE manga AS ENUM ('JOJO', 'ONE_PIECE');
-- +goose StatementEnd

-- stages replaces StaticStageCatalog's hardcoded list (8 JoJo parts, 11 One
-- Piece sagas) with an admin-editable catalog. Stage names are proper nouns
-- ("Stardust Crusaders", "Wano Country") - same reasoning 00006_locales.sql
-- gives for leaving powers.name untranslated - so there is no translations
-- table here, only the manga code itself is i18n'd, client-side.
--
-- No unique constraint on (manga, position): an admin swapping two stages'
-- positions in one transaction would otherwise deadlock on it, and
-- game.Interleave/Gauntlet's round order sorts by position regardless, so a
-- transient duplicate is harmless.
-- +goose StatementBegin
CREATE TABLE stages (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    manga      manga       NOT NULL,
    position   integer     NOT NULL CHECK (position >= 0),
    name       text        NOT NULL CHECK (name <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (manga, name)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX stages_manga_position_idx ON stages (manga, position);
-- +goose StatementEnd

-- Seed: the same 19 stages StaticStageCatalog hardcoded (game/static_stage_catalog.go),
-- so removing that adapter changes nothing a player sees.
-- +goose StatementBegin
INSERT INTO stages (manga, position, name) VALUES
    ('JOJO', 0, 'Phantom Blood'),
    ('JOJO', 1, 'Battle Tendency'),
    ('JOJO', 2, 'Stardust Crusaders'),
    ('JOJO', 3, 'Diamond is Unbreakable'),
    ('JOJO', 4, 'Golden Wind'),
    ('JOJO', 5, 'Stone Ocean'),
    ('JOJO', 6, 'Steel Ball Run'),
    ('JOJO', 7, 'JoJolion'),
    ('ONE_PIECE', 0, 'East Blue'),
    ('ONE_PIECE', 1, 'Alabasta'),
    ('ONE_PIECE', 2, 'Sky Island'),
    ('ONE_PIECE', 3, 'Water Seven'),
    ('ONE_PIECE', 4, 'Thriller Bark'),
    ('ONE_PIECE', 5, 'Summit War'),
    ('ONE_PIECE', 6, 'Fish-Man Island'),
    ('ONE_PIECE', 7, 'Dressrosa'),
    ('ONE_PIECE', 8, 'Whole Cake Island'),
    ('ONE_PIECE', 9, 'Wano Country'),
    ('ONE_PIECE', 10, 'Egghead');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stages;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE manga;
-- +goose StatementEnd
