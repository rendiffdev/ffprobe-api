package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/config"
)

// DB holds database connections and configuration
type DB struct {
	Pool       *pgxpool.Pool // Only used for PostgreSQL
	SQLX       *sqlx.DB
	DB         *sqlx.DB // Alias for SQLX to match repository expectations
	Config     *config.Config
	Logger     zerolog.Logger
	DbType     string // "sqlite" or "postgres"
}

// New creates a new database connection
func New(cfg *config.Config, logger zerolog.Logger) (*DB, error) {
	var pool *pgxpool.Pool
	var sqlxDB *sqlx.DB
	var err error

	switch cfg.DatabaseType {
	case "sqlite":
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

	case "postgres":
		// Configure PostgreSQL connection pool
		poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PostgreSQL database URL: %w", err)
		}

		// Set pool configuration
		poolConfig.MaxConns = 30
		poolConfig.MinConns = 5
		poolConfig.MaxConnLifetime = time.Hour
		poolConfig.MaxConnIdleTime = time.Minute * 30
		poolConfig.HealthCheckPeriod = time.Minute

		// Create connection pool
		pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
		}

		// Create sqlx connection for compatibility
		sqlxDB, err = sqlx.Connect("pgx", cfg.DatabaseURL)
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to create PostgreSQL sqlx connection: %w", err)
		}

		logger.Info().Str("host", cfg.DatabaseHost).Str("database", cfg.DatabaseName).Msg("PostgreSQL database connection established")

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DatabaseType)
	}

	// Test the sqlx connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlxDB.PingContext(ctx); err != nil {
		if pool != nil {
			pool.Close()
		}
		sqlxDB.Close()
		return nil, fmt.Errorf("failed to ping database via sqlx: %w", err)
	}

	db := &DB{
		Pool:   pool,   // Will be nil for SQLite
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
	if db.Pool != nil {
		db.Pool.Close()
	}
	if db.SQLX != nil {
		db.SQLX.Close()
	}
	db.Logger.Info().Msg("Database connections closed")
}

// Health checks the database connection health
func (db *DB) Health(ctx context.Context) error {
	// Check pool connection if PostgreSQL
	if db.Pool != nil {
		if err := db.Pool.Ping(ctx); err != nil {
			return fmt.Errorf("pgxpool health check failed: %w", err)
		}
	}

	// Check sqlx connection (works for both SQLite and PostgreSQL)
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
			"in_use":              sqlxStats.InUse,
			"idle":                sqlxStats.Idle,
		},
	}

	// Add pool stats if PostgreSQL
	if db.Pool != nil {
		poolStats := db.Pool.Stat()
		stats["pgxpool"] = map[string]interface{}{
			"acquired_conns":     poolStats.AcquiredConns(),
			"constructing_conns": poolStats.ConstructingConns(),
			"idle_conns":         poolStats.IdleConns(),
			"max_conns":          poolStats.MaxConns(),
			"total_conns":        poolStats.TotalConns(),
		}
	}

	return stats
}