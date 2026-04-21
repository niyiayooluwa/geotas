CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    matriculation_number TEXT NOT NULL UNIQUE,
    department TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
)