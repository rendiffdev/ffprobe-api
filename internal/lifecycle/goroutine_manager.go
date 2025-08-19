package lifecycle

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// GoroutineManager manages the lifecycle of goroutines in the application
type GoroutineManager struct {
	logger          zerolog.Logger
	mu              sync.RWMutex
	goroutines      map[string]*ManagedGoroutine
	shutdown        chan struct{}
	shutdownOnce    sync.Once
	activeCount     int64
	maxGoroutines   int
	shutdownTimeout time.Duration
}

// ManagedGoroutine represents a managed goroutine
type ManagedGoroutine struct {
	ID        string
	Name      string
	StartTime time.Time
	Context   context.Context
	Cancel    context.CancelFunc
	Done      chan struct{}
	Status    GoroutineStatus
	mu        sync.RWMutex
}

// GoroutineStatus represents the status of a goroutine
type GoroutineStatus string

const (
	StatusStarting GoroutineStatus = "STARTING"
	StatusRunning  GoroutineStatus = "RUNNING"
	StatusStopping GoroutineStatus = "STOPPING"
	StatusStopped  GoroutineStatus = "STOPPED"
	StatusError    GoroutineStatus = "ERROR"
)

// GoroutineConfig configures a managed goroutine
type GoroutineConfig struct {
	Name            string
	Ctx             context.Context
	MaxRetries      int
	RetryDelay      time.Duration
	HealthCheckFunc func() error
	OnError         func(error)
}

// NewGoroutineManager creates a new goroutine manager
func NewGoroutineManager(logger zerolog.Logger, maxGoroutines int) *GoroutineManager {
	return &GoroutineManager{
		logger:          logger,
		goroutines:      make(map[string]*ManagedGoroutine),
		shutdown:        make(chan struct{}),
		maxGoroutines:   maxGoroutines,
		shutdownTimeout: 30 * time.Second,
	}
}

// Start starts a managed goroutine
func (gm *GoroutineManager) Start(config GoroutineConfig, fn func(context.Context) error) (string, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if we've reached the maximum number of goroutines
	if len(gm.goroutines) >= gm.maxGoroutines {
		return "", fmt.Errorf("maximum goroutines limit reached (%d)", gm.maxGoroutines)
	}

	// Generate unique ID
	id := fmt.Sprintf("%s-%d", config.Name, time.Now().UnixNano())

	// Create context
	ctx := config.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	// Create managed goroutine
	mg := &ManagedGoroutine{
		ID:        id,
		Name:      config.Name,
		StartTime: time.Now(),
		Context:   ctx,
		Cancel:    cancel,
		Done:      make(chan struct{}),
		Status:    StatusStarting,
	}

	gm.goroutines[id] = mg
	atomic.AddInt64(&gm.activeCount, 1)

	// Start the goroutine
	go gm.runManagedGoroutine(mg, config, fn)

	gm.logger.Info().
		Str("goroutine_id", id).
		Str("goroutine_name", config.Name).
		Int("active_count", int(atomic.LoadInt64(&gm.activeCount))).
		Msg("Started managed goroutine")

	return id, nil
}

// Stop stops a specific goroutine by ID
func (gm *GoroutineManager) Stop(id string) error {
	gm.mu.RLock()
	mg, exists := gm.goroutines[id]
	gm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("goroutine with ID %s not found", id)
	}

	mg.mu.Lock()
	if mg.Status == StatusStopping || mg.Status == StatusStopped {
		mg.mu.Unlock()
		return nil
	}
	mg.Status = StatusStopping
	mg.mu.Unlock()

	mg.Cancel()

	// Wait for goroutine to finish
	select {
	case <-mg.Done:
		gm.logger.Info().
			Str("goroutine_id", id).
			Str("goroutine_name", mg.Name).
			Msg("Goroutine stopped successfully")
		return nil
	case <-time.After(gm.shutdownTimeout):
		gm.logger.Warn().
			Str("goroutine_id", id).
			Str("goroutine_name", mg.Name).
			Msg("Goroutine stop timeout")
		return fmt.Errorf("timeout waiting for goroutine %s to stop", id)
	}
}

// StopAll stops all managed goroutines
func (gm *GoroutineManager) StopAll() error {
	gm.shutdownOnce.Do(func() {
		close(gm.shutdown)
	})

	gm.mu.RLock()
	goroutineIDs := make([]string, 0, len(gm.goroutines))
	for id := range gm.goroutines {
		goroutineIDs = append(goroutineIDs, id)
	}
	gm.mu.RUnlock()

	gm.logger.Info().
		Int("count", len(goroutineIDs)).
		Msg("Stopping all managed goroutines")

	// Cancel all goroutines
	for _, id := range goroutineIDs {
		gm.mu.RLock()
		mg, exists := gm.goroutines[id]
		gm.mu.RUnlock()

		if exists {
			mg.mu.Lock()
			mg.Status = StatusStopping
			mg.mu.Unlock()
			mg.Cancel()
		}
	}

	// Wait for all to finish
	timeout := time.After(gm.shutdownTimeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			remaining := atomic.LoadInt64(&gm.activeCount)
			if remaining > 0 {
				gm.logger.Warn().
					Int64("remaining_count", remaining).
					Msg("Timeout waiting for all goroutines to stop")
				return fmt.Errorf("timeout: %d goroutines still running", remaining)
			}
			return nil
		case <-ticker.C:
			if atomic.LoadInt64(&gm.activeCount) == 0 {
				gm.logger.Info().Msg("All managed goroutines stopped successfully")
				return nil
			}
		}
	}
}

// GetStatus returns the status of all managed goroutines
func (gm *GoroutineManager) GetStatus() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	status := map[string]interface{}{
		"total_goroutines":   len(gm.goroutines),
		"active_count":       atomic.LoadInt64(&gm.activeCount),
		"max_goroutines":     gm.maxGoroutines,
		"runtime_goroutines": runtime.NumGoroutine(),
		"goroutines":         make(map[string]interface{}),
	}

	for id, mg := range gm.goroutines {
		mg.mu.RLock()
		status["goroutines"].(map[string]interface{})[id] = map[string]interface{}{
			"name":       mg.Name,
			"status":     mg.Status,
			"start_time": mg.StartTime,
			"uptime":     time.Since(mg.StartTime),
		}
		mg.mu.RUnlock()
	}

	return status
}

// GetGoroutineCount returns the current number of active goroutines
func (gm *GoroutineManager) GetGoroutineCount() int64 {
	return atomic.LoadInt64(&gm.activeCount)
}

// runManagedGoroutine runs the actual goroutine with management
func (gm *GoroutineManager) runManagedGoroutine(mg *ManagedGoroutine, config GoroutineConfig, fn func(context.Context) error) {
	defer func() {
		// Cleanup
		mg.mu.Lock()
		mg.Status = StatusStopped
		mg.mu.Unlock()

		close(mg.Done)
		atomic.AddInt64(&gm.activeCount, -1)

		gm.mu.Lock()
		delete(gm.goroutines, mg.ID)
		gm.mu.Unlock()

		// Handle panics
		if r := recover(); r != nil {
			gm.logger.Error().
				Str("goroutine_id", mg.ID).
				Str("goroutine_name", mg.Name).
				Interface("panic", r).
				Msg("Managed goroutine panicked")

			if config.OnError != nil {
				config.OnError(fmt.Errorf("panic in goroutine %s: %v", mg.Name, r))
			}
		}
	}()

	mg.mu.Lock()
	mg.Status = StatusRunning
	mg.mu.Unlock()

	// Run the function with retry logic
	retries := 0
	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3 // Default retry count
	}

	retryDelay := config.RetryDelay
	if retryDelay == 0 {
		retryDelay = 1 * time.Second // Default retry delay
	}

	for {
		select {
		case <-mg.Context.Done():
			gm.logger.Debug().
				Str("goroutine_id", mg.ID).
				Str("goroutine_name", mg.Name).
				Msg("Goroutine context cancelled")
			return
		case <-gm.shutdown:
			gm.logger.Debug().
				Str("goroutine_id", mg.ID).
				Str("goroutine_name", mg.Name).
				Msg("Goroutine shutdown requested")
			return
		default:
			err := fn(mg.Context)
			if err == nil {
				// Success, reset retry count
				retries = 0

				// If function returns without error, it completed successfully
				gm.logger.Debug().
					Str("goroutine_id", mg.ID).
					Str("goroutine_name", mg.Name).
					Msg("Goroutine function completed successfully")
				return
			}

			// Handle error
			gm.logger.Error().
				Err(err).
				Str("goroutine_id", mg.ID).
				Str("goroutine_name", mg.Name).
				Int("retry_count", retries).
				Msg("Goroutine function error")

			if config.OnError != nil {
				config.OnError(err)
			}

			retries++
			if retries > maxRetries {
				mg.mu.Lock()
				mg.Status = StatusError
				mg.mu.Unlock()

				gm.logger.Error().
					Str("goroutine_id", mg.ID).
					Str("goroutine_name", mg.Name).
					Int("max_retries", maxRetries).
					Msg("Goroutine exceeded max retries")
				return
			}

			// Wait before retry
			select {
			case <-mg.Context.Done():
				return
			case <-gm.shutdown:
				return
			case <-time.After(retryDelay):
				// Continue to retry
			}
		}
	}
}

// HealthCheck performs health checks on all running goroutines
func (gm *GoroutineManager) HealthCheck() error {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	var errors []string
	for _, mg := range gm.goroutines {
		mg.mu.RLock()
		if mg.Status == StatusError {
			errors = append(errors, fmt.Sprintf("goroutine %s (%s) is in error state", mg.ID, mg.Name))
		}
		mg.mu.RUnlock()
	}

	if len(errors) > 0 {
		return fmt.Errorf("health check failed: %v", errors)
	}

	return nil
}
