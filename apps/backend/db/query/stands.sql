-- name: UpsertPower :one
INSERT INTO powers (id, kind, name, description, rarity, picture)
VALUES ($1, 'STAND', $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
    SET name         = EXCLUDED.name,
        description  = EXCLUDED.description,
        rarity       = EXCLUDED.rarity,
        picture      = EXCLUDED.picture,
        updated_at   = now()
RETURNING id;

-- name: DeletePowerSkills :exec
DELETE FROM power_skills WHERE power_id = $1;

-- name: InsertPowerSkill :exec
INSERT INTO power_skills (power_id, position, skill)
VALUES ($1, $2, $3);

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
DELETE FROM powers WHERE id = $1;

-- Returns the stand matching `name` (matched = true) plus its full ancestor
-- chain (matched = false), so the caller can hydrate Stand.EvolvesFrom(...)
-- without extra round trips, then discard everything but the matched row.
-- name: GetStandRowsByName :many
WITH RECURSIVE chain AS (SELECT p.id,
                                 p.name,
                                 p.description,
                                 p.rarity,
                                 p.picture,
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
                                 p2.description,
                                 p2.rarity,
                                 p2.picture,
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
                      description,
                      rarity,
                      picture,
                      attack_power,
                      speed,
                      attack_range,
                      endurance,
                      "precision",
                      potential,
                      evolves_from_id,
                      bool_or(matched) AS matched
               FROM chain
               GROUP BY id, name, description, rarity, picture, attack_power, speed, attack_range, endurance,
                        "precision", potential, evolves_from_id)
SELECT d.id,
       d.name,
       d.description,
       d.rarity,
       d.picture,
       d.attack_power,
       d.speed,
       d.attack_range,
       d.endurance,
       d."precision",
       d.potential,
       d.evolves_from_id,
       d.matched,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')::text[] AS skills
FROM dedup d
         LEFT JOIN power_skills ps ON ps.power_id = d.id
GROUP BY d.id, d.name, d.description, d.rarity, d.picture, d.attack_power, d.speed, d.attack_range, d.endurance,
         d."precision", d.potential, d.evolves_from_id, d.matched
ORDER BY d.name;

-- Same shape as GetStandRowsByName, keyed by id instead of name.
-- name: GetStandRowsByID :many
WITH RECURSIVE chain AS (SELECT p.id,
                                 p.name,
                                 p.description,
                                 p.rarity,
                                 p.picture,
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
                                 p2.description,
                                 p2.rarity,
                                 p2.picture,
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
                      description,
                      rarity,
                      picture,
                      attack_power,
                      speed,
                      attack_range,
                      endurance,
                      "precision",
                      potential,
                      evolves_from_id,
                      bool_or(matched) AS matched
               FROM chain
               GROUP BY id, name, description, rarity, picture, attack_power, speed, attack_range, endurance,
                        "precision", potential, evolves_from_id)
SELECT d.id,
       d.name,
       d.description,
       d.rarity,
       d.picture,
       d.attack_power,
       d.speed,
       d.attack_range,
       d.endurance,
       d."precision",
       d.potential,
       d.evolves_from_id,
       d.matched,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')::text[] AS skills
FROM dedup d
         LEFT JOIN power_skills ps ON ps.power_id = d.id
GROUP BY d.id, d.name, d.description, d.rarity, d.picture, d.attack_power, d.speed, d.attack_range, d.endurance,
         d."precision", d.potential, d.evolves_from_id, d.matched
ORDER BY d.name;

-- Returns every stand (matched = true always, no filter applied). Kept in
-- the same shape as GetStandRowsByName/FilterStandRows so all three share
-- one mapper.
-- name: ListStandRows :many
SELECT p.id,
       p.name,
       p.description,
       p.rarity,
       p.picture,
       s.attack_power,
       s.speed,
       s.attack_range,
       s.endurance,
       s."precision",
       s.potential,
       s.evolves_from_id,
       true                                                                                            AS matched,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')::text[] AS skills
FROM stands s
         JOIN powers p ON p.id = s.id
         LEFT JOIN power_skills ps ON ps.power_id = p.id
GROUP BY p.id, p.name, p.description, p.rarity, p.picture, s.attack_power, s.speed, s.attack_range, s.endurance,
         s."precision", s.potential, s.evolves_from_id
ORDER BY p.name;

-- Returns every stand matching the (all-optional) filters, marked
-- matched = true, plus their full ancestor chains (matched = false) needed
-- to hydrate Stand.EvolvesFrom(...). Callers must drop matched = false rows
-- from the final result set.
-- name: FilterStandRows :many
WITH RECURSIVE base AS (SELECT p.id,
                                p.name,
                                p.description,
                                p.rarity,
                                p.picture,
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
                                efp.name = sqlc.narg('evolves_from_name')::text)),
     chain AS (SELECT *
               FROM base
               UNION
               SELECT p2.id,
                      p2.name,
                      p2.description,
                      p2.rarity,
                      p2.picture,
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
                      description,
                      rarity,
                      picture,
                      attack_power,
                      speed,
                      attack_range,
                      endurance,
                      "precision",
                      potential,
                      evolves_from_id,
                      bool_or(matched) AS matched
               FROM chain
               GROUP BY id, name, description, rarity, picture, attack_power, speed, attack_range, endurance,
                        "precision", potential, evolves_from_id)
SELECT d.id,
       d.name,
       d.description,
       d.rarity,
       d.picture,
       d.attack_power,
       d.speed,
       d.attack_range,
       d.endurance,
       d."precision",
       d.potential,
       d.evolves_from_id,
       d.matched,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')::text[] AS skills
FROM dedup d
         LEFT JOIN power_skills ps ON ps.power_id = d.id
GROUP BY d.id, d.name, d.description, d.rarity, d.picture, d.attack_power, d.speed, d.attack_range, d.endurance,
         d."precision", d.potential, d.evolves_from_id, d.matched
ORDER BY d.name;
