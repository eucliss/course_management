-- Data Migration: Extract JSON data to relational structure
-- Date: 2025-01-09
-- Description: Migrate existing course data from JSON to relational tables

-- ===================================================================
-- PHASE 1: MIGRATE COURSE DATA
-- ===================================================================

-- Migrate basic course information
INSERT INTO courses_new (
    id, name, address, description, hash, latitude, longitude, 
    created_by, updated_by, created_at, updated_at,
    overall_rating, review
)
SELECT 
    cd.id,
    cd.name,
    cd.address,
    -- Extract description from JSON
    COALESCE(cd.course_data::json->>'description', ''),
    cd.hash,
    cd.latitude,
    cd.longitude,
    cd.created_by,
    cd.updated_by,
    cd.created_at,
    cd.updated_at,
    -- Extract overall rating from JSON
    COALESCE(cd.course_data::json->>'overallRating', ''),
    -- Extract review from JSON  
    COALESCE(cd.course_data::json->>'review', '')
FROM course_dbs cd
WHERE cd.course_data IS NOT NULL AND cd.course_data != '';

-- ===================================================================
-- PHASE 2: MIGRATE HOLES DATA
-- ===================================================================

-- Extract holes from JSON and insert into course_holes table
INSERT INTO course_holes (course_id, hole_number, par, yardage, description)
SELECT 
    cd.id as course_id,
    (hole_data->>'number')::int as hole_number,
    COALESCE((hole_data->>'par')::int, 4) as par,
    COALESCE((hole_data->>'yardage')::int, 0) as yardage,
    COALESCE(hole_data->>'description', '') as description
FROM course_dbs cd,
     json_array_elements(cd.course_data::json->'holes') as hole_data
WHERE cd.course_data IS NOT NULL 
  AND cd.course_data != ''
  AND cd.course_data::json->'holes' IS NOT NULL
  AND json_array_length(cd.course_data::json->'holes') > 0;

-- ===================================================================
-- PHASE 3: MIGRATE RANKINGS DATA
-- ===================================================================

-- Extract rankings from JSON and insert into course_rankings table
INSERT INTO course_rankings (
    course_id, price, handicap_difficulty, hazard_difficulty,
    merch, condition, enjoyment_rating, vibe, range_rating, amenities, glizzies
)
SELECT 
    cd.id as course_id,
    COALESCE(cd.course_data::json->'ranks'->>'price', '') as price,
    COALESCE((cd.course_data::json->'ranks'->>'handicapDifficulty')::int, 5) as handicap_difficulty,
    COALESCE((cd.course_data::json->'ranks'->>'hazardDifficulty')::int, 5) as hazard_difficulty,
    COALESCE(cd.course_data::json->'ranks'->>'merch', '') as merch,
    COALESCE(cd.course_data::json->'ranks'->>'condition', '') as condition,
    COALESCE(cd.course_data::json->'ranks'->>'enjoymentRating', '') as enjoyment_rating,
    COALESCE(cd.course_data::json->'ranks'->>'vibe', '') as vibe,
    COALESCE(cd.course_data::json->'ranks'->>'range', '') as range_rating,
    COALESCE(cd.course_data::json->'ranks'->>'amenities', '') as amenities,
    COALESCE(cd.course_data::json->'ranks'->>'glizzies', '') as glizzies
FROM course_dbs cd
WHERE cd.course_data IS NOT NULL 
  AND cd.course_data != ''
  AND cd.course_data::json->'ranks' IS NOT NULL;

-- ===================================================================
-- PHASE 4: MIGRATE SCORES DATA
-- ===================================================================

-- Extract scores from JSON and insert into user_course_scores_new table
-- Note: Since scores in JSON don't have user_id, we'll need to handle this differently
-- For now, we'll create placeholder entries that can be updated later

INSERT INTO user_course_scores_new (user_id, course_id, score, handicap)
SELECT 
    COALESCE(cd.created_by, 1) as user_id, -- Use course creator as default user
    cd.id as course_id,
    (score_data->>'score')::int as score,
    COALESCE((score_data->>'handicap')::decimal, 0.0) as handicap
FROM course_dbs cd,
     json_array_elements(cd.course_data::json->'scores') as score_data
WHERE cd.course_data IS NOT NULL 
  AND cd.course_data != ''
  AND cd.course_data::json->'scores' IS NOT NULL
  AND json_array_length(cd.course_data::json->'scores') > 0;

-- ===================================================================
-- PHASE 5: VERIFICATION QUERIES
-- ===================================================================

-- Verify migration counts
-- SELECT 'Courses migrated:' as info, COUNT(*) as count FROM courses_new
-- UNION ALL
-- SELECT 'Holes migrated:', COUNT(*) FROM course_holes
-- UNION ALL  
-- SELECT 'Rankings migrated:', COUNT(*) FROM course_rankings
-- UNION ALL
-- SELECT 'Scores migrated:', COUNT(*) FROM user_course_scores_new;

-- ===================================================================
-- PHASE 6: FOREIGN KEY CONSTRAINTS (run after verification)
-- ===================================================================

-- Add foreign key constraints after data is migrated and verified
-- ALTER TABLE course_holes ADD CONSTRAINT fk_holes_course 
--     FOREIGN KEY (course_id) REFERENCES courses_new(id) ON DELETE CASCADE;

-- ALTER TABLE course_rankings ADD CONSTRAINT fk_rankings_course 
--     FOREIGN KEY (course_id) REFERENCES courses_new(id) ON DELETE CASCADE;

-- ALTER TABLE user_course_scores_new ADD CONSTRAINT fk_scores_course 
--     FOREIGN KEY (course_id) REFERENCES courses_new(id) ON DELETE CASCADE;