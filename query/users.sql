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
ON CONFLICT (email) DO UPDATE SET
    google_id  = COALESCE(users.google_id, EXCLUDED.google_id),
    first_name = COALESCE(NULLIF(users.first_name, ''), EXCLUDED.first_name),
    last_name  = COALESCE(NULLIF(users.last_name, ''), EXCLUDED.last_name),
    avatar_url = COALESCE(users.avatar_url, EXCLUDED.avatar_url)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    first_name,
    last_name,
    role
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;