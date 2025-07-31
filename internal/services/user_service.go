package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/models"
)

// UserService handles user-related operations
type UserService struct {
	db     *sqlx.DB
	logger zerolog.Logger
}

// NewUserService creates a new user service
func NewUserService(db *sqlx.DB, logger zerolog.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, role, status, created_at, updated_at FROM users WHERE id = $1`
	
	err := s.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// UpdateUserRole updates a user's role
func (s *UserService) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	query := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`
	
	result, err := s.db.ExecContext(ctx, query, role, userID)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// ListUsers lists users with optional role filtering
func (s *UserService) ListUsers(ctx context.Context, role string, limit, offset int) ([]models.User, int, error) {
	var users []models.User
	var total int
	
	// Build query
	countQuery := `SELECT COUNT(*) FROM users`
	query := `SELECT id, email, role, status, created_at, updated_at FROM users`
	
	if role != "" {
		countQuery += ` WHERE role = $1`
		query += ` WHERE role = $1`
	}
	
	query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	
	// Get total count
	var args []interface{}
	if role != "" {
		args = append(args, role)
		err := s.db.GetContext(ctx, &total, countQuery, role)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get user count: %w", err)
		}
	} else {
		err := s.db.GetContext(ctx, &total, countQuery)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get user count: %w", err)
		}
	}
	
	// Get users
	if role != "" {
		err := s.db.SelectContext(ctx, &users, query, role, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list users: %w", err)
		}
	} else {
		// Adjust query for no role filter
		query = `SELECT id, email, role, status, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		err := s.db.SelectContext(ctx, &users, query, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list users: %w", err)
		}
	}
	
	return users, total, nil
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers     int            `json:"total_users"`
	UsersByRole    map[string]int `json:"users_by_role"`
	ActiveUsers    int            `json:"active_users"`
	TotalAnalyses  int            `json:"total_analyses"`
	StorageUsed    int64          `json:"storage_used_bytes"`
	ProcessingJobs int            `json:"processing_jobs"`
}

// GetSystemStats returns system-wide statistics
func (s *UserService) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	stats := &SystemStats{
		UsersByRole: make(map[string]int),
	}
	
	// Get total users
	err := s.db.GetContext(ctx, &stats.TotalUsers, `SELECT COUNT(*) FROM users`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}
	
	// Get users by role
	rows, err := s.db.QueryContext(ctx, `SELECT role, COUNT(*) FROM users GROUP BY role`)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var role string
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			return nil, fmt.Errorf("failed to scan role count: %w", err)
		}
		stats.UsersByRole[role] = count
	}
	
	// Get active users (logged in within last 24 hours)
	err = s.db.GetContext(ctx, &stats.ActiveUsers, 
		`SELECT COUNT(*) FROM users WHERE last_login_at > NOW() - INTERVAL '24 hours'`)
	if err != nil {
		// Non-critical error, continue
		s.logger.Warn().Err(err).Msg("Failed to get active users count")
	}
	
	// Get total analyses
	err = s.db.GetContext(ctx, &stats.TotalAnalyses, `SELECT COUNT(*) FROM analyses`)
	if err != nil {
		// Non-critical error, continue
		s.logger.Warn().Err(err).Msg("Failed to get total analyses count")
	}
	
	// Get storage used
	err = s.db.GetContext(ctx, &stats.StorageUsed, 
		`SELECT COALESCE(SUM(file_size), 0) FROM analyses`)
	if err != nil {
		// Non-critical error, continue
		s.logger.Warn().Err(err).Msg("Failed to get storage used")
	}
	
	// Get processing jobs
	err = s.db.GetContext(ctx, &stats.ProcessingJobs, 
		`SELECT COUNT(*) FROM analyses WHERE status IN ('pending', 'processing')`)
	if err != nil {
		// Non-critical error, continue
		s.logger.Warn().Err(err).Msg("Failed to get processing jobs count")
	}
	
	return stats, nil
}