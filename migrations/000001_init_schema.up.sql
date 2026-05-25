-- Users
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Case-insensitive unique email without requiring the citext extension.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (LOWER(email));

-- Projects
CREATE TABLE IF NOT EXISTS projects (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    owner_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects (owner_id);

-- Project members (M:N users<->projects)
CREATE TABLE IF NOT EXISTS project_members (
    id         BIGSERIAL PRIMARY KEY,
    project_id BIGINT      NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users (id)    ON DELETE CASCADE,
    role       TEXT        NOT NULL DEFAULT 'member'
                            CHECK (role IN ('owner', 'admin', 'member')),
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_members_user_id    ON project_members (user_id);
CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members (project_id);

-- Tasks
CREATE TABLE IF NOT EXISTS tasks (
    id          BIGSERIAL PRIMARY KEY,
    project_id  BIGINT      NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    creator_id  BIGINT      NOT NULL REFERENCES users (id)    ON DELETE RESTRICT,
    assignee_id BIGINT               REFERENCES users (id)    ON DELETE SET NULL,
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    status      TEXT        NOT NULL DEFAULT 'todo'
                            CHECK (status IN ('todo', 'in_progress', 'review', 'done')),
    priority    TEXT        NOT NULL DEFAULT 'medium'
                            CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    due_date    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_project_id  ON tasks (project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks (assignee_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status      ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority    ON tasks (priority);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at  ON tasks (created_at DESC);

-- Task comments
CREATE TABLE IF NOT EXISTS task_comments (
    id         BIGSERIAL PRIMARY KEY,
    task_id    BIGINT      NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    author_id  BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_comments_task_id ON task_comments (task_id);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type       TEXT        NOT NULL,
    message    TEXT        NOT NULL,
    is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications (user_id, is_read);

-- Trigger to keep tasks.updated_at fresh
CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_tasks_updated_at ON tasks;
CREATE TRIGGER trg_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION touch_updated_at();
