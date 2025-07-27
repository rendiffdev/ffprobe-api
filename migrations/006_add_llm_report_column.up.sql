-- Migration: Add LLM report column to analyses table
-- This allows storing AI-generated analysis reports alongside technical data

-- Add LLM report column to store AI-generated analysis
ALTER TABLE analyses ADD COLUMN llm_report TEXT;