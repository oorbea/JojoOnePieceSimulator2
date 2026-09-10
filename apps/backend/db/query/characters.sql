-- name: UpsertCharacter :one
INSERT INTO characters (id, manga, name, rarity, picture, picture_thumb, picture_card, picture_status, picture_lqip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
    SET name           = EXCLUDED.name,
        rarity         = EXCLUDED.rarity,
        picture        = EXCLUDED.picture,
        picture_thumb  = EXCLUDED.picture_thumb,
        picture_card   = EXCLUDED.picture_card,
        picture_status = EXCLUDED.picture_status,
        picture_lqip   = EXCLUDED.picture_lqip,
        updated_at     = now()
RETURNING id;

-- Upserts a single locale's description for a character. Only en-GB is
-- mandatory (like powers, unlike stages) - enforced by the application
-- layer's request validation, not here.
-- name: UpsertCharacterTranslation :exec
INSERT INTO character_translations (character_id, locale, description)
VALUES ($1, $2, $3)
ON CONFLICT (character_id, locale) DO UPDATE
    SET description = EXCLUDED.description;

-- Deletes translation rows for locales no longer present in an update
-- request (en-GB can never be deleted this way - callers must not pass it).
-- name: DeleteCharacterTranslations :exec
DELETE FROM character_translations WHERE character_id = $1 AND locale::text = ANY (sqlc.arg('locales')::text[]);

-- Every translation row for a character, for admin read/write forms that
-- need all locales at once instead of one resolved locale.
-- name: GetCharacterTranslations :many
SELECT character_id, locale, description
FROM character_translations
WHERE character_id = $1;

-- Updates only a Character's picture renditions and pipeline status,
-- without touching name/rarity/translations/stats - same shape as
-- UpdatePowerPicture (stands.sql).
-- name: UpdateCharacterPicture :exec
UPDATE characters
SET picture          = COALESCE(sqlc.narg('picture')::text, picture),
    picture_thumb    = COALESCE(sqlc.narg('picture_thumb')::text, picture_thumb),
    picture_card     = COALESCE(sqlc.narg('picture_card')::text, picture_card),
    picture_status   = sqlc.arg('picture_status')::picture_status,
    picture_lqip     = COALESCE(sqlc.narg('picture_lqip')::text, picture_lqip),
    picture_media_id = COALESCE(sqlc.narg('picture_media_id')::text, picture_media_id),
    updated_at       = now()
WHERE id = sqlc.arg('id');

-- Sets only a Character's content-addressed media group id - see
-- UpdatePowerMediaID (stands.sql).
-- name: UpdateCharacterMediaID :exec
UPDATE characters SET picture_media_id = $1 WHERE id = $2;

-- name: DeleteCharacterByID :execrows
DELETE FROM characters WHERE id = $1;

-- name: UpsertJojoCharacter :exec
INSERT INTO jojo_characters (id, hamon, spin, battle_iq)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
    SET hamon     = EXCLUDED.hamon,
        spin      = EXCLUDED.spin,
        battle_iq = EXCLUDED.battle_iq;

-- name: GetJojoCharacterRowByID :one
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       j.hamon, j.spin, j.battle_iq
FROM jojo_characters j
         JOIN characters c ON c.id = j.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE c.id = $1;

-- name: GetJojoCharacterRowByName :one
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       j.hamon, j.spin, j.battle_iq
FROM jojo_characters j
         JOIN characters c ON c.id = j.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE c.name = $1;

-- name: ListJojoCharacterRows :many
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       j.hamon, j.spin, j.battle_iq
FROM jojo_characters j
         JOIN characters c ON c.id = j.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
ORDER BY c.name;

-- name: FilterJojoCharacterRows :many
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       j.hamon, j.spin, j.battle_iq
FROM jojo_characters j
         JOIN characters c ON c.id = j.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR c.rarity = sqlc.narg('rarity')::power_rarity)
  AND (sqlc.narg('hamon')::hamon_level IS NULL OR j.hamon = sqlc.narg('hamon')::hamon_level)
  AND (sqlc.narg('spin')::spin_level IS NULL OR j.spin = sqlc.narg('spin')::spin_level)
  AND (sqlc.narg('battle_iq')::smallint IS NULL OR j.battle_iq = sqlc.narg('battle_iq')::smallint)
  AND (sqlc.narg('search')::text IS NULL
       OR c.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
       OR tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\')
ORDER BY c.name;

-- Keyset-paginated counterpart of FilterJojoCharacterRows - same contract as
-- PageDevilFruitRows (devil_fruits.sql): Go passes page_limit = limit + 1
-- and detects HasMore from the extra row.
-- name: PageJojoCharacterRows :many
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       j.hamon, j.spin, j.battle_iq
FROM jojo_characters j
         JOIN characters c ON c.id = j.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR c.rarity = sqlc.narg('rarity')::power_rarity)
  AND (sqlc.narg('hamon')::hamon_level IS NULL OR j.hamon = sqlc.narg('hamon')::hamon_level)
  AND (sqlc.narg('spin')::spin_level IS NULL OR j.spin = sqlc.narg('spin')::spin_level)
  AND (sqlc.narg('battle_iq')::smallint IS NULL OR j.battle_iq = sqlc.narg('battle_iq')::smallint)
  AND (sqlc.narg('search')::text IS NULL
       OR c.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
       OR tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\')
  AND (sqlc.narg('after_name')::text IS NULL OR c.name > sqlc.narg('after_name')::text)
ORDER BY c.name
LIMIT sqlc.arg('page_limit')::int;

-- name: CountJojoCharacterRows :one
SELECT count(*)
FROM jojo_characters j
         JOIN characters c ON c.id = j.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR c.rarity = sqlc.narg('rarity')::power_rarity)
  AND (sqlc.narg('hamon')::hamon_level IS NULL OR j.hamon = sqlc.narg('hamon')::hamon_level)
  AND (sqlc.narg('spin')::spin_level IS NULL OR j.spin = sqlc.narg('spin')::spin_level)
  AND (sqlc.narg('battle_iq')::smallint IS NULL OR j.battle_iq = sqlc.narg('battle_iq')::smallint)
  AND (sqlc.narg('search')::text IS NULL
       OR c.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
       OR tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\');

-- name: UpsertOnePieceCharacter :exec
INSERT INTO one_piece_characters (id, physical_form, armament_haki, observation_haki, conqueror_haki, fruit_mastery)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
    SET physical_form    = EXCLUDED.physical_form,
        armament_haki    = EXCLUDED.armament_haki,
        observation_haki = EXCLUDED.observation_haki,
        conqueror_haki   = EXCLUDED.conqueror_haki,
        fruit_mastery    = EXCLUDED.fruit_mastery;

-- name: GetOnePieceCharacterRowByID :one
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       o.physical_form, o.armament_haki, o.observation_haki, o.conqueror_haki, o.fruit_mastery
FROM one_piece_characters o
         JOIN characters c ON c.id = o.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE c.id = $1;

-- name: GetOnePieceCharacterRowByName :one
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       o.physical_form, o.armament_haki, o.observation_haki, o.conqueror_haki, o.fruit_mastery
FROM one_piece_characters o
         JOIN characters c ON c.id = o.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE c.name = $1;

-- name: ListOnePieceCharacterRows :many
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       o.physical_form, o.armament_haki, o.observation_haki, o.conqueror_haki, o.fruit_mastery
FROM one_piece_characters o
         JOIN characters c ON c.id = o.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
ORDER BY c.name;

-- name: FilterOnePieceCharacterRows :many
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       o.physical_form, o.armament_haki, o.observation_haki, o.conqueror_haki, o.fruit_mastery
FROM one_piece_characters o
         JOIN characters c ON c.id = o.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR c.rarity = sqlc.narg('rarity')::power_rarity)
  AND (sqlc.narg('physical_form')::physical_form IS NULL OR o.physical_form = sqlc.narg('physical_form')::physical_form)
  AND (sqlc.narg('armament_haki')::haki_level IS NULL OR o.armament_haki = sqlc.narg('armament_haki')::haki_level)
  AND (sqlc.narg('observation_haki')::haki_level IS NULL OR o.observation_haki = sqlc.narg('observation_haki')::haki_level)
  AND (sqlc.narg('conqueror_haki')::haki_level IS NULL OR o.conqueror_haki = sqlc.narg('conqueror_haki')::haki_level)
  AND (sqlc.narg('fruit_mastery')::fruit_mastery IS NULL OR o.fruit_mastery = sqlc.narg('fruit_mastery')::fruit_mastery)
  AND (sqlc.narg('search')::text IS NULL
       OR c.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
       OR tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\')
ORDER BY c.name;

-- name: PageOnePieceCharacterRows :many
SELECT c.id, c.name, COALESCE(tr.description, '') AS description, c.rarity,
       c.picture, c.picture_thumb, c.picture_card, c.picture_status, c.picture_lqip, c.picture_media_id,
       o.physical_form, o.armament_haki, o.observation_haki, o.conqueror_haki, o.fruit_mastery
FROM one_piece_characters o
         JOIN characters c ON c.id = o.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR c.rarity = sqlc.narg('rarity')::power_rarity)
  AND (sqlc.narg('physical_form')::physical_form IS NULL OR o.physical_form = sqlc.narg('physical_form')::physical_form)
  AND (sqlc.narg('armament_haki')::haki_level IS NULL OR o.armament_haki = sqlc.narg('armament_haki')::haki_level)
  AND (sqlc.narg('observation_haki')::haki_level IS NULL OR o.observation_haki = sqlc.narg('observation_haki')::haki_level)
  AND (sqlc.narg('conqueror_haki')::haki_level IS NULL OR o.conqueror_haki = sqlc.narg('conqueror_haki')::haki_level)
  AND (sqlc.narg('fruit_mastery')::fruit_mastery IS NULL OR o.fruit_mastery = sqlc.narg('fruit_mastery')::fruit_mastery)
  AND (sqlc.narg('search')::text IS NULL
       OR c.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
       OR tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\')
  AND (sqlc.narg('after_name')::text IS NULL OR c.name > sqlc.narg('after_name')::text)
ORDER BY c.name
LIMIT sqlc.arg('page_limit')::int;

-- name: CountOnePieceCharacterRows :one
SELECT count(*)
FROM one_piece_characters o
         JOIN characters c ON c.id = o.id
         LEFT JOIN LATERAL (
    SELECT ct.description
    FROM character_translations ct
    WHERE ct.character_id = c.id AND ct.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], ct.locale::text)
    LIMIT 1
    ) tr ON true
WHERE (sqlc.narg('rarity')::power_rarity IS NULL OR c.rarity = sqlc.narg('rarity')::power_rarity)
  AND (sqlc.narg('physical_form')::physical_form IS NULL OR o.physical_form = sqlc.narg('physical_form')::physical_form)
  AND (sqlc.narg('armament_haki')::haki_level IS NULL OR o.armament_haki = sqlc.narg('armament_haki')::haki_level)
  AND (sqlc.narg('observation_haki')::haki_level IS NULL OR o.observation_haki = sqlc.narg('observation_haki')::haki_level)
  AND (sqlc.narg('conqueror_haki')::haki_level IS NULL OR o.conqueror_haki = sqlc.narg('conqueror_haki')::haki_level)
  AND (sqlc.narg('fruit_mastery')::fruit_mastery IS NULL OR o.fruit_mastery = sqlc.narg('fruit_mastery')::fruit_mastery)
  AND (sqlc.narg('search')::text IS NULL
       OR c.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
       OR tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\');
