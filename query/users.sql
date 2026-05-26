-- name: UpsertGoogleUser :one
INSERT INTO users (
    email,
    google_id,
    first_name,
    last_name,
    avatar_url
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (email) DO UPDATE 
SET google_id = EXCLUDED.google_id, 
    first_name = EXCLUDED.first_name, 
    last_name = EXCLUDED.last_name, 
    avatar_url = EXCLUDED.avatar_url
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;