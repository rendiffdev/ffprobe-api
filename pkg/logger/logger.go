package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// New creates a new logger with the specified level
func New(level string) zerolog.Logger {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set up console writer for human-readable output in development
	if strings.ToLower(os.Getenv("GO_ENV")) != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Set global log level
	zerolog.SetGlobalLevel(logLevel)

	return log.Logger
}