-- name: CreateCourse :one
INSERT INTO courses (
    owner_id,
    title,
    code,
    department,
    invite_code
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetCourseByID :one
SELECT * FROM courses
WHERE id = $1;

-- name: GetCoursesByOwner :many
SELECT * FROM courses
WHERE owner_id = $1;

-- name: GetCourseByCode :one
SELECT * FROM courses
WHERE code = $1;

-- name: GetCourseByInviteCode :one
SELECT * FROM courses
WHERE invite_code = $1;

-- name: DeleteCourse :exec
DELETE FROM courses
WHERE id = $1;