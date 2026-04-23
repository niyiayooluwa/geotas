-- name: AddCourseMember :one
INSERT INTO course_members (
    course_id,
    user_id,
    role
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetCourseMembersByCourse :many
SELECT * FROM course_members
WHERE course_id = $1;

-- name: GetCoursesByMember :many
SELECT * FROM course_members
WHERE user_id = $1;

-- name: GetCourseMember :one
SELECT * FROM course_members
WHERE course_id = $1 AND user_id = $2;