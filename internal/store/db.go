package store

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite" // registers "sqlite" driver
)

const schema = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS todos (
  id         TEXT PRIMARY KEY,
  title      TEXT NOT NULL,
  completed  INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);

-- helpful composite index if you later filter/sort
CREATE INDEX IF NOT EXISTS idx_todos_completed_created ON todos(completed, created_at);
`

// InMemory opens a single-process in-memory SQLite with shared cache.
// NOTE: state is per-process and disappears on restart; perfect for this assessment.
// Use file-based (e.g., "file:todo.db?_fk=1") later for persistence.
func InMemory(ctx context.Context) (*sql.DB, error) {
	// shared in-memory DB (lives as long as THIS *process* has at least one open conn)
	// docs: https://www.sqlite.org/inmemorydb.html
	dsn := "file:todo?mode=memory&cache=shared&_fk=1" // enable FKs
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// limit to a single connection so ":memory: shared" definitely sees the same db
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.ExecContext(ctx, schema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
