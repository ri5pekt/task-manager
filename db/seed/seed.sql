-- DEV-ONLY SEED. Run after `migrate up` on an empty DB (safe to re-run).

-- 1) One demo user
INSERT INTO users (email, password_hash, name)
VALUES ('demo@example.com', 'x', 'Demo User')
ON CONFLICT (email) DO NOTHING;

-- 2) Demo workspace (idempotent)
WITH w AS (
  INSERT INTO workspaces (name, slug)
  VALUES ('Demo Workspace', 'demo')
  ON CONFLICT (slug) DO UPDATE SET name=EXCLUDED.name
  RETURNING id
),
wu AS (
  SELECT id AS user_id FROM users WHERE email='demo@example.com' LIMIT 1
)
INSERT INTO workspace_members (workspace_id, user_id, role)
SELECT (SELECT id FROM w), (SELECT user_id FROM wu), 'owner'
ON CONFLICT DO NOTHING;

-- 3) Demo board (create once; otherwise pick the existing one)
WITH u AS (
  SELECT id AS user_id FROM users WHERE email='demo@example.com' LIMIT 1
),
maybe_board AS (
  SELECT id FROM boards WHERE name='Demo Board' AND owner_id=(SELECT user_id FROM u) LIMIT 1
),
created AS (
  INSERT INTO boards (name, owner_id, workspace_id)
  SELECT 'Demo Board',
         (SELECT user_id FROM u),
         (SELECT id FROM workspaces WHERE slug='demo' LIMIT 1)
  WHERE NOT EXISTS (SELECT 1 FROM maybe_board)
  RETURNING id
)
SELECT 1; -- no-op to close CTE

-- Resolve board id deterministically
WITH b AS (
  SELECT id FROM boards
  WHERE name='Demo Board'
  ORDER BY created_at ASC
  LIMIT 1
)
-- 4) Ensure lists exist exactly once
INSERT INTO lists (board_id, name, position)
SELECT (SELECT id FROM b), 'To Do', 0
WHERE NOT EXISTS (
  SELECT 1 FROM lists WHERE board_id=(SELECT id FROM b) AND name='To Do'
);

INSERT INTO lists (board_id, name, position)
SELECT (SELECT id FROM b), 'In Progress', 1
WHERE NOT EXISTS (
  SELECT 1 FROM lists WHERE board_id=(SELECT id FROM b) AND name='In Progress'
);

INSERT INTO lists (board_id, name, position)
SELECT (SELECT id FROM b), 'Done', 2
WHERE NOT EXISTS (
  SELECT 1 FROM lists WHERE board_id=(SELECT id FROM b) AND name='Done'
);

-- 5) A few tasks into "To Do" (no status column)
WITH
  b AS (SELECT id FROM boards WHERE name='Demo Board' ORDER BY created_at ASC LIMIT 1),
  l AS (SELECT id FROM lists WHERE board_id=(SELECT id FROM b) AND name='To Do' ORDER BY position ASC LIMIT 1),
  u AS (SELECT id FROM users WHERE email='demo@example.com' LIMIT 1)
INSERT INTO tasks (list_id, title, description, position, created_by, due_date)
SELECT (SELECT id FROM l), 'Wire API → DB', 'Ping DB + version()', 0, (SELECT id FROM u), CURRENT_DATE + 3
WHERE NOT EXISTS (
  SELECT 1 FROM tasks WHERE list_id=(SELECT id FROM l) AND title='Wire API → DB'
);

INSERT INTO tasks (list_id, title, description, position, created_by, due_date)
SELECT (SELECT id FROM l), 'Add migrations', 'users, boards, lists, tasks, comments', 1, (SELECT id FROM u), CURRENT_DATE + 5
WHERE NOT EXISTS (
  SELECT 1 FROM tasks WHERE list_id=(SELECT id FROM l) AND title='Add migrations'
);

INSERT INTO tasks (list_id, title, description, position, created_by, due_date)
SELECT (SELECT id FROM l), 'Vue proxy', 'Vite → Go via /api/*', 2, (SELECT id FROM u), NULL
WHERE NOT EXISTS (
  SELECT 1 FROM tasks WHERE list_id=(SELECT id FROM l) AND title='Vue proxy'
);

-- 6) Assign demo user to first task (idempotent)
WITH t AS (
  SELECT id FROM tasks ORDER BY created_at ASC LIMIT 1
),
u AS (SELECT id FROM users WHERE email='demo@example.com' LIMIT 1)
INSERT INTO task_assignees (task_id, user_id)
SELECT t.id, u.id FROM t, u
ON CONFLICT DO NOTHING;

-- 7) One comment on first task (insert once)
WITH t AS (
  SELECT id FROM tasks ORDER BY created_at ASC LIMIT 1
),
u AS (SELECT id FROM users WHERE email='demo@example.com' LIMIT 1)
INSERT INTO comments (task_id, author_id, body)
SELECT t.id, u.id, 'First!' FROM t, u
WHERE NOT EXISTS (
  SELECT 1 FROM comments WHERE task_id=(SELECT id FROM t) AND author_id=(SELECT id FROM u) AND body='First!'
);

-- 8) Attach all boards to demo workspace if not set
UPDATE boards
SET workspace_id = (SELECT id FROM workspaces WHERE slug='demo' LIMIT 1)
WHERE workspace_id IS NULL;
