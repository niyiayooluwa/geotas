-- name: CreateSession :one
INSERT INTO sessions (
    course_id,
    created_by,
    title,
    week_number,
    latitude,
    longitude,
    radius_meters,
    qr_rotation_secs,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, 'active'
)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions
WHERE id = $1;

-- name: GetSessionsByCourse :many
SELECT * FROM sessions
WHERE course_id = $1
ORDER BY started_at DESC;

-- name: GetActiveSessionByCourse :one
SELECT * FROM sessions
WHERE course_id = $1 AND status = 'active';

-- name: CloseSession :one
UPDATE sessions
SET status = 'closed', closed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;