-- Migration: Add walkability column to course reviews
-- Date: 2025-07-10
-- Description: Add walkability metric to course review system

-- Add walkability column to course_review_dbs table
ALTER TABLE course_review_dbs ADD COLUMN walkability VARCHAR(1) CHECK (walkability IN ('S', 'A', 'B', 'C', 'D', 'F'));

-- Add walkability column to course_rankings table (for legacy support)
ALTER TABLE course_rankings ADD COLUMN walkability VARCHAR(1) CHECK (walkability IN ('S', 'A', 'B', 'C', 'D', 'F'));

-- ===================================================================
-- ROLLBACK INSTRUCTIONS
-- ===================================================================

-- To rollback this migration:
-- ALTER TABLE course_review_dbs DROP COLUMN walkability;
-- ALTER TABLE course_rankings DROP COLUMN walkability;