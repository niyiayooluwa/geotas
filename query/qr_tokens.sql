-- name: CreateQRToken :one
INSERT INTO qr_tokens (
    session_id,
    token,
    expires_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetValidQRToken :one
SELECT * FROM qr_tokens
WHERE token = $1
AND used = false
AND expires_at > NOW();

-- name: MarkQRTokenUsed :exec
UPDATE qr_tokens
SET used = true
WHERE id = $1;

-- name: InvalidatePreviousTokens :exec
UPDATE qr_tokens
SET used = true
WHERE session_id = $1
AND used = false;