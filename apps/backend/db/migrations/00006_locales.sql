-- +goose Up
-- +goose StatementBegin
CREATE TYPE locale AS ENUM ('en-GB', 'es-ES', 'ca-ES');
-- +goose StatementEnd

-- Per-locale content for a Power's description and skills. `name` stays on
-- `powers` untranslated (proper nouns like "Star Platinum" read the same in
-- every locale), so only description/skills move here. en-GB must always
-- have a row for every power - that is the final link of the fallback
-- chain (ca-ES -> es-ES -> en-GB) - enforced in the application layer, not
-- by a constraint, since goose has no easy way to express "at least one
-- row per group" declaratively.
-- +goose StatementBegin
CREATE TABLE power_translations (
    power_id    uuid   NOT NULL REFERENCES powers (id) ON DELETE CASCADE,
    locale      locale NOT NULL,
    description text   NOT NULL CHECK (description <> ''),
    skills      text[] NOT NULL DEFAULT '{}',
    PRIMARY KEY (power_id, locale)
);
-- +goose StatementEnd

-- Backfill: every existing power's description/skills become its en-GB
-- translation before the source columns are dropped below.
-- +goose StatementBegin
INSERT INTO power_translations (power_id, locale, description, skills)
SELECT p.id,
       'en-GB',
       p.description,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')
FROM powers p
         LEFT JOIN power_skills ps ON ps.power_id = p.id
GROUP BY p.id, p.description;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE power_skills;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE powers DROP COLUMN description;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users ADD COLUMN language locale NOT NULL DEFAULT 'en-GB';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN language;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE powers ADD COLUMN description text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE powers p
SET description = pt.description
FROM power_translations pt
WHERE pt.power_id = p.id AND pt.locale = 'en-GB';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE powers ADD CONSTRAINT powers_description_not_empty CHECK (description <> '');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE power_skills (
    power_id uuid NOT NULL REFERENCES powers (id) ON DELETE CASCADE,
    position integer NOT NULL CHECK (position >= 0),
    skill    text NOT NULL CHECK (skill <> ''),
    PRIMARY KEY (power_id, position)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO power_skills (power_id, position, skill)
SELECT pt.power_id, s.ord - 1, s.skill
FROM power_translations pt
         CROSS JOIN LATERAL unnest(pt.skills) WITH ORDINALITY AS s (skill, ord)
WHERE pt.locale = 'en-GB';
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE power_translations;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE locale;
-- +goose StatementEnd
