-- name: UpsertPower :one
INSERT INTO powers (id, kind, name, rarity, picture, picture_thumb, picture_status)
VALUES ($1, 'STAND', $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
    SET name           = EXCLUDED.name,
        rarity         = EXCLUDED.rarity,
        picture        = EXCLUDED.picture,
        picture_thumb  = EXCLUDED.picture_thumb,
        picture_status = EXCLUDED.picture_status,
        updated_at     = now()
RETURNING id;

-- Upserts a single locale's description/skills for a power. Callers write
-- one row per locale present in the request (en-GB is mandatory, es-ES and
-- ca-ES optional), never touching the other locales' rows.
-- name: UpsertPowerTranslation :exec
INSERT INTO power_translations (power_id, locale, description, skills)
VALUES ($1, $2, $3, $4)
ON CONFLICT (power_id, locale) DO UPDATE
    SET description = EXCLUDED.description,
        skills       = EXCLUDED.skills;

-- Deletes translation rows for locales no longer present in an update
-- request (en-GB can never be deleted this way - callers must not pass it).
-- name: DeletePowerTranslations :exec
DELETE FROM power_translations WHERE power_id = $1 AND locale::text = ANY (sqlc.arg('locales')::text[]);

-- Every translation row for a power, for admin read/write forms that need
-- all locales at once instead of one resolved locale.
-- name: GetPowerTranslations :many
SELECT power_id, locale, description, skills
FROM power_translations
WHERE power_id = $1;

-- name: UpsertStand :exec
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential, evolves_from_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
    SET attack_power    = EXCLUDED.attack_power,
        speed           = EXCLUDED.speed,
        attack_range    = EXCLUDED.attack_range,
        endurance       = EXCLUDED.endurance,
        "precision"     = EXCLUDED.precision,
        potential       = EXCLUDED.potential,
        evolves_from_id = EXCLUDED.evolves_from_id;

-- name: GetStandIDByName :one
SELECT p.id
FROM powers p
         JOIN stands s ON s.id = p.id
WHERE p.name = $1;

-- name: DeleteStandByID :execrows
DELETE FROM powers WHERE id = $1 AND kind = 'STAND';

-- Updates only a Power's picture renditions and pipeline status, without
-- touching name/description/skills/stats - used by the PATCH .../picture
-- handler (status -> PENDING) and by the background compression worker
-- (status -> READY/FAILED). picture/picture_thumb are left untouched when
-- NULL is passed, so the handler can move a row to PENDING without
-- clobbering the renditions currently being served.
-- name: UpdatePowerPicture :exec
UPDATE powers
SET picture        = COALESCE(sqlc.narg('picture')::text, picture),
    picture_thumb  = COALESCE(sqlc.narg('picture_thumb')::text, picture_thumb),
    picture_status = sqlc.arg('picture_status')::picture_status,
    updated_at     = now()
WHERE id = sqlc.arg('id');

-- Returns the stand matching `name` (matched = true) plus its full ancestor
-- chain (matched = false), so the caller can hydrate Stand.EvolvesFrom(...)
-- without extra round trips, then discard everything but the matched row.
-- `locales` is the requested locale's fallback chain, most specific first
-- (e.g. ['ca-ES','es-ES','en-GB']); the LATERAL join below picks the first
-- translation row that exists in that order, so unmatched locales fall
-- back all the way to en-GB without any COALESCE juggling in Go.
-- name: GetStandRowsByName :many
WITH RECURSIVE chain AS (SELECT p.id,
                                 p.name,
                                 p.rarity,
                                 p.picture,
                                 p.picture_thumb,
                                 p.picture_status,
                                 s.attack_power,
                                 s.speed,
                                 s.attack_range,
                                 s.endurance,
                                 s."precision",
                                 s.potential,
                                 s.evolves_from_id,
                                 true AS matched
                          FROM stands s
                                   JOIN powers p ON p.id = s.id
                          WHERE p.name = $1
                          UNION
                          SELECT p2.id,
                                 p2.name,
                                 p2.rarity,
                                 p2.picture,
                                 p2.picture_thumb,
                                 p2.picture_status,
                                 s2.attack_power,
                                 s2.speed,
                                 s2.attack_range,
                                 s2.endurance,
                                 s2."precision",
                                 s2.potential,
                                 s2.evolves_from_id,
                                 false AS matched
                          FROM stands s2
                                   JOIN powers p2 ON p2.id = s2.id
                                   JOIN chain c ON c.evolves_from_id = s2.id),
     dedup AS (SELECT id,
                      name,
                      rarity,
                      picture,
                      picture_thumb,
                      picture_status,
                      attack_power,
                      speed,
                      attack_range,
                      endurance,
                      "precision",
                      potential,
                      evolves_from_id,
                      bool_or(matched) AS matched
               FROM chain
               GROUP BY id, name, rarity, picture, picture_thumb, picture_status, attack_power, speed,
                        attack_range, endurance, "precision", potential, evolves_from_id)
SELECT d.id,
       d.name,
       COALESCE(tr.description, '') AS description,
       d.rarity,
       d.picture,
       d.picture_thumb,
       d.picture_status,
       d.attack_power,
       d.speed,
       d.attack_range,
       d.endurance,
       d."precision",
       d.potential,
       d.evolves_from_id,
       d.matched,
       COALESCE(tr.skills, '{}')::text[] AS skills
FROM dedup d
         LEFT JOIN LATERAL (
    SELECT pt.description, pt.skills
    FROM power_translations pt
    WHERE pt.power_id = d.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
    LIMIT 1
    ) tr ON true
ORDER BY d.name;

-- Same shape as GetStandRowsByName, keyed by id instead of name.
-- name: GetStandRowsByID :many
WITH RECURSIVE chain AS (SELECT p.id,
                                 p.name,
                                 p.rarity,
                                 p.picture,
                                 p.picture_thumb,
                                 p.picture_status,
                                 s.attack_power,
                                 s.speed,
                                 s.attack_range,
                                 s.endurance,
                                 s."precision",
                                 s.potential,
                                 s.evolves_from_id,
                                 true AS matched
                          FROM stands s
                                   JOIN powers p ON p.id = s.id
                          WHERE p.id = $1
                          UNION
                          SELECT p2.id,
                                 p2.name,
                                 p2.rarity,
                                 p2.picture,
                                 p2.picture_thumb,
                                 p2.picture_status,
                                 s2.attack_power,
                                 s2.speed,
                                 s2.attack_range,
                                 s2.endurance,
                                 s2."precision",
                                 s2.potential,
                                 s2.evolves_from_id,
                                 false AS matched
                          FROM stands s2
                                   JOIN powers p2 ON p2.id = s2.id
                                   JOIN chain c ON c.evolves_from_id = s2.id),
     dedup AS (SELECT id,
                      name,
                      rarity,
                      picture,
                      picture_thumb,
                      picture_status,
                      attack_power,
                      speed,
                      attack_range,
                      endurance,
                      "precision",
                      potential,
                      evolves_from_id,
                      bool_or(matched) AS matched
               FROM chain
               GROUP BY id, name, rarity, picture, picture_thumb, picture_status, attack_power, speed,
                        attack_range, endurance, "precision", potential, evolves_from_id)
SELECT d.id,
       d.name,
       COALESCE(tr.description, '') AS description,
       d.rarity,
       d.picture,
       d.picture_thumb,
       d.picture_status,
       d.attack_power,
       d.speed,
       d.attack_range,
       d.endurance,
       d."precision",
       d.potential,
       d.evolves_from_id,
       d.matched,
       COALESCE(tr.skills, '{}')::text[] AS skills
FROM dedup d
         LEFT JOIN LATERAL (
    SELECT pt.description, pt.skills
    FROM power_translations pt
    WHERE pt.power_id = d.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
    LIMIT 1
    ) tr ON true
ORDER BY d.name;

-- Returns every stand (matched = true always, no filter applied). Kept in
-- the same shape as GetStandRowsByName/FilterStandRows so all three share
-- one mapper.
-- name: ListStandRows :many
SELECT p.id,
       p.name,
       COALESCE(tr.description, '') AS description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       s.attack_power,
       s.speed,
       s.attack_range,
       s.endurance,
       s."precision",
       s.potential,
       s.evolves_from_id,
       true                       AS matched,
       COALESCE(tr.skills, '{}')::text[] AS skills
FROM stands s
         JOIN powers p ON p.id = s.id
         LEFT JOIN LATERAL (
    SELECT pt.description, pt.skills
    FROM power_translations pt
    WHERE pt.power_id = p.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
    LIMIT 1
    ) tr ON true
ORDER BY p.name;

-- Returns every stand matching the (all-optional) filters, marked
-- matched = true, plus their full ancestor chains (matched = false) needed
-- to hydrate Stand.EvolvesFrom(...). Callers must drop matched = false rows
-- from the final result set.
-- name: FilterStandRows :many
WITH RECURSIVE base AS (SELECT p.id,
                                p.name,
                                p.rarity,
                                p.picture,
                                p.picture_thumb,
                                p.picture_status,
                                s.attack_power,
                                s.speed,
                                s.attack_range,
                                s.endurance,
                                s."precision",
                                s.potential,
                                s.evolves_from_id,
                                true AS matched
                         FROM stands s
                                  JOIN powers p ON p.id = s.id
                                  LEFT JOIN stands ef ON ef.id = s.evolves_from_id
                                  LEFT JOIN powers efp ON efp.id = ef.id
                                  LEFT JOIN LATERAL (
                             SELECT pt.description
                             FROM power_translations pt
                             WHERE pt.power_id = p.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
                             ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
                             LIMIT 1
                             ) base_tr ON true
                         WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR p.rarity = sqlc.narg('rarity')::power_rarity)
                           AND (sqlc.narg('attack_power')::stand_stat IS NULL OR
                                s.attack_power = sqlc.narg('attack_power')::stand_stat)
                           AND (sqlc.narg('speed')::stand_stat IS NULL OR s.speed = sqlc.narg('speed')::stand_stat)
                           AND (sqlc.narg('attack_range')::stand_stat IS NULL OR
                                s.attack_range = sqlc.narg('attack_range')::stand_stat)
                           AND (sqlc.narg('endurance')::stand_stat IS NULL OR
                                s.endurance = sqlc.narg('endurance')::stand_stat)
                           AND (sqlc.narg('precision')::stand_stat IS NULL OR
                                s."precision" = sqlc.narg('precision')::stand_stat)
                           AND (sqlc.narg('potential')::stand_stat IS NULL OR
                                s.potential = sqlc.narg('potential')::stand_stat)
                           AND (sqlc.narg('evolves_from_name')::text IS NULL OR
                                efp.name = sqlc.narg('evolves_from_name')::text)
                           AND (sqlc.narg('search')::text IS NULL
                                OR p.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
                                OR base_tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\')),
     chain AS (SELECT *
               FROM base
               UNION
               SELECT p2.id,
                      p2.name,
                      p2.rarity,
                      p2.picture,
                      p2.picture_thumb,
                      p2.picture_status,
                      s2.attack_power,
                      s2.speed,
                      s2.attack_range,
                      s2.endurance,
                      s2."precision",
                      s2.potential,
                      s2.evolves_from_id,
                      false AS matched
               FROM stands s2
                        JOIN powers p2 ON p2.id = s2.id
                        JOIN chain c ON c.evolves_from_id = s2.id),
     dedup AS (SELECT id,
                      name,
                      rarity,
                      picture,
                      picture_thumb,
                      picture_status,
                      attack_power,
                      speed,
                      attack_range,
                      endurance,
                      "precision",
                      potential,
                      evolves_from_id,
                      bool_or(matched) AS matched
               FROM chain
               GROUP BY id, name, rarity, picture, picture_thumb, picture_status, attack_power, speed,
                        attack_range, endurance, "precision", potential, evolves_from_id)
SELECT d.id,
       d.name,
       COALESCE(tr.description, '') AS description,
       d.rarity,
       d.picture,
       d.picture_thumb,
       d.picture_status,
       d.attack_power,
       d.speed,
       d.attack_range,
       d.endurance,
       d."precision",
       d.potential,
       d.evolves_from_id,
       d.matched,
       COALESCE(tr.skills, '{}')::text[] AS skills
FROM dedup d
         LEFT JOIN LATERAL (
    SELECT pt.description, pt.skills
    FROM power_translations pt
    WHERE pt.power_id = d.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
    LIMIT 1
    ) tr ON true
ORDER BY d.name;
