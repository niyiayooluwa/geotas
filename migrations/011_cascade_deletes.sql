-- Migration: enforce cascade deletes for clean course/session teardown
-- Run this in your Neon console after existing migrations.

-- attendance_records → sessions (already likely set, but enforce)
ALTER TABLE attendance_records
    DROP CONSTRAINT IF EXISTS attendance_records_session_id_fkey,
    ADD CONSTRAINT attendance_records_session_id_fkey
        FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE;

-- attendance_records → users
ALTER TABLE attendance_records
    DROP CONSTRAINT IF EXISTS attendance_records_user_id_fkey,
    ADD CONSTRAINT attendance_records_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- qr_tokens → sessions
ALTER TABLE qr_tokens
    DROP CONSTRAINT IF EXISTS qr_tokens_session_id_fkey,
    ADD CONSTRAINT qr_tokens_session_id_fkey
        FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE;

-- otp_codes → sessions
ALTER TABLE otp_codes
    DROP CONSTRAINT IF EXISTS otp_codes_session_id_fkey,
    ADD CONSTRAINT otp_codes_session_id_fkey
        FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE;

-- sessions → courses
ALTER TABLE sessions
    DROP CONSTRAINT IF EXISTS sessions_course_id_fkey,
    ADD CONSTRAINT sessions_course_id_fkey
        FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE;

-- course_members → courses
ALTER TABLE course_members
    DROP CONSTRAINT IF EXISTS course_members_course_id_fkey,
    ADD CONSTRAINT course_members_course_id_fkey
        FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE;