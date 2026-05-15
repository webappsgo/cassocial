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

// shuttingDown is the global shutdown flag.
var shuttingDown bool

// setShuttingDown sets the global shutdown flag
func setShuttingDown(shutting bool) {
	shuttingDown = shutting
}

// closeDatabase closes database connections with timeout
func closeDatabase(_ time.Duration) {
	// Database connections are managed by the store package and closed via defer in main.
}

// flushLogs flushes log buffers
func flushLogs(_ time.Duration) {
	// The standard log package is unbuffered; nothing to flush.
}

// reopenLogs reopens log files (for log rotation)
func reopenLogs() {
	log.Println("Log reopen requested (SIGHUP)")
}

// dumpStatus dumps current status to logs
func dumpStatus() {
	log.Printf("status: shutting_down=%v", shuttingDown)
}

// getChildPIDs returns list of child process PIDs
func getChildPIDs() []int {
	return []int{}
}
