-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE power_rarity AS ENUM ('COMMON', 'RARE', 'EPIC', 'LEGENDARY');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE stand_stat AS ENUM ('E', 'D', 'C', 'B', 'A', 'INFINITE', 'NULL');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE power_kind AS ENUM ('STAND', 'DEVIL_FRUIT');
-- +goose StatementEnd

-- Base table of the class table inheritance hierarchy: every Power (Stand,
-- DevilFruit, ...) has a row here. `kind` plus the (id, kind) unique
-- constraint lets child tables enforce, via a composite FK, that a given
-- power id can only ever belong to exactly one subtype table.
-- +goose StatementBegin
CREATE TABLE powers (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind        power_kind NOT NULL,
    name        text NOT NULL UNIQUE CHECK (name <> ''),
    description text NOT NULL CHECK (description <> ''),
    rarity      power_rarity NOT NULL,
    picture     text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (id, kind)
);
-- +goose StatementEnd

-- Ordered skills belonging to a power (Power.skills []string in the domain).
-- +goose StatementBegin
CREATE TABLE power_skills (
    power_id uuid NOT NULL REFERENCES powers (id) ON DELETE CASCADE,
    position integer NOT NULL CHECK (position >= 0),
    skill    text NOT NULL CHECK (skill <> ''),
    PRIMARY KEY (power_id, position)
);
-- +goose StatementEnd

-- Subtype table for Stand. The composite FK to powers(id, kind), combined
-- with the CHECK forcing kind = 'STAND', is what makes the inheritance
-- disjoint: a power row of kind DEVIL_FRUIT can never gain a stands row.
-- +goose StatementBegin
CREATE TABLE stands (
    id              uuid PRIMARY KEY,
    kind            power_kind NOT NULL DEFAULT 'STAND' CHECK (kind = 'STAND'),
    attack_power    stand_stat NOT NULL,
    speed           stand_stat NOT NULL,
    attack_range    stand_stat NOT NULL,
    endurance       stand_stat NOT NULL,
    "precision"     stand_stat NOT NULL,
    potential       stand_stat NOT NULL,
    evolves_from_id uuid REFERENCES stands (id) ON DELETE SET NULL,
    CONSTRAINT stands_power_fk FOREIGN KEY (id, kind)
        REFERENCES powers (id, kind) ON DELETE CASCADE,
    CONSTRAINT stands_no_self_evolution CHECK (evolves_from_id <> id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX stands_evolves_from_idx ON stands (evolves_from_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stands;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS power_skills;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS powers;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS power_kind;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS stand_stat;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS power_rarity;
-- +goose StatementEnd
