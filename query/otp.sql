-- name: CreateOTP :one
INSERT INTO otp_codes (
    session_id,
    code,
    expires_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetValidOTP :one
SELECT * FROM otp_codes
WHERE session_id = $1 
  AND code = $2 
  AND used = false 
  AND expires_at > NOW();

-- name: GetLatestOTPBySession :one
SELECT * FROM otp_codes
WHERE session_id = $1
AND used = false
AND expires_at > NOW()
ORDER BY issued_at DESC
LIMIT 1;

-- name: InvalidatePreviousOTPs :exec
UPDATE otp_codes
SET used = true
WHERE session_id = $1
AND used = false;

-- name: MarkOTPUsed :exec
UPDATE otp_codes 
SET used = true 
WHERE id = $1;
