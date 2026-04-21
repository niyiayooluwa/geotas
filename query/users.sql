-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    first_name,
    last_name,
    matriculation_number,
    department
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;