-- name: UpsertUser :one
INSERT INTO users (id, google_sub, email, username, complete_name, google_picture, role)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE
    SET email          = EXCLUDED.email,
        username       = EXCLUDED.username,
        complete_name  = EXCLUDED.complete_name,
        google_picture = EXCLUDED.google_picture,
        role           = EXCLUDED.role,
        updated_at     = now()
RETURNING id;

-- name: GetUserByID :one
SELECT id, google_sub, email, username, complete_name, google_picture, role,
       avatar_key, avatar_thumb_key, avatar_status
FROM users
WHERE id = $1;

-- name: GetUserByGoogleSub :one
SELECT id, google_sub, email, username, complete_name, google_picture, role,
       avatar_key, avatar_thumb_key, avatar_status
FROM users
WHERE google_sub = $1;

-- name: GetUserByEmail :one
SELECT id, google_sub, email, username, complete_name, google_picture, role,
       avatar_key, avatar_thumb_key, avatar_status
FROM users
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT id, google_sub, email, username, complete_name, google_picture, role,
       avatar_key, avatar_thumb_key, avatar_status
FROM users
WHERE username = $1;

-- name: UpdateUsername :exec
UPDATE users
SET username   = $1,
    updated_at = now()
WHERE id = $2;

-- Updates only a User's avatar renditions and pipeline status, without
-- touching username/email/role - used by PATCH /users/me/picture (status ->
-- PENDING), by the background compression worker (status -> READY/FAILED),
-- and by DELETE /users/me/picture (key/thumb -> "", status -> NONE).
-- avatar_key/avatar_thumb_key are left untouched when NULL is passed.
-- name: UpdateUserAvatar :exec
UPDATE users
SET avatar_key       = COALESCE(sqlc.narg('avatar_key')::text, avatar_key),
    avatar_thumb_key = COALESCE(sqlc.narg('avatar_thumb_key')::text, avatar_thumb_key),
    avatar_status    = sqlc.arg('avatar_status')::picture_status,
    updated_at       = now()
WHERE id = sqlc.arg('id');

-- name: GetUserAvatarKeys :one
SELECT avatar_key, avatar_thumb_key
FROM users
WHERE id = $1;

-- name: UpdateUserRole :exec
UPDATE users
SET role       = $1,
    updated_at = now()
WHERE id = $2;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: CountAdmins :one
SELECT count(*)
FROM users
WHERE role = 'ADMIN';

-- name: ListUsers :many
SELECT id, google_sub, email, username, complete_name, google_picture, role,
       avatar_key, avatar_thumb_key, avatar_status
FROM users
ORDER BY created_at, id
LIMIT $1 OFFSET $2;
