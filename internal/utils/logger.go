package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger provides structured logging for the application
type Logger struct {
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
	warningLogger *log.Logger
	logFile       *os.File
}

// NewLogger creates a new logger that writes to both console and file
func NewLogger(logDir string) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Create log file with current date
	logPath := filepath.Join(logDir, fmt.Sprintf("linefilecatcher_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	// Create multi-writer to log to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Create loggers with prefixes
	infoLogger := log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger := log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger := log.New(multiWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	warningLogger := log.New(multiWriter, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

	return &Logger{
		infoLogger:    infoLogger,
		errorLogger:   errorLogger,
		debugLogger:   debugLogger,
		warningLogger: warningLogger,
		logFile:       logFile,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	return l.logFile.Close()
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		l.debugLogger.Printf(format, v...)
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, v ...interface{}) {
	l.warningLogger.Printf(format, v...)
}
