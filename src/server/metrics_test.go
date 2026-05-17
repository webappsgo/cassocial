package server

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}
	if m.startTime.IsZero() {
		t.Error("startTime is zero")
	}
}

func TestRecordRequest_NoError(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest(100*time.Millisecond, false)
	if m.requestCount != 1 {
		t.Errorf("requestCount = %d, want 1", m.requestCount)
	}
	if m.errorCount != 0 {
		t.Errorf("errorCount = %d, want 0", m.errorCount)
	}
}

func TestRecordRequest_WithError(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest(50*time.Millisecond, true)
	if m.requestCount != 1 {
		t.Errorf("requestCount = %d, want 1", m.requestCount)
	}
	if m.errorCount != 1 {
		t.Errorf("errorCount = %d, want 1", m.errorCount)
	}
}

func TestRecordRequest_Multiple(t *testing.T) {
	m := NewMetrics()
	for i := 0; i < 5; i++ {
		m.RecordRequest(10*time.Millisecond, i%2 == 0)
	}
	if m.requestCount != 5 {
		t.Errorf("requestCount = %d, want 5", m.requestCount)
	}
	if m.errorCount != 3 {
		t.Errorf("errorCount = %d, want 3 (i=0,2,4)", m.errorCount)
	}
}

func TestRecordDBQuery(t *testing.T) {
	m := NewMetrics()
	m.RecordDBQuery(5 * time.Millisecond)
	m.RecordDBQuery(10 * time.Millisecond)
	if m.dbQueryCount != 2 {
		t.Errorf("dbQueryCount = %d, want 2", m.dbQueryCount)
	}
}

func TestIncrementDecrementConnections(t *testing.T) {
	m := NewMetrics()
	m.IncrementConnections()
	m.IncrementConnections()
	if m.activeConnections != 2 {
		t.Errorf("activeConnections = %d, want 2", m.activeConnections)
	}
	m.DecrementConnections()
	if m.activeConnections != 1 {
		t.Errorf("activeConnections = %d, want 1", m.activeConnections)
	}
}

func TestServeMetrics_PrometheusFormat(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest(20*time.Millisecond, false)
	m.RecordDBQuery(3 * time.Millisecond)
	m.IncrementConnections()

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	m.ServeMetrics(rr, req)

	if rr.Code != 200 {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	required := []string{
		"cassocial_uptime_seconds",
		"cassocial_requests_total",
		"cassocial_errors_total",
		"cassocial_db_queries_total",
		"cassocial_active_connections",
		"cassocial_goroutines",
		"cassocial_memory_alloc_bytes",
	}
	for _, label := range required {
		if !strings.Contains(body, label) {
			t.Errorf("metrics output missing %q", label)
		}
	}
}

func TestServeMetrics_ZeroRequests(t *testing.T) {
	m := NewMetrics()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	m.ServeMetrics(rr, req)
	if rr.Code != 200 {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestGetStats_ContainsExpectedKeys(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest(10*time.Millisecond, false)

	stats := m.GetStats()
	expectedKeys := []string{
		"uptime_seconds", "request_count", "error_count",
		"db_query_count", "active_connections", "goroutines",
		"memory_alloc_bytes", "memory_sys_bytes", "cpu_count",
	}
	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("GetStats() missing key %q", key)
		}
	}
}

func TestGetStats_RequestCount(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest(1*time.Millisecond, false)
	m.RecordRequest(1*time.Millisecond, true)

	stats := m.GetStats()
	if stats["request_count"].(int64) != 2 {
		t.Errorf("request_count = %v, want 2", stats["request_count"])
	}
	if stats["error_count"].(int64) != 1 {
		t.Errorf("error_count = %v, want 1", stats["error_count"])
	}
}
