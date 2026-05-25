DROP TRIGGER IF EXISTS trg_tasks_updated_at ON tasks;
DROP FUNCTION IF EXISTS touch_updated_at;

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
