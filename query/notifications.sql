-- name: CreateNotification :one
INSERT INTO notifications (user_id, course_id, type, payload)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateNotificationsForCourseMembers :exec
INSERT INTO notifications (user_id, course_id, type, payload)
SELECT user_id, $1, $2, $3
FROM course_members
WHERE course_id = $1;

-- name: GetUnseenNotificationsByUser :many
SELECT * FROM notifications
WHERE user_id = $1 AND seen_at IS NULL AND created_at >= $2
ORDER BY created_at DESC
LIMIT 50;

-- name: MarkNotificationAsSeen :exec
UPDATE notifications
SET seen_at = NOW()
WHERE id = $1 AND user_id = $2;
