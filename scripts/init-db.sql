-- Database initialization script for PostgreSQL
-- This script is run when the PostgreSQL container starts for the first time

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create development database if it doesn't exist
SELECT 'CREATE DATABASE course_management_dev'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'course_management_dev')\gexec

-- Create test database if it doesn't exist
SELECT 'CREATE DATABASE course_management_test'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'course_management_test')\gexec

-- Connect to development database
\c course_management_dev;

-- Create development user if needed
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_user WHERE usename = 'course_mgmt_user') THEN
        CREATE USER course_mgmt_user WITH PASSWORD 'REPLACE_WITH_SECURE_PASSWORD';
    END IF;
END
$$;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE course_management_dev TO course_mgmt_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO course_mgmt_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO course_mgmt_user;

-- Set default permissions for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO course_mgmt_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO course_mgmt_user;

-- Create indexes for performance (these will be created by GORM migration, but good to have as backup)
-- Course indexes
-- CREATE INDEX IF NOT EXISTS idx_courses_name ON courses(name);
-- CREATE INDEX IF NOT EXISTS idx_courses_location ON courses(city, state);
-- CREATE INDEX IF NOT EXISTS idx_courses_created_by ON courses(created_by);

-- Review indexes
-- CREATE INDEX IF NOT EXISTS idx_reviews_course_user ON course_reviews(course_id, user_id);
-- CREATE INDEX IF NOT EXISTS idx_reviews_user ON course_reviews(user_id);
-- CREATE INDEX IF NOT EXISTS idx_reviews_course ON course_reviews(course_id);

-- Score indexes
-- CREATE INDEX IF NOT EXISTS idx_scores_user_course ON user_course_scores(user_id, course_id);

-- Output success message
SELECT 'Database initialization completed successfully' AS status;