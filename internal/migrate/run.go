package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
)

// IMPORTANT: Paths here are RELATIVE TO THIS FILE'S DIRECTORY.
// Your tree has: internal/migrate/migrations/*.sql
// So embed exactly that subfolder:

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Apply runs pending migrations using embedded SQL files.
func Apply(ctx context.Context, db *sql.DB) error {
	// Point Goose at our embedded FS
	goose.SetBaseFS(migrationsFS)

	// Use the sqlite3 dialect (works fine with modernc.org/sqlite driver)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	// VERY IMPORTANT: The dir we pass MUST match the embedded root ("migrations")
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	// Optional: show status in logs for debugging
	if err := goose.StatusContext(ctx, db, "migrations"); err != nil {
		log.Printf("goose status: %v", err)
	}
	return nil
}

func DownOne(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.DownContext(ctx, db, "migrations") // rolls back ONE migration
}

func Reset(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.ResetContext(ctx, db, "migrations") // all the way down, then up
}
