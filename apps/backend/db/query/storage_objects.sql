-- name: RecordStorageObject :exec
INSERT INTO storage_objects (key, provider, bytes)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE
    SET provider = EXCLUDED.provider,
        bytes    = EXCLUDED.bytes;

-- name: ForgetStorageObject :exec
DELETE FROM storage_objects
WHERE key = $1;

-- name: GetStorageObject :one
SELECT key, provider, bytes
FROM storage_objects
WHERE key = $1;

-- name: SumStorageUsage :many
SELECT provider, COALESCE(SUM(bytes), 0)::bigint AS bytes, count(*)::bigint AS objects
FROM storage_objects
GROUP BY provider;

-- Used by the reconciler to swap a provider's whole inventory in one
-- transaction (see repositories.StorageLedger.Replace).
-- name: DeleteStorageObjectsByProvider :exec
DELETE FROM storage_objects
WHERE provider = $1;

-- name: ListStorageObjectsByProvider :many
SELECT key, provider, bytes
FROM storage_objects
WHERE provider = $1;
