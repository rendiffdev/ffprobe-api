-- Migration: Drop video_comparisons table and related structures
-- Version: 004_create_video_comparisons.down.sql

-- Drop tables
DROP TABLE IF EXISTS video_comparisons CASCADE;

-- Drop enums
DROP TYPE IF EXISTS comparison_type CASCADE;
DROP TYPE IF EXISTS comparison_status CASCADE;