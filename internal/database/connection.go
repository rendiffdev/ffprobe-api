package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rendiffdev/ffprobe-api/internal/config"
	"github.com/rs/zerolog"
)

// DB holds database connections and configuration
type DB struct {
	SQLX   *sqlx.DB
	DB     *sqlx.DB // Alias for SQLX to match repository expectations
	Config *config.Config
	Logger zerolog.Logger
	DbType string // "sqlite" only
}

// New creates a new database connection
func New(cfg *config.Config, logger zerolog.Logger) (*DB, error) {
	var sqlxDB *sqlx.DB
	var err error

	if cfg.DatabaseType != "sqlite" {
		return nil, fmt.Errorf("only SQLite is supported, got: %s", cfg.DatabaseType)
	}

	// Ensure database directory exists
	if err := ensureDatabaseDir(cfg.DatabasePath); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Create SQLite connection
	sqlxDB, err = sqlx.Connect("sqlite3", cfg.DatabasePath+"?_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to create SQLite connection: %w", err)
	}

	// Configure SQLite connection for better performance
	sqlxDB.SetMaxOpenConns(1) // SQLite works best with single connection
	sqlxDB.SetMaxIdleConns(1)
	sqlxDB.SetConnMaxLifetime(time.Hour)

	logger.Info().Str("path", cfg.DatabasePath).Msg("SQLite database connection established")

	// Test the sqlx connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlxDB.PingContext(ctx); err != nil {
		sqlxDB.Close()
		return nil, fmt.Errorf("failed to ping database via sqlx: %w", err)
	}

	db := &DB{
		SQLX:   sqlxDB,
		DB:     sqlxDB, // Set the alias
		Config: cfg,
		Logger: logger,
		DbType: cfg.DatabaseType,
	}

	logger.Info().Str("type", cfg.DatabaseType).Msg("Database connection established successfully")
	return db, nil
}

// ensureDatabaseDir creates the directory for the SQLite database if it doesn't exist
func ensureDatabaseDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." {
		return nil // Current directory, no need to create
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory %s: %w", dir, err)
	}
	return nil
}

// Close closes all database connections
func (db *DB) Close() {
	if db.SQLX != nil {
		db.SQLX.Close()
	}
	db.Logger.Info().Msg("Database connections closed")
}

// Health checks the database connection health
func (db *DB) Health(ctx context.Context) error {
	// Check sqlx connection for SQLite
	if err := db.SQLX.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlx health check failed: %w", err)
	}

	return nil
}

// Stats returns database connection statistics
func (db *DB) Stats() map[string]interface{} {
	sqlxStats := db.SQLX.Stats()

	stats := map[string]interface{}{
		"database_type": db.DbType,
		"sqlx": map[string]interface{}{
			"max_open_connections": sqlxStats.MaxOpenConnections,
			"open_connections":     sqlxStats.OpenConnections,
			"in_use":               sqlxStats.InUse,
			"idle":                 sqlxStats.Idle,
		},
	}

	return stats
}
