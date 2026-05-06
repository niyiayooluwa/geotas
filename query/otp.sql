-- name: CreateOTP :one
INSERT INTO otp_codes (
    session_id,
    user_id,
    code,
    expires_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetValidOTP :one
SELECT * FROM otp_codes
WHERE user_id = $1 
  AND session_id = $2 
  AND code = $3 
  AND used = false 
  AND expires_at > NOW();

-- name: MarkOTPUsed :exec
UPDATE otp_codes 
SET used = true 
WHERE id = $1;
