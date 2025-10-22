-- +goose Up
CREATE INDEX IF NOT EXISTS idx_todos_completed_created ON todos (completed, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_todos_completed_created;