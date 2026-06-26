-- Migration: 003_add_task_count_to_learning_goals
-- Description: Add task_count and completed_task_count columns to learning_goals table
--              to match the trigger created in 002_add_generation_sessions.sql
-- Date: 2026-06-26

-- Add columns if they don't already exist (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'learning_goals' AND column_name = 'task_count'
    ) THEN
        ALTER TABLE learning_goals ADD COLUMN task_count INTEGER NOT NULL DEFAULT 0;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'learning_goals' AND column_name = 'completed_task_count'
    ) THEN
        ALTER TABLE learning_goals ADD COLUMN completed_task_count INTEGER NOT NULL DEFAULT 0;
    END IF;
END $$;
