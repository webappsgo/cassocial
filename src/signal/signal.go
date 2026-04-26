package signal

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

// gracefulShutdown performs orderly shutdown (cross-platform)
func gracefulShutdown(server *http.Server, pidFile string) {
	// Set shutdown flag for health checks
	setShuttingDown(true)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop accepting new connections, wait for in-flight
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Stop child processes (Tor, etc.) - platform-specific
	stopChildProcesses(10 * time.Second)

	// Close database connections
	closeDatabase(5 * time.Second)

	// Flush logs
	flushLogs(2 * time.Second)

	// Remove PID file
	if pidFile != "" {
		os.Remove(pidFile)
	}

	log.Println("Graceful shutdown complete")
	os.Exit(0)
}

// setShuttingDown sets the global shutdown flag
func setShuttingDown(shutting bool) {
	// TODO: Set server shutdown flag
}

// closeDatabase closes database connections with timeout
func closeDatabase(timeout time.Duration) {
	// TODO: Close database connections
}

// flushLogs flushes log buffers
func flushLogs(timeout time.Duration) {
	// TODO: Flush log buffers
}

// reopenLogs reopens log files (for log rotation)
func reopenLogs() {
	// TODO: Reopen log files
}

// dumpStatus dumps current status to logs
func dumpStatus() {
	// TODO: Dump status information
}

// getChildPIDs returns list of child process PIDs
func getChildPIDs() []int {
	// TODO: Track child processes (Tor, etc.)
	return []int{}
}
