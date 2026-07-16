-- Add configurable OTP rotation window to sessions (default 60s — longer than QR since it must be typed)
ALTER TABLE sessions ADD COLUMN otp_rotation_secs INT NOT NULL DEFAULT 60;

-- Compound index for CheckDuplicateDevice query (session_id + device_id lookup was unindexed)
CREATE INDEX idx_attendance_records_device ON attendance_records(session_id, device_id);
