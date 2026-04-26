package server

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Metrics tracks application metrics
type Metrics struct {
	mu                sync.RWMutex
	requestCount      int64
	requestDuration   time.Duration
	errorCount        int64
	dbQueryCount      int64
	dbQueryDuration   time.Duration
	activeConnections int64
	startTime         time.Time
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// RecordRequest records a request with duration
func (m *Metrics) RecordRequest(duration time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount++
	m.requestDuration += duration

	if isError {
		m.errorCount++
	}
}

// RecordDBQuery records a database query
func (m *Metrics) RecordDBQuery(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dbQueryCount++
	m.dbQueryDuration += duration
}

// IncrementConnections increments active connections
func (m *Metrics) IncrementConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeConnections++
}

// DecrementConnections decrements active connections
func (m *Metrics) DecrementConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeConnections--
}

// ServeMetrics serves Prometheus-compatible metrics
func (m *Metrics) ServeMetrics(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime).Seconds()

	// Calculate averages
	avgRequestDuration := float64(0)
	if m.requestCount > 0 {
		avgRequestDuration = float64(m.requestDuration.Milliseconds()) / float64(m.requestCount)
	}

	avgDBQueryDuration := float64(0)
	if m.dbQueryCount > 0 {
		avgDBQueryDuration = float64(m.dbQueryDuration.Milliseconds()) / float64(m.dbQueryCount)
	}

	errorRate := float64(0)
	if m.requestCount > 0 {
		errorRate = float64(m.errorCount) / float64(m.requestCount) * 100
	}

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Write Prometheus format
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	fmt.Fprintf(w, "# HELP cassocial_uptime_seconds Application uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE cassocial_uptime_seconds gauge\n")
	fmt.Fprintf(w, "cassocial_uptime_seconds %.2f\n\n", uptime)

	fmt.Fprintf(w, "# HELP cassocial_requests_total Total number of HTTP requests\n")
	fmt.Fprintf(w, "# TYPE cassocial_requests_total counter\n")
	fmt.Fprintf(w, "cassocial_requests_total %d\n\n", m.requestCount)

	fmt.Fprintf(w, "# HELP cassocial_errors_total Total number of errors\n")
	fmt.Fprintf(w, "# TYPE cassocial_errors_total counter\n")
	fmt.Fprintf(w, "cassocial_errors_total %d\n\n", m.errorCount)

	fmt.Fprintf(w, "# HELP cassocial_error_rate Error rate percentage\n")
	fmt.Fprintf(w, "# TYPE cassocial_error_rate gauge\n")
	fmt.Fprintf(w, "cassocial_error_rate %.2f\n\n", errorRate)

	fmt.Fprintf(w, "# HELP cassocial_request_duration_ms Average request duration in milliseconds\n")
	fmt.Fprintf(w, "# TYPE cassocial_request_duration_ms gauge\n")
	fmt.Fprintf(w, "cassocial_request_duration_ms %.2f\n\n", avgRequestDuration)

	fmt.Fprintf(w, "# HELP cassocial_db_queries_total Total number of database queries\n")
	fmt.Fprintf(w, "# TYPE cassocial_db_queries_total counter\n")
	fmt.Fprintf(w, "cassocial_db_queries_total %d\n\n", m.dbQueryCount)

	fmt.Fprintf(w, "# HELP cassocial_db_query_duration_ms Average DB query duration in milliseconds\n")
	fmt.Fprintf(w, "# TYPE cassocial_db_query_duration_ms gauge\n")
	fmt.Fprintf(w, "cassocial_db_query_duration_ms %.2f\n\n", avgDBQueryDuration)

	fmt.Fprintf(w, "# HELP cassocial_active_connections Current active connections\n")
	fmt.Fprintf(w, "# TYPE cassocial_active_connections gauge\n")
	fmt.Fprintf(w, "cassocial_active_connections %d\n\n", m.activeConnections)

	fmt.Fprintf(w, "# HELP cassocial_goroutines Current number of goroutines\n")
	fmt.Fprintf(w, "# TYPE cassocial_goroutines gauge\n")
	fmt.Fprintf(w, "cassocial_goroutines %d\n\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "# HELP cassocial_memory_alloc_bytes Memory allocated in bytes\n")
	fmt.Fprintf(w, "# TYPE cassocial_memory_alloc_bytes gauge\n")
	fmt.Fprintf(w, "cassocial_memory_alloc_bytes %d\n\n", memStats.Alloc)

	fmt.Fprintf(w, "# HELP cassocial_memory_sys_bytes Total memory obtained from OS in bytes\n")
	fmt.Fprintf(w, "# TYPE cassocial_memory_sys_bytes gauge\n")
	fmt.Fprintf(w, "cassocial_memory_sys_bytes %d\n\n", memStats.Sys)
}

// GetStats returns current metrics as a map
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"uptime_seconds":       uptime.Seconds(),
		"request_count":        m.requestCount,
		"error_count":          m.errorCount,
		"db_query_count":       m.dbQueryCount,
		"active_connections":   m.activeConnections,
		"goroutines":           runtime.NumGoroutine(),
		"memory_alloc_bytes":   memStats.Alloc,
		"memory_sys_bytes":     memStats.Sys,
		"cpu_count":            runtime.NumCPU(),
	}
}
