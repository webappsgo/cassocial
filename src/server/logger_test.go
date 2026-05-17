package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestLogger(t *testing.T) *Logger {
	t.Helper()
	dir := t.TempDir()
	lg, err := NewLogger(dir, "text")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	return lg
}

func TestNewLogger_Text(t *testing.T) {
	dir := t.TempDir()
	lg, err := NewLogger(dir, "text")
	if err != nil {
		t.Fatalf("NewLogger(text): %v", err)
	}
	if lg == nil {
		t.Fatal("NewLogger returned nil")
	}
}

func TestNewLogger_JSON(t *testing.T) {
	dir := t.TempDir()
	lg, err := NewLogger(dir, "json")
	if err != nil {
		t.Fatalf("NewLogger(json): %v", err)
	}
	if lg == nil {
		t.Fatal("NewLogger returned nil")
	}
}

func TestNewLogger_CreatesDirectory(t *testing.T) {
	base := t.TempDir()
	dir := base + "/subdir/logs"
	_, err := NewLogger(dir, "text")
	if err != nil {
		t.Fatalf("NewLogger with nested dir: %v", err)
	}
}

func TestLogAccess_NoPanic(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogAccess("GET", "/test", 200, 10*time.Millisecond, "127.0.0.1", "test-agent")
}

func TestLogError_NilError(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogError("component", "some error message", nil)
}

func TestLogError_WithError(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogError("component", "something failed", errors.New("disk full"))
}

func TestLogAudit_NoPanic(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogAudit("user-1", "delete", "post/42", map[string]interface{}{"reason": "spam"})
}

func TestLogAudit_NilDetails(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogAudit("user-1", "login", "session", nil)
}

func TestLogSecurity_NoPanic(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogSecurity("brute-force", "too many login attempts", "1.2.3.4", map[string]interface{}{"attempts": 5})
}

func TestLogSecurity_NilDetails(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogSecurity("csrf", "csrf token mismatch", "::1", nil)
}

func TestLogInfo_NoPanic(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogInfo("server", "server started")
}

func TestLogDebug_NoPanic(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogDebug("router", "request matched", map[string]interface{}{"route": "/api/v1/users"})
}

func TestLogDebug_NilDetails(t *testing.T) {
	lg := newTestLogger(t)
	lg.LogDebug("router", "request matched", nil)
}

func TestWriteLog_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	lg, err := NewLogger(dir, "json")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	// Exercises the json branch in writeLog
	lg.LogInfo("test", "json log entry")
	lg.LogError("test", "json error entry", errors.New("err"))
}

func TestWriteLog_TextFormatWithComponent(t *testing.T) {
	lg := newTestLogger(t)
	// Component present path
	lg.LogInfo("mycomponent", "hello")
}

func TestWriteLog_TextFormatNoComponent(t *testing.T) {
	lg := newTestLogger(t)
	entry := LogEntry{
		Level:   "info",
		Message: "no component",
	}
	lg.writeLog(lg.serverLog, entry)
}

func TestLoggingMiddleware_200(t *testing.T) {
	lg := newTestLogger(t)
	handler := LoggingMiddleware(lg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestLoggingMiddleware_404(t *testing.T) {
	lg := newTestLogger(t)
	handler := LoggingMiddleware(lg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest("GET", "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestLoggingMiddleware_XForwardedFor(t *testing.T) {
	lg := newTestLogger(t)
	var capturedReq *http.Request
	handler := LoggingMiddleware(lg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	_ = capturedReq
}

func TestLoggingMiddleware_XRealIP(t *testing.T) {
	lg := newTestLogger(t)
	handler := LoggingMiddleware(lg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rr, statusCode: http.StatusOK}
	rw.WriteHeader(http.StatusTeapot)
	if rw.statusCode != http.StatusTeapot {
		t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusTeapot)
	}
	if rr.Code != http.StatusTeapot {
		t.Errorf("recorder code = %d, want %d", rr.Code, http.StatusTeapot)
	}
}
