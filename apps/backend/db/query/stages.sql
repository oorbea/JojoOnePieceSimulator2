-- Returns every stage, description resolved for locale via the same
-- fallback-chain LATERAL join power_translations reads use (see stands.sql).
-- name: ListStages :many
SELECT s.id, s.manga, s.position, s.name, s.picture, s.picture_thumb, s.picture_card, s.picture_status,
       s.picture_lqip,
       s.picture_media_id,
       COALESCE(tr.description, '') AS description
FROM stages s
         LEFT JOIN LATERAL (
    SELECT st.description
    FROM stage_translations st
    WHERE st.stage_id = s.id AND st.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], st.locale::text)
    LIMIT 1
    ) tr ON true
ORDER BY s.manga, s.position, s.name;

-- Returns every stage matching the (all-optional) filters, description
-- resolved for locale - same sqlc.narg(...) IS NULL OR ... pattern as
-- FilterStandRows/FilterDevilFruitRows (stands.sql/devil_fruits.sql).
-- name: FilterStageRows :many
SELECT s.id, s.manga, s.position, s.name, s.picture, s.picture_thumb, s.picture_card, s.picture_status,
       s.picture_lqip,
       s.picture_media_id,
       COALESCE(tr.description, '') AS description
FROM stages s
         LEFT JOIN LATERAL (
    SELECT st.description
    FROM stage_translations st
    WHERE st.stage_id = s.id AND st.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], st.locale::text)
    LIMIT 1
    ) tr ON true
WHERE (sqlc.narg('manga')::manga IS NULL OR s.manga = sqlc.narg('manga')::manga)
  AND (sqlc.narg('search')::text IS NULL
       OR s.name ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\'
       OR tr.description ILIKE '%' || sqlc.narg('search')::text || '%' ESCAPE '\')
ORDER BY s.manga, s.position, s.name;

-- name: GetStageByID :one
SELECT s.id, s.manga, s.position, s.name, s.picture, s.picture_thumb, s.picture_card, s.picture_status,
       s.picture_lqip,
       s.picture_media_id,
       COALESCE(tr.description, '') AS description
FROM stages s
         LEFT JOIN LATERAL (
    SELECT st.description
    FROM stage_translations st
    WHERE st.stage_id = s.id AND st.locale::text = ANY (sqlc.arg('locales')::text[])
    ORDER BY array_position(sqlc.arg('locales')::text[], st.locale::text)
    LIMIT 1
    ) tr ON true
WHERE s.id = sqlc.arg('id');

-- name: UpsertStage :one
INSERT INTO stages (id, manga, position, name, picture, picture_thumb, picture_card, picture_status, picture_lqip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
    SET manga          = EXCLUDED.manga,
        position       = EXCLUDED.position,
        name           = EXCLUDED.name,
        picture        = EXCLUDED.picture,
        picture_thumb  = EXCLUDED.picture_thumb,
        picture_card   = EXCLUDED.picture_card,
        picture_status = EXCLUDED.picture_status,
        picture_lqip   = EXCLUDED.picture_lqip,
        updated_at     = now()
RETURNING id, manga, position, name, picture, picture_thumb, picture_card, picture_status, picture_lqip;

-- Updates only a Stage's picture renditions and pipeline status, without
-- touching manga/position/name/translations - same shape as
-- UpdatePowerPicture (stands.sql).
-- name: UpdateStagePicture :exec
UPDATE stages
SET picture          = COALESCE(sqlc.narg('picture')::text, picture),
    picture_thumb    = COALESCE(sqlc.narg('picture_thumb')::text, picture_thumb),
    picture_card     = COALESCE(sqlc.narg('picture_card')::text, picture_card),
    picture_status   = sqlc.arg('picture_status')::picture_status,
    picture_lqip     = COALESCE(sqlc.narg('picture_lqip')::text, picture_lqip),
    picture_media_id = COALESCE(sqlc.narg('picture_media_id')::text, picture_media_id),
    updated_at       = now()
WHERE id = sqlc.arg('id');

-- name: DeleteStageByID :execrows
DELETE FROM stages WHERE id = $1;

-- Upserts a single locale's description for a stage. All three locales are
-- mandatory for a Stage (unlike powers, where only en-GB is) - enforced by
-- the application layer's request validation, not here.
-- name: UpsertStageTranslation :exec
INSERT INTO stage_translations (stage_id, locale, description)
VALUES ($1, $2, $3)
ON CONFLICT (stage_id, locale) DO UPDATE
    SET description = EXCLUDED.description;

-- Deletes translation rows for locales no longer present in an update
-- request - kept for symmetry with DeletePowerTranslations, even though in
-- practice every Stage write always includes all three locales.
-- name: DeleteStageTranslations :exec
DELETE FROM stage_translations WHERE stage_id = $1 AND locale::text = ANY (sqlc.arg('locales')::text[]);

-- Every translation row for a stage, for the admin edit form that needs all
-- locales at once instead of one resolved locale.
-- name: GetStageTranslations :many
SELECT stage_id, locale, description
FROM stage_translations
WHERE stage_id = $1;

-- Sets only a Stage's content-addressed media group id - see
-- UpdatePowerMediaID (stands.sql).
-- name: UpdateStageMediaID :exec
UPDATE stages SET picture_media_id = $1 WHERE id = $2;
