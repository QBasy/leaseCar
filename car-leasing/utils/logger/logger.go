package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	var err error
	log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// Info logs an info level message
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Error logs an error level message
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Warn logs a warning level message
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Debug logs a debug level message
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// WithFields returns a logger with structured fields
func WithFields(fields ...zap.Field) *zap.Logger {
	return log.With(fields...)
}

// GetLogger returns the raw zap logger
func GetLogger() *zap.Logger {
	return log
}

// Sync flushes any buffered log entries
func Sync() error {
	return log.Sync()
}
