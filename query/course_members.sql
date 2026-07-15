
-- name: AddCourseMember :one
INSERT INTO course_members (
    course_id,
    user_id,
    role,
    matriculation_number,
    co_lecturer
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;
 
-- name: GetCourseMembersByCourse :many
SELECT 
    cm.course_id,
    cm.user_id,
    cm.role,
    cm.matriculation_number,
    cm.co_lecturer,
    cm.joined_at,
    u.first_name,
    u.last_name,
    u.email,
    u.avatar_url
FROM course_members cm
JOIN users u ON cm.user_id = u.id
WHERE cm.course_id = $1
ORDER BY u.first_name ASC, u.last_name ASC;
 
-- name: GetCourseMember :one
SELECT * FROM course_members
WHERE course_id = $1 AND user_id = $2;
 
-- name: RemoveCourseMember :exec
DELETE FROM course_members
WHERE course_id = $1 AND user_id = $2;