package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestNewLogger_InvalidDir(t *testing.T) {
	// A path that cannot be created (parent is a file, not a dir)
	base := t.TempDir()
	// Create a file where the log subdir should go
	if err := os.WriteFile(base+"/blocker", []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// logDir is base/blocker/logs — blocker is a file, MkdirAll will fail
	_, err := NewLogger(base+"/blocker/logs", "text")
	if err == nil {
		t.Error("NewLogger should fail when log directory cannot be created")
	}
}

func TestOpenLogFile_Failure(t *testing.T) {
	// Pass a path whose parent directory does not exist
	_, err := openLogFile("/nonexistent/path/that/cannot/exist/file.log")
	if err == nil {
		t.Error("openLogFile should fail when parent directory does not exist")
	}
}

func TestGetClientIPFromRequest_NoHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	// RemoteAddr has host:port form — SplitHostPort should succeed and strip port
	req.RemoteAddr = "10.0.0.1:9999"
	ip := getClientIPFromRequest(req)
	if ip != "10.0.0.1" {
		t.Errorf("getClientIPFromRequest = %q, want 10.0.0.1", ip)
	}
}

func TestGetClientIPFromRequest_BareAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	// No port — SplitHostPort fails, should return raw addr
	req.RemoteAddr = "10.0.0.2"
	ip := getClientIPFromRequest(req)
	if ip != "10.0.0.2" {
		t.Errorf("getClientIPFromRequest bare = %q, want 10.0.0.2", ip)
	}
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

// TestNewLogger_AccessLogBlocked exercises the branch where access.log cannot be opened.
func TestNewLogger_AccessLogBlocked(t *testing.T) {
	base := t.TempDir()
	// Create access.log as a directory so os.OpenFile fails.
	if err := os.MkdirAll(base+"/access.log", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	_, err := NewLogger(base, "text")
	if err == nil {
		t.Error("NewLogger should fail when access.log cannot be opened")
	}
}

// TestNewLogger_ServerLogBlocked exercises the branch where server.log cannot be opened.
func TestNewLogger_ServerLogBlocked(t *testing.T) {
	base := t.TempDir()
	// Create server.log as a directory so os.OpenFile fails.
	if err := os.MkdirAll(base+"/server.log", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	_, err := NewLogger(base, "text")
	if err == nil {
		t.Error("NewLogger should fail when server.log cannot be opened")
	}
}

// TestNewLogger_ErrorLogBlocked exercises the branch where error.log cannot be opened.
func TestNewLogger_ErrorLogBlocked(t *testing.T) {
	base := t.TempDir()
	// Create error.log as a directory so os.OpenFile fails.
	if err := os.MkdirAll(base+"/error.log", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	_, err := NewLogger(base, "text")
	if err == nil {
		t.Error("NewLogger should fail when error.log cannot be opened")
	}
}

// TestNewLogger_AuditLogBlocked exercises the branch where audit.log cannot be opened.
func TestNewLogger_AuditLogBlocked(t *testing.T) {
	base := t.TempDir()
	// Create audit.log as a directory so os.OpenFile fails.
	if err := os.MkdirAll(base+"/audit.log", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	_, err := NewLogger(base, "text")
	if err == nil {
		t.Error("NewLogger should fail when audit.log cannot be opened")
	}
}

// TestNewLogger_SecurityLogBlocked exercises the branch where security.log cannot be opened.
func TestNewLogger_SecurityLogBlocked(t *testing.T) {
	base := t.TempDir()
	// Create security.log as a directory so os.OpenFile fails.
	if err := os.MkdirAll(base+"/security.log", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	_, err := NewLogger(base, "text")
	if err == nil {
		t.Error("NewLogger should fail when security.log cannot be opened")
	}
}
