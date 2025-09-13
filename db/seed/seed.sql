-- DEV-ONLY SEED. Run after `migrate up` on an empty DB.

-- 1) One demo user
INSERT INTO users (email, password_hash, name)
VALUES ('demo@example.com', 'x', 'Demo User')
ON CONFLICT (email) DO NOTHING;

-- 2) One board
WITH u AS (
  SELECT id FROM users WHERE email = 'demo@example.com'
),
b AS (
  INSERT INTO boards (name, owner_id)
  SELECT 'Demo Board', u.id FROM u
  RETURNING id
)
-- 3) Three lists
INSERT INTO lists (board_id, name, position)
SELECT id, 'To Do', 0 FROM b
UNION ALL SELECT id, 'In Progress', 1 FROM b
UNION ALL SELECT id, 'Done', 2 FROM b;

-- 4) A few tasks into "To Do"
WITH
  b AS (SELECT id FROM boards WHERE name = 'Demo Board'),
  l AS (
    SELECT id FROM lists WHERE board_id = (SELECT id FROM b) AND name = 'To Do'
  ),
  u AS (SELECT id FROM users WHERE email = 'demo@example.com')
INSERT INTO tasks (list_id, title, description, position, status, created_by, due_date)
SELECT (SELECT id FROM l), 'Wire API → DB', 'Ping DB + version()', 0, 'in_progress', (SELECT id FROM u), CURRENT_DATE + 3
UNION ALL
SELECT (SELECT id FROM l), 'Add migrations', 'users, boards, lists, tasks, comments', 1, 'todo', (SELECT id FROM u), CURRENT_DATE + 5
UNION ALL
SELECT (SELECT id FROM l), 'Vue proxy', 'Vite → Go via /api/*', 2, 'todo', (SELECT id FROM u), NULL;

-- 5) Assign demo user to first task
WITH t AS (
  SELECT id FROM tasks ORDER BY created_at ASC LIMIT 1
),
u AS (SELECT id FROM users WHERE email = 'demo@example.com')
INSERT INTO task_assignees (task_id, user_id)
SELECT t.id, u.id FROM t, u
ON CONFLICT DO NOTHING;

-- 6) One comment on first task
WITH t AS (
  SELECT id FROM tasks ORDER BY created_at ASC LIMIT 1
),
u AS (SELECT id FROM users WHERE email = 'demo@example.com')
INSERT INTO comments (task_id, author_id, body)
SELECT t.id, u.id, 'First!' FROM t, u;
