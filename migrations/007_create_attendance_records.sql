CREATE TABLE IF NOT EXISTS attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    marked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    method TEXT NOT NULL CHECK (method IN ('qr', 'otp')),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    distance_from_center DOUBLE PRECISION NOT NULL,
    mock_location_detected BOOLEAN NOT NULL DEFAULT FALSE,
    confidence_score DOUBLE PRECISION NOT NULL,
    week_number INT NOT NULL,
    UNIQUE (session_id, user_id)
);

CREATE INDEX idx_attendance_records_session_id ON attendance_records(session_id);
CREATE INDEX idx_attendance_records_user_id ON attendance_records(user_id);