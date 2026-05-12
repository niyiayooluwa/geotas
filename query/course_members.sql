-- name: AddCourseMember :one
INSERT INTO course_members (
    course_id,
    user_id,
    role,
    matriculation_number
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetCourseMembersByCourse :many
SELECT * FROM course_members
WHERE course_id = $1;

-- name: GetCourseMember :one
SELECT * FROM course_members
WHERE course_id = $1 AND user_id = $2;