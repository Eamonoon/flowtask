-- Migration: 002_add_generation_sessions
-- Description: Add generation_sessions and generated_tasks tables for session-based goal generation
-- Branch: 002-fix-goal-plan-generation
-- Date: 2026-06-23

-- 1. Create generation_sessions table
CREATE TABLE IF NOT EXISTS generation_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learning_goal_id UUID NOT NULL REFERENCES learning_goals(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'generating' CHECK (status IN ('generating', 'completed', 'expired')),
    task_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '24 hours'
);

CREATE INDEX IF NOT EXISTS idx_generation_sessions_learning_goal_id ON generation_sessions(learning_goal_id);
CREATE INDEX IF NOT EXISTS idx_generation_sessions_expires_at ON generation_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_generation_sessions_status ON generation_sessions(status);

-- 2. Create generated_tasks table
CREATE TABLE IF NOT EXISTS generated_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES generation_sessions(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    estimated_duration VARCHAR(50),
    recommended_resources JSONB DEFAULT '[]',
    parent_task_id UUID,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_generated_tasks_session_id ON generated_tasks(session_id);
CREATE INDEX IF NOT EXISTS idx_generated_tasks_parent_task_id ON generated_tasks(parent_task_id);

-- 3. Create function to update learning goal task counts
CREATE OR REPLACE FUNCTION update_learning_goal_task_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE learning_goals
        SET
            task_count = (SELECT COUNT(*) FROM tasks WHERE learning_goal_id = OLD.learning_goal_id),
            completed_task_count = (SELECT COUNT(*) FROM tasks WHERE learning_goal_id = OLD.learning_goal_id AND status = 'done'),
            updated_at = NOW()
        WHERE id = OLD.learning_goal_id;
        RETURN OLD;
    ELSE
        UPDATE learning_goals
        SET
            task_count = (SELECT COUNT(*) FROM tasks WHERE learning_goal_id = NEW.learning_goal_id),
            completed_task_count = (SELECT COUNT(*) FROM tasks WHERE learning_goal_id = NEW.learning_goal_id AND status = 'done'),
            updated_at = NOW()
        WHERE id = NEW.learning_goal_id;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 4. Create trigger for task count synchronization
DROP TRIGGER IF EXISTS trigger_update_task_counts ON tasks;
CREATE TRIGGER trigger_update_task_counts
AFTER INSERT OR UPDATE OR DELETE ON tasks
FOR EACH ROW
EXECUTE FUNCTION update_learning_goal_task_counts();

-- 5. Create function to clean up expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    -- Delete generated tasks from expired sessions
    DELETE FROM generated_tasks
    WHERE session_id IN (
        SELECT id FROM generation_sessions
        WHERE expires_at < NOW() AND status = 'generating'
    );

    -- Mark expired sessions
    UPDATE generation_sessions
    SET status = 'expired'
    WHERE expires_at < NOW() AND status = 'generating';
END;
$$ LANGUAGE plpgsql;

-- 6. Add comment for documentation
COMMENT ON TABLE generation_sessions IS 'Stores generation sessions for learning goal plan generation with 24-hour expiration';
COMMENT ON TABLE generated_tasks IS 'Temporary storage for AI-generated tasks before user confirmation';
COMMENT ON FUNCTION cleanup_expired_sessions() IS 'Cleanup function to be called by scheduled job (every hour)';
