-- Migration rollback: Remove LLM report column from analyses table

-- Remove LLM report column
ALTER TABLE analyses DROP COLUMN IF EXISTS llm_report;