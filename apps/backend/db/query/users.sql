-- name: UpsertUser :one
INSERT INTO users (id, google_sub, email, username, complete_name, picture, role)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE
    SET email         = EXCLUDED.email,
        username      = EXCLUDED.username,
        complete_name = EXCLUDED.complete_name,
        picture       = EXCLUDED.picture,
        role          = EXCLUDED.role,
        updated_at    = now()
RETURNING id;

-- name: GetUserByID :one
SELECT id, google_sub, email, username, complete_name, picture, role
FROM users
WHERE id = $1;

-- name: GetUserByGoogleSub :one
SELECT id, google_sub, email, username, complete_name, picture, role
FROM users
WHERE google_sub = $1;

-- name: GetUserByEmail :one
SELECT id, google_sub, email, username, complete_name, picture, role
FROM users
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT id, google_sub, email, username, complete_name, picture, role
FROM users
WHERE username = $1;
