-- Upserts one variant's row for a group - called once per card/thumb/main
-- after a successful transcode. ON CONFLICT keeps scope/content_type/bytes
-- in sync with whatever the latest write says, though in practice a given
-- (group_id, variant) pair is only ever written once: group_id is derived
-- from the bytes themselves, so re-uploading identical content always
-- produces the same group and the same row.
-- name: UpsertMediaObject :exec
INSERT INTO media_objects (group_id, variant, storage_key, content_type, bytes, scope)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (group_id, variant) DO UPDATE
    SET storage_key  = EXCLUDED.storage_key,
        content_type = EXCLUDED.content_type,
        bytes        = EXCLUDED.bytes,
        scope        = EXCLUDED.scope;

-- Every variant row for a group, for the media handler's card->thumb->main
-- fallback ladder (a group backfilled or transcoded before "card" existed
-- may only have thumb/main).
-- name: GetMediaObjectsByGroup :many
SELECT group_id, variant, storage_key, content_type, bytes, scope
FROM media_objects
WHERE group_id = $1;
