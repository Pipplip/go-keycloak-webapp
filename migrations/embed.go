package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
)

//go:embed app/*.sql
var appFiles embed.FS

func Apply(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin application migrations: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(20261013)"); err != nil {
		return fmt.Errorf("lock application migrations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS app_schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	files, err := fs.Glob(appFiles, "app/*.sql")
	if err != nil {
		return fmt.Errorf("list application migrations: %w", err)
	}
	sort.Strings(files)

	for _, filename := range files {
		version := path.Base(filename)
		var applied bool
		if err := tx.QueryRowContext(
			ctx,
			"SELECT EXISTS (SELECT 1 FROM app_schema_migrations WHERE version = $1)",
			version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			continue
		}

		migration, err := appFiles.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, string(migration)); err != nil {
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(
			ctx,
			"INSERT INTO app_schema_migrations (version) VALUES ($1)",
			version,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit application migrations: %w", err)
	}
	return nil
}
