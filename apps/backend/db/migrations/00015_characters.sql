-- +goose Up
-- Five stat enums that have existed in Go since the game domain layer
-- (haki_level.go, physical_form.go, fruit_mastery.go, hamon_level.go,
-- spin_level.go) but were never persisted anywhere - live game state lives
-- in Redis, not Postgres. Literals mirror each type's String() form exactly,
-- same convention as every other domain enum type in this schema.
-- +goose StatementBegin
CREATE TYPE haki_level AS ENUM ('NONE', 'PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE physical_form AS ENUM ('PRIVATE', 'STRONG_FISHMAN', 'MARINE_CAPTAIN', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE fruit_mastery AS ENUM ('NONE', 'REGULAR', 'ADVANCED', 'AWAKENED');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE hamon_level AS ENUM ('NONE', 'BASIC', 'ADVANCED', 'PERFECT');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE spin_level AS ENUM ('NONE', 'BASIC', 'GOLDEN', 'INFINITE');
-- +goose StatementEnd

-- Base table of a class-table-inheritance hierarchy separate from powers'
-- (a character is not a Power: it never enters the random draw, PoolFilter
-- or banlist). `manga` is the discriminator, playing the role power_kind
-- plays for stands/devil_fruits - already typed, already the natural split
-- between the two subtype tables. The composite FK + CHECK each subtype
-- table declares below is what makes the inheritance disjoint and stops an
-- admin from repointing a character's manga once it has a subtype row.
--
-- No `description` column: character_translations below is the only home
-- for it, same reasoning 00006_locales.sql gives for power_translations.
-- Picture columns mirror powers/stages' full pipeline (00001, 00003,
-- 00013, 00014) as of today, not their older intermediate shapes.
-- +goose StatementBegin
CREATE TABLE characters (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    manga            manga        NOT NULL,
    name             text         NOT NULL CHECK (name <> ''),
    rarity           power_rarity NOT NULL,
    picture          text         NOT NULL DEFAULT '',
    picture_thumb    text         NOT NULL DEFAULT '',
    picture_card     text         NOT NULL DEFAULT '',
    picture_lqip     text         NOT NULL DEFAULT '',
    picture_status   picture_status NOT NULL DEFAULT 'NONE',
    picture_media_id text         NOT NULL DEFAULT '',
    created_at       timestamptz  NOT NULL DEFAULT now(),
    updated_at       timestamptz  NOT NULL DEFAULT now(),
    UNIQUE (manga, name),
    UNIQUE (id, manga)
);
-- +goose StatementEnd

-- character_translations is power_translations' shape without `skills` (a
-- Character has none, decision: content only) - only en-GB is mandatory,
-- enforced in the application layer the same way as Stands/Devil Fruits,
-- not Stages' all-three rule.
-- +goose StatementBegin
CREATE TABLE character_translations (
    character_id uuid   NOT NULL REFERENCES characters (id) ON DELETE CASCADE,
    locale       locale NOT NULL,
    description  text   NOT NULL CHECK (description <> ''),
    PRIMARY KEY (character_id, locale)
);
-- +goose StatementEnd

-- Subtype table for a JoJo character. battle_iq is smallint (Postgres has
-- no unsigned byte) but stays `byte` in Go; it is NOT NULL because every
-- JoJo character - unlike a live loadout's battleIQ, which can be absent
-- for a non-JoJo lobby - always has a WAIS-IV score assigned by the admin
-- authoring it.
-- +goose StatementBegin
CREATE TABLE jojo_characters (
    id        uuid PRIMARY KEY,
    manga     manga NOT NULL DEFAULT 'JOJO' CHECK (manga = 'JOJO'),
    hamon     hamon_level NOT NULL,
    spin      spin_level  NOT NULL,
    battle_iq smallint    NOT NULL CHECK (battle_iq BETWEEN 0 AND 255),
    CONSTRAINT jojo_characters_character_fk FOREIGN KEY (id, manga)
        REFERENCES characters (id, manga) ON DELETE CASCADE
);
-- +goose StatementEnd

-- Subtype table for a One Piece character.
-- +goose StatementBegin
CREATE TABLE one_piece_characters (
    id               uuid PRIMARY KEY,
    manga            manga NOT NULL DEFAULT 'ONE_PIECE' CHECK (manga = 'ONE_PIECE'),
    physical_form    physical_form NOT NULL,
    armament_haki    haki_level    NOT NULL,
    observation_haki haki_level    NOT NULL,
    conqueror_haki   haki_level    NOT NULL,
    fruit_mastery    fruit_mastery NOT NULL,
    CONSTRAINT one_piece_characters_character_fk FOREIGN KEY (id, manga)
        REFERENCES characters (id, manga) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS one_piece_characters;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS jojo_characters;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS character_translations;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS characters;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS spin_level;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS hamon_level;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS fruit_mastery;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS physical_form;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS haki_level;
-- +goose StatementEnd
