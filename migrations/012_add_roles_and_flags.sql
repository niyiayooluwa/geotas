-- Add user_role enum
CREATE TYPE user_role AS ENUM ('student', 'lecturer');

-- Add role column to users
ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'student';

-- Add co_lecturer flag to course_members
ALTER TABLE course_members ADD COLUMN co_lecturer BOOLEAN NOT NULL DEFAULT false;
