CREATE TABLE IF NOT EXISTS courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    department TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)

CREATE INDEX idx_courses_owner_id ON courses(owner_id);