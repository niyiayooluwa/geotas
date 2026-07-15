-- name: CreateSchedule :one
INSERT INTO schedules (course_id, day_of_week, start_time, end_time, venue)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateSchedule :one
UPDATE schedules
SET day_of_week = $2, start_time = $3, end_time = $4, venue = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetSchedulesByCourse :many
SELECT * FROM schedules
WHERE course_id = $1
ORDER BY day_of_week ASC, start_time ASC;

-- name: GetScheduleByID :one
SELECT * FROM schedules
WHERE id = $1;

-- name: DeleteSchedule :exec
DELETE FROM schedules
WHERE id = $1;
