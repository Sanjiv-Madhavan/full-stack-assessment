-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS todos (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS todos;