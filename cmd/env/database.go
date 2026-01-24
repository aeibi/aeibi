package env

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"aeibi/internal/config"
	"aeibi/internal/repository/db"

	_ "modernc.org/sqlite"
)

// InitDB opens the database connection and pings it to ensure readiness.
func InitDB(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	dbConn, err := sql.Open("sqlite", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	dbConn.SetMaxOpenConns(1)
	dbConn.SetMaxIdleConns(1)

	if err := dbConn.PingContext(ctx); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return dbConn, nil
}

// MigrateDB runs the database migrations defined in the configured directory.
func MigrateDB(cfg config.DatabaseConfig) error {
	migrationPath, err := filepath.Abs(cfg.MigrationsDir)
	if err != nil {
		return fmt.Errorf("resolve migrations dir: %w", err)
	}

	migrationDSN := strings.TrimPrefix(cfg.DSN, "file:")
	if err := db.Migration(fmt.Sprintf("file://%s", migrationPath), fmt.Sprintf("sqlite://%s", migrationDSN)); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
