-- Add department (required for all)
ALTER TABLE users ADD COLUMN department TEXT NOT NULL DEFAULT '';

-- Add matric_number (students only, unique)
ALTER TABLE users ADD COLUMN matric_number TEXT UNIQUE;

-- Enforce rule: lecturers cannot have a matric number
ALTER TABLE users ADD CONSTRAINT chk_lecturer_no_matric
    CHECK (role != 'lecturer' OR matric_number IS NULL);

-- Remove matriculation_number from course_members
ALTER TABLE course_members DROP COLUMN matriculation_number;
