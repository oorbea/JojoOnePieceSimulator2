-- name: UpsertDevilFruitPower :one
INSERT INTO powers (id, kind, name, description, rarity, picture, picture_thumb, picture_status)
VALUES ($1, 'DEVIL_FRUIT', $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE
    SET name           = EXCLUDED.name,
        description    = EXCLUDED.description,
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

-- Returns the devil fruit matching `id`, along with its ordered skills.
-- Unlike stands, devil fruits have no evolves_from chain to hydrate, so a
-- single flat row (no WITH RECURSIVE, no `matched` column) is enough.
-- name: GetDevilFruitRowByID :one
SELECT p.id,
       p.name,
       p.description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       d.fruit_type,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')::text[] AS skills
FROM devil_fruits d
         JOIN powers p ON p.id = d.id
         LEFT JOIN power_skills ps ON ps.power_id = p.id
WHERE p.id = $1
GROUP BY p.id, p.name, p.description, p.rarity, p.picture, p.picture_thumb, p.picture_status, d.fruit_type;

-- Same shape as GetDevilFruitRowByID, keyed by name instead of id.
-- name: GetDevilFruitRowByName :one
SELECT p.id,
       p.name,
       p.description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       d.fruit_type,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')::text[] AS skills
FROM devil_fruits d
         JOIN powers p ON p.id = d.id
         LEFT JOIN power_skills ps ON ps.power_id = p.id
WHERE p.name = $1
GROUP BY p.id, p.name, p.description, p.rarity, p.picture, p.picture_thumb, p.picture_status, d.fruit_type;

-- Returns every devil fruit in the system.
-- name: ListDevilFruitRows :many
SELECT p.id,
       p.name,
       p.description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       d.fruit_type,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')::text[] AS skills
FROM devil_fruits d
         JOIN powers p ON p.id = d.id
         LEFT JOIN power_skills ps ON ps.power_id = p.id
GROUP BY p.id, p.name, p.description, p.rarity, p.picture, p.picture_thumb, p.picture_status, d.fruit_type
ORDER BY p.name;

-- Returns every devil fruit matching the (all-optional) filters.
-- name: FilterDevilFruitRows :many
SELECT p.id,
       p.name,
       p.description,
       p.rarity,
       p.picture,
       p.picture_thumb,
       p.picture_status,
       d.fruit_type,
       COALESCE(array_agg(ps.skill ORDER BY ps.position) FILTER (WHERE ps.skill IS NOT NULL), '{}')::text[] AS skills
FROM devil_fruits d
         JOIN powers p ON p.id = d.id
         LEFT JOIN power_skills ps ON ps.power_id = p.id
WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR p.rarity = sqlc.narg('rarity')::power_rarity)
  AND (sqlc.narg('fruit_type')::fruit_type IS NULL OR d.fruit_type = sqlc.narg('fruit_type')::fruit_type)
GROUP BY p.id, p.name, p.description, p.rarity, p.picture, p.picture_thumb, p.picture_status, d.fruit_type
ORDER BY p.name;
