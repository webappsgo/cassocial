package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Logger provides structured logging capabilities
type Logger struct {
	accessLog   *log.Logger
	serverLog   *log.Logger
	errorLog    *log.Logger
	auditLog    *log.Logger
	securityLog *log.Logger
	format      string // text or json
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	Component  string                 `json:"component,omitempty"`
	Message    string                 `json:"message"`
	Error      string                 `json:"error,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	IP         string                 `json:"ip,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Status     int                    `json:"status,omitempty"`
	Duration   int64                  `json:"duration_ms,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// NewLogger creates a new logger instance
func NewLogger(logDir string, format string) (*Logger, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log files
	accessFile, err := openLogFile(filepath.Join(logDir, "access.log"))
	if err != nil {
		return nil, err
	}

	serverFile, err := openLogFile(filepath.Join(logDir, "server.log"))
	if err != nil {
		return nil, err
	}

	errorFile, err := openLogFile(filepath.Join(logDir, "error.log"))
	if err != nil {
		return nil, err
	}

	auditFile, err := openLogFile(filepath.Join(logDir, "audit.log"))
	if err != nil {
		return nil, err
	}

	securityFile, err := openLogFile(filepath.Join(logDir, "security.log"))
	if err != nil {
		return nil, err
	}

	return &Logger{
		accessLog:   log.New(accessFile, "", 0),
		serverLog:   log.New(serverFile, "", 0),
		errorLog:    log.New(errorFile, "", 0),
		auditLog:    log.New(auditFile, "", 0),
		securityLog: log.New(securityFile, "", 0),
		format:      format,
	}, nil
}

// LogAccess logs an HTTP access
func (l *Logger) LogAccess(method, path string, status int, duration time.Duration, ip, userAgent string) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "info",
		Component: "http",
		Message:   fmt.Sprintf("%s %s", method, path),
		IP:        ip,
		Method:    method,
		Path:      path,
		Status:    status,
		Duration:  duration.Milliseconds(),
	}

	l.writeLog(l.accessLog, entry)
}

// LogError logs an error
func (l *Logger) LogError(component, message string, err error) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "error",
		Component: component,
		Message:   message,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.writeLog(l.errorLog, entry)
}

// LogAudit logs an audit event
func (l *Logger) LogAudit(userID, action, resource string, details map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "audit",
		UserID:    userID,
		Message:   fmt.Sprintf("%s on %s", action, resource),
		Extra:     details,
	}

	// Audit log is ALWAYS JSON format
	if data, err := json.Marshal(entry); err == nil {
		l.auditLog.Println(string(data))
	}
}

// LogSecurity logs a security event
func (l *Logger) LogSecurity(eventType, message, ip string, details map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "security",
		Component: eventType,
		Message:   message,
		IP:        ip,
		Extra:     details,
	}

	l.writeLog(l.securityLog, entry)
}

// LogInfo logs an informational message
func (l *Logger) LogInfo(component, message string) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "info",
		Component: component,
		Message:   message,
	}

	l.writeLog(l.serverLog, entry)
}

// LogDebug logs a debug message
func (l *Logger) LogDebug(component, message string, details map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "debug",
		Component: component,
		Message:   message,
		Extra:     details,
	}

	l.writeLog(l.serverLog, entry)
}

// writeLog writes a log entry in the configured format
func (l *Logger) writeLog(logger *log.Logger, entry LogEntry) {
	if l.format == "json" {
		// JSON format
		if data, err := json.Marshal(entry); err == nil {
			logger.Println(string(data))
		}
	} else {
		// Text format
		msg := fmt.Sprintf("[%s] %s", entry.Level, entry.Message)
		if entry.Component != "" {
			msg = fmt.Sprintf("[%s] [%s] %s", entry.Level, entry.Component, entry.Message)
		}
		if entry.Error != "" {
			msg += fmt.Sprintf(" - error: %s", entry.Error)
		}
		logger.Println(msg)
	}
}

// openLogFile opens a log file for appending
func openLogFile(path string) (io.Writer, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
	}
	return file, nil
}

// LoggingMiddleware is middleware that logs all requests
func LoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log request
			duration := time.Since(start)
			ip := getClientIPFromRequest(r)

			logger.LogAccess(r.Method, r.URL.Path, wrapped.statusCode, duration, ip, r.UserAgent())
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func getClientIPFromRequest(r *http.Request) string {
	// Check X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Strip port from RemoteAddr if present.
	addr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}
