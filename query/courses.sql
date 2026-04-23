-- name: CreateCourse :one
INSERT INTO courses (
    owner_id,
    title,
    code,
    department
) VALUES (
    $1, $2, $3, $4
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