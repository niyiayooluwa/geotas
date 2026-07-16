-- Drop user_id from otp_codes so OTPs are session-scoped instead of user-scoped
ALTER TABLE otp_codes DROP COLUMN user_id;
