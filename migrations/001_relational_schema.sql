-- Migration: Convert from JSON storage to relational schema
-- Date: 2025-01-09
-- Description: Replace CourseDB.CourseData JSON field with proper relational structure

-- ===================================================================
-- PHASE 1: CREATE NEW TABLES
-- ===================================================================

-- Enhanced courses table (replace JSON storage)
CREATE TABLE IF NOT EXISTS courses_new (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    description TEXT,
    city VARCHAR(50),
    state VARCHAR(2),
    zip_code VARCHAR(10),
    phone VARCHAR(20),
    website VARCHAR(255),
    overall_rating VARCHAR(1) CHECK (overall_rating IN ('S', 'A', 'B', 'C', 'D', 'F')),
    review TEXT,
    hash VARCHAR(255) UNIQUE NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    created_by INTEGER,
    updated_by INTEGER,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT
);

-- Course holes table (extracted from JSON)
CREATE TABLE IF NOT EXISTS course_holes (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    hole_number INTEGER NOT NULL CHECK (hole_number BETWEEN 1 AND 18),
    par INTEGER CHECK (par BETWEEN 3 AND 6),
    yardage INTEGER CHECK (yardage BETWEEN 0 AND 800),
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    UNIQUE(course_id, hole_number)
);

-- Course rankings table (extracted from JSON)
CREATE TABLE IF NOT EXISTS course_rankings (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL UNIQUE,
    price VARCHAR(10),
    handicap_difficulty INTEGER CHECK (handicap_difficulty BETWEEN 1 AND 10),
    hazard_difficulty INTEGER CHECK (hazard_difficulty BETWEEN 1 AND 10),
    merch VARCHAR(1) CHECK (merch IN ('S', 'A', 'B', 'C', 'D', 'F')),
    condition VARCHAR(1) CHECK (condition IN ('S', 'A', 'B', 'C', 'D', 'F')),
    enjoyment_rating VARCHAR(1) CHECK (enjoyment_rating IN ('S', 'A', 'B', 'C', 'D', 'F')),
    vibe VARCHAR(1) CHECK (vibe IN ('S', 'A', 'B', 'C', 'D', 'F')),
    range_rating VARCHAR(1) CHECK (range_rating IN ('S', 'A', 'B', 'C', 'D', 'F')),
    amenities VARCHAR(1) CHECK (amenities IN ('S', 'A', 'B', 'C', 'D', 'F')),
    glizzies VARCHAR(1) CHECK (glizzies IN ('S', 'A', 'B', 'C', 'D', 'F')),
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT
);

-- User course scores table (extracted from JSON)
CREATE TABLE IF NOT EXISTS user_course_scores_new (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    course_id INTEGER NOT NULL,
    score INTEGER NOT NULL CHECK (score BETWEEN 1 AND 200),
    handicap DECIMAL(4,1) CHECK (handicap BETWEEN -5 AND 40),
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    INDEX(user_id, course_id)
);

-- ===================================================================
-- PHASE 2: CREATE INDEXES FOR PERFORMANCE
-- ===================================================================

-- Course indexes
CREATE INDEX IF NOT EXISTS idx_courses_name ON courses_new(name);
CREATE INDEX IF NOT EXISTS idx_courses_location ON courses_new(city, state);
CREATE INDEX IF NOT EXISTS idx_courses_rating ON courses_new(overall_rating);
CREATE INDEX IF NOT EXISTS idx_courses_created_by ON courses_new(created_by);
CREATE INDEX IF NOT EXISTS idx_courses_name_address ON courses_new(name, address);

-- Hole indexes
CREATE INDEX IF NOT EXISTS idx_holes_course_id ON course_holes(course_id);
CREATE INDEX IF NOT EXISTS idx_holes_par ON course_holes(par);

-- Ranking indexes
CREATE INDEX IF NOT EXISTS idx_rankings_course_id ON course_rankings(course_id);

-- Score indexes
CREATE INDEX IF NOT EXISTS idx_scores_user_course ON user_course_scores_new(user_id, course_id);
CREATE INDEX IF NOT EXISTS idx_scores_course ON user_course_scores_new(course_id);
CREATE INDEX IF NOT EXISTS idx_scores_user ON user_course_scores_new(user_id);

-- ===================================================================
-- PHASE 3: FOREIGN KEY CONSTRAINTS (after data migration)
-- ===================================================================

-- Note: Foreign keys will be added after data migration
-- ALTER TABLE course_holes ADD CONSTRAINT fk_holes_course FOREIGN KEY (course_id) REFERENCES courses_new(id) ON DELETE CASCADE;
-- ALTER TABLE course_rankings ADD CONSTRAINT fk_rankings_course FOREIGN KEY (course_id) REFERENCES courses_new(id) ON DELETE CASCADE;
-- ALTER TABLE user_course_scores_new ADD CONSTRAINT fk_scores_course FOREIGN KEY (course_id) REFERENCES courses_new(id) ON DELETE CASCADE;

-- ===================================================================
-- ROLLBACK INSTRUCTIONS
-- ===================================================================

-- To rollback this migration:
-- DROP TABLE IF EXISTS user_course_scores_new;
-- DROP TABLE IF EXISTS course_rankings;
-- DROP TABLE IF EXISTS course_holes;
-- DROP TABLE IF EXISTS courses_new;