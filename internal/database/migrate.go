package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
)

// MigrateUp runs all available migrations
func MigrateUp(databaseURL string, migrationsPath string, logger zerolog.Logger) error {
	// Determine migration path based on database type
	finalMigrationsPath := getMigrationPath(databaseURL, migrationsPath)
	
	m, err := migrate.New(
		fmt.Sprintf("file://%s", finalMigrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info().Msg("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info().Msg("Successfully applied all migrations")
	return nil
}

// getMigrationPath returns the appropriate migration path based on database type
func getMigrationPath(databaseURL, basePath string) string {
	if strings.HasPrefix(databaseURL, "sqlite3://") {
		// Use SQLite-specific migrations if they exist
		return basePath // For now, use the same path but we could organize differently
	}
	return basePath
}

// MigrateDown rolls back one migration
func MigrateDown(databaseURL string, migrationsPath string, logger zerolog.Logger) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info().Msg("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	logger.Info().Msg("Successfully rolled back one migration")
	return nil
}

// MigrateToVersion migrates to a specific version
func MigrateToVersion(databaseURL string, migrationsPath string, version uint, logger zerolog.Logger) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Migrate(version); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info().Uint("version", version).Msg("Already at target version")
			return nil
		}
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	logger.Info().Uint("version", version).Msg("Successfully migrated to version")
	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(databaseURL string, migrationsPath string) (uint, bool, error) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}