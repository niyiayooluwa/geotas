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

-- name: GetCoursesByMember :many
SELECT
    c.id,
    c.owner_id,
    c.title,
    c.code,
    c.department,
    c.invite_code,
    c.confidence_threshold,
    c.created_at,
    cm.role,
    u.matric_number,
    COUNT(all_members.id) FILTER (WHERE all_members.role = 'student') AS student_count
FROM course_members cm
JOIN courses c ON cm.course_id = c.id
JOIN users u ON cm.user_id = u.id
LEFT JOIN course_members all_members ON all_members.course_id = c.id
WHERE cm.user_id = $1
GROUP BY c.id, c.owner_id, c.title, c.code, c.department, c.invite_code, c.confidence_threshold, c.created_at, cm.role, u.matric_number
ORDER BY c.created_at DESC;

-- name: GetCoursesWithStudentCountByOwner :many
SELECT
	c.*,
    COUNT(cm.id) FILTER (WHERE cm.role = 'student') AS student_count
FROM courses c
LEFT JOIN course_members cm ON cm.course_id = c.id
WHERE c.owner_id = $1 OR EXISTS (
    SELECT 1 FROM course_members cm2 
    WHERE cm2.course_id = c.id AND cm2.user_id = $1 AND cm2.co_lecturer = true
)
GROUP BY c.id
ORDER BY c.created_at DESC;

-- name: UpdateCourseInviteCode :one
UPDATE courses
SET invite_code = $2
WHERE id = $1
RETURNING *;

-- name: UpdateCourseConfidenceThreshold :one
UPDATE courses 
SET confidence_threshold = $1 
WHERE id = $2 
RETURNING *;