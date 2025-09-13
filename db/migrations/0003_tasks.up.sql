CREATE TABLE IF NOT EXISTS tasks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  position INT NOT NULL DEFAULT 0,                    -- order within a list (left→right top→bottom)
  status TEXT NOT NULL DEFAULT 'todo',                -- we'll normalize later if needed
  assignee_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
  due_date DATE NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- quick lookups & ordered pulls
CREATE INDEX IF NOT EXISTS idx_tasks_list ON tasks(list_id);
CREATE INDEX IF NOT EXISTS idx_tasks_list_position ON tasks(list_id, position);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee ON tasks(assignee_id);
