-- Docker initialization script for PostgreSQL
-- This ensures the database is properly initialized on first startup

-- Create the database if it doesn't exist
SELECT 'CREATE DATABASE ffprobe_api'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ffprobe_api')\gexec

-- Connect to the database
\c ffprobe_api;

-- Grant all privileges to the ffprobe user
GRANT ALL PRIVILEGES ON DATABASE ffprobe_api TO ffprobe;
GRANT ALL ON SCHEMA public TO ffprobe;