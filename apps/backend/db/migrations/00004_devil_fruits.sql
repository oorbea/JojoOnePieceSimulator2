-- +goose Up
-- +goose StatementBegin
CREATE TYPE fruit_type AS ENUM ('PARAMECIA', 'ZOAN', 'LOGIA', 'SPECIAL_PARAMECIA', 'ANCIENT_ZOAN', 'MYTHICAL_ZOAN');
-- +goose StatementEnd

-- Subtype table for DevilFruit, same class-table-inheritance shape as
-- stands: the composite FK to powers(id, kind), combined with the CHECK
-- forcing kind = 'DEVIL_FRUIT', makes the inheritance disjoint - a power row
-- of kind STAND can never gain a devil_fruits row.
-- +goose StatementBegin
CREATE TABLE devil_fruits (
    id         uuid PRIMARY KEY,
    kind       power_kind NOT NULL DEFAULT 'DEVIL_FRUIT' CHECK (kind = 'DEVIL_FRUIT'),
    fruit_type fruit_type NOT NULL,
    CONSTRAINT devil_fruits_power_fk FOREIGN KEY (id, kind)
        REFERENCES powers (id, kind) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX devil_fruits_fruit_type_idx ON devil_fruits (fruit_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS devil_fruits;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS fruit_type;
-- +goose StatementEnd
