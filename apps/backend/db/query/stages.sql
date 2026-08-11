-- name: ListStages :many
SELECT id, manga, position, name FROM stages ORDER BY manga, position, name;

-- name: ListStagesByManga :many
SELECT id, manga, position, name FROM stages WHERE manga = $1 ORDER BY position, name;

-- name: GetStageByID :one
SELECT id, manga, position, name FROM stages WHERE id = $1;

-- name: UpsertStage :one
INSERT INTO stages (id, manga, position, name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
    SET manga      = EXCLUDED.manga,
        position   = EXCLUDED.position,
        name       = EXCLUDED.name,
        updated_at = now()
RETURNING id, manga, position, name;

-- name: DeleteStageByID :execrows
DELETE FROM stages WHERE id = $1;
