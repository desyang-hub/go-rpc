package observability

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog for structured logging in RPC components.
type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new Logger with default configuration.
func NewLogger() *Logger {
	return &Logger{
		logger: log.With().Timestamp().Logger().Output(os.Stdout),
	}
}

// WithLevel sets the minimum log level.
func (l *Logger) WithLevel(level zerolog.Level) *Logger {
	l.logger.Level(level)
	return l
}

// Info logs an informational message with optional fields.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	entry := l.logger.Info()
	l.addFields(entry, fields)
}

// Error logs an error message with optional fields.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	entry := l.logger.Error()
	l.addFields(entry, fields)
}

// Warn logs a warning message with optional fields.
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	entry := l.logger.Warn()
	l.addFields(entry, fields)
}

// Debug logs a debug message with optional fields.
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	entry := l.logger.Debug()
	l.addFields(entry, fields)
}

// addFields adds key-value pairs to the log entry.
func (l *Logger) addFields(entry *zerolog.Event, fields map[string]interface{}) {
	for key, value := range fields {
		switch v := value.(type) {
		case string:
			entry.Str(key, v)
		case int, int8, int16, int32, int64:
			entry.Int64(key, v.(int64))
		case float64:
			entry.Float64(key, v)
		case time.Time:
			entry.Time(key, v)
		case time.Duration:
			entry.Int64(key+" us", v.Microseconds())
		case error:
			entry.Err(v)
		}
	}
}

// GetLogger returns the underlying zerolog.Logger.
func (l *Logger) GetLogger() *zerolog.Logger {
	return &l.logger
}
