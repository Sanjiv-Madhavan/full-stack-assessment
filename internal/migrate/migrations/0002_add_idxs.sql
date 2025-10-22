-- +goose Up
CREATE INDEX IF NOT EXISTS idx_tasks_project_status ON tasks (project_id, status, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_tasks_title_like ON tasks (title);

-- +goose Down
DROP INDEX IF EXISTS idx_tasks_project_status;

DROP INDEX IF EXISTS idx_tasks_title_like;