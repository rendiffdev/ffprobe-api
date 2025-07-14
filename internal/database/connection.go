package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/config"
)

// DB holds database connections and configuration
type DB struct {
	Pool   *pgxpool.Pool
	SQLX   *sqlx.DB
	DB     *sqlx.DB // Alias for SQLX to match repository expectations
	Config *config.Config
	Logger zerolog.Logger
}

// New creates a new database connection
func New(cfg *config.Config, logger zerolog.Logger) (*DB, error) {
	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = 30
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create sqlx connection for compatibility
	sqlxDB, err := sqlx.Connect("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlx connection: %w", err)
	}

	db := &DB{
		Pool:   pool,
		SQLX:   sqlxDB,
		DB:     sqlxDB, // Set the alias
		Config: cfg,
		Logger: logger,
	}

	logger.Info().Msg("Database connection established successfully")
	return db, nil
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
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("pgxpool health check failed: %w", err)
	}

	if err := db.SQLX.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlx health check failed: %w", err)
	}

	return nil
}

// Stats returns database connection statistics
func (db *DB) Stats() map[string]interface{} {
	poolStats := db.Pool.Stat()
	sqlxStats := db.SQLX.Stats()

	return map[string]interface{}{
		"pgxpool": map[string]interface{}{
			"acquired_conns":     poolStats.AcquiredConns(),
			"constructing_conns": poolStats.ConstructingConns(),
			"idle_conns":         poolStats.IdleConns(),
			"max_conns":          poolStats.MaxConns(),
			"total_conns":        poolStats.TotalConns(),
		},
		"sqlx": map[string]interface{}{
			"max_open_connections": sqlxStats.MaxOpenConnections,
			"open_connections":     sqlxStats.OpenConnections,
			"in_use":              sqlxStats.InUse,
			"idle":                sqlxStats.Idle,
		},
	}
}