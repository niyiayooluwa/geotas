-- name: CreateAttendanceRecord :one
INSERT INTO attendance_records (
    session_id,
    user_id,
    method,
    latitude,
    longitude,
    distance_from_center,
    mock_location_detected,
    confidence_score,
    week_number,
    device_id,
    device_model,
    os_version
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetAttendanceBySession :many
SELECT
    ar.id,
    ar.session_id,
    ar.user_id,
    ar.marked_at,
    ar.method,
    ar.latitude,
    ar.longitude,
    ar.distance_from_center,
    ar.mock_location_detected,
    ar.confidence_score,
    ar.week_number,
    ar.device_id,
    ar.device_model,
    ar.os_version,
    u.first_name,
    u.last_name,
    u.matriculation_number,
    u.department
FROM attendance_records ar
JOIN users u ON ar.user_id = u.id
WHERE ar.session_id = $1
ORDER BY ar.marked_at ASC;

-- name: GetAttendanceByUserAndSession :one
SELECT * FROM attendance_records
WHERE user_id = $1 AND session_id = $2;

-- name: CheckDuplicateDevice :one
SELECT * FROM attendance_records
WHERE session_id = $1 AND device_id = $2
LIMIT 1;

-- name: GetAttendanceByCourse :many
SELECT
    ar.id,
    ar.session_id,
    ar.user_id,
    ar.marked_at,
    ar.method,
    ar.latitude,
    ar.longitude,
    ar.distance_from_center,
    ar.mock_location_detected,
    ar.confidence_score,
    ar.week_number,
    ar.device_id,
    ar.device_model,
    ar.os_version,
    u.first_name,
    u.last_name,
    u.matriculation_number,
    u.department
FROM attendance_records ar
JOIN users u ON ar.user_id = u.id
WHERE ar.session_id IN (
    SELECT id FROM sessions WHERE course_id = $1
)
ORDER BY u.last_name ASC, ar.week_number ASC;