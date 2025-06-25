package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"bookminderapi/internal/models"
)

// Logger manages structured logging for the application
type Logger struct {
	logFile *os.File
}

// NewLogger creates a new logger instance
func NewLogger(logFilePath string) (*Logger, error) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	log.Printf("Structured logging initialized: %s", logFilePath)
	
	return &Logger{
		logFile: logFile,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// LogStructured writes a structured log entry to both console and file
func (l *Logger) LogStructured(level, component, message string, data map[string]interface{}) {
	entry := models.LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Component: component,
		Data:      data,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	// Write to file if available
	if l.logFile != nil {
		l.logFile.WriteString(string(jsonData) + "\n")
		l.logFile.Sync()
	}
}

// Global logger instance (to be initialized by main)
var globalLogger *Logger

// InitLogging initializes the global logger
func InitLogging(logFilePath string) error {
	logger, err := NewLogger(logFilePath)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// LogStructured is a convenience function that uses the global logger
func LogStructured(level, component, message string, data map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.LogStructured(level, component, message, data)
	}
}

// CloseLogging closes the global logger
func CloseLogging() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}