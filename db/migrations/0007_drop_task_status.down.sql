ALTER TABLE tasks ADD COLUMN status TEXT NOT NULL DEFAULT 'todo';
UPDATE tasks SET status = 'todo' WHERE status IS NULL;