CREATE TABLE IF NOT EXISTS otp_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code TEXT NOT NULL UNIQUE,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_otp_codes_session_id ON otp_codes(session_id);
CREATE INDEX idx_otp_codes_user_id ON otp_codes(user_id);