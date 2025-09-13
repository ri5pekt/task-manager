-- Add creator to tasks
ALTER TABLE tasks
  ADD COLUMN created_by UUID NULL REFERENCES users(id) ON DELETE SET NULL;

-- Many-to-many assignees
CREATE TABLE IF NOT EXISTS task_assignees (
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (task_id, user_id)
);

-- Helpful indexes (optional but nice)
CREATE INDEX IF NOT EXISTS idx_task_assignees_user ON task_assignees(user_id);
