-- name: UpsertDevilFruitPower :one
INSERT INTO powers (id, kind, name, rarity, picture, picture_thumb, picture_status)
VALUES ($1, 'DEVIL_FRUIT', $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
    SET name           = EXCLUDED.name,
        rarity         = EXCLUDED.rarity,
        picture        = EXCLUDED.picture,
        picture_thumb  = EXCLUDED.picture_thumb,
        picture_status = EXCLUDED.picture_status,
        updated_at     = now()
RETURNING id;

-- name: UpsertDevilFruit :exec
INSERT INTO devil_fruits (id, fruit_type)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET fruit_type = EXCLUDED.fruit_type;

-- name: DeleteDevilFruitByID :execrows
DELETE FROM powers WHERE id = $1 AND kind = 'DEVIL_FRUIT';

-- Returns the devil fruit matching `id`, along with its resolved
-- description/skills. `locales` is the requested locale's fallback chain,
-- most specific first (e.g. ['ca-ES','es-ES','en-GB']) - see
-- GetStandRowsByID in stands.sql for the same LATERAL-join pattern.
-- name: GetDevilFruitRowByID :one
SELECT p.id,
       p.name,
       COALESCE(tr.description, '') AS description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       d.fruit_type,
       COALESCE(tr.skills, '{}')::text[] AS skills
FROM devil_fruits d
         JOIN powers p ON p.id = d.id
         LEFT JOIN LATERAL (
    SELECT pt.description, pt.skills
    FROM power_translations pt
    WHERE pt.power_id = p.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
    LIMIT 1
    ) tr ON true
WHERE p.id = $1;

-- Same shape as GetDevilFruitRowByID, keyed by name instead of id.
-- name: GetDevilFruitRowByName :one
SELECT p.id,
       p.name,
       COALESCE(tr.description, '') AS description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       d.fruit_type,
       COALESCE(tr.skills, '{}')::text[] AS skills
FROM devil_fruits d
         JOIN powers p ON p.id = d.id
         LEFT JOIN LATERAL (
    SELECT pt.description, pt.skills
    FROM power_translations pt
    WHERE pt.power_id = p.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
    LIMIT 1
    ) tr ON true
WHERE p.name = $1;

-- Returns every devil fruit in the system.
-- name: ListDevilFruitRows :many
SELECT p.id,
       p.name,
       COALESCE(tr.description, '') AS description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       d.fruit_type,
       COALESCE(tr.skills, '{}')::text[] AS skills
FROM devil_fruits d
         JOIN powers p ON p.id = d.id
         LEFT JOIN LATERAL (
    SELECT pt.description, pt.skills
    FROM power_translations pt
    WHERE pt.power_id = p.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
    LIMIT 1
    ) tr ON true
ORDER BY p.name;

-- Returns every devil fruit matching the (all-optional) filters.
-- name: FilterDevilFruitRows :many
SELECT p.id,
       p.name,
       COALESCE(tr.description, '') AS description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       d.fruit_type,
       COALESCE(tr.skills, '{}')::text[] AS skills
FROM devil_fruits d
         JOIN powers p ON p.id = d.id
         LEFT JOIN LATERAL (
    SELECT pt.description, pt.skills
    FROM power_translations pt
    WHERE pt.power_id = p.id AND pt.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], pt.locale::text)
    LIMIT 1
    ) tr ON true
WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR p.rarity = sqlc.narg('rarity')::power_rarity)
  AND (sqlc.narg('fruit_type')::fruit_type IS NULL OR d.fruit_type = sqlc.narg('fruit_type')::fruit_type)
  AND (sqlc.narg('search')::text IS NULL
       OR p.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
       OR tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\')
ORDER BY p.name;
