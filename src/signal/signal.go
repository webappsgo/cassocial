package signal

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

// osExitFn is the function used to exit the process.
// Tests may replace it to prevent os.Exit from terminating the test binary.
var osExitFn = os.Exit

// httpShutdownFn is the function used to shut down the HTTP server.
// Tests may replace it to inject a shutdown error.
var httpShutdownFn = func(server *http.Server, ctx context.Context) error {
	return server.Shutdown(ctx)
}

// gracefulShutdown performs orderly shutdown (cross-platform)
func gracefulShutdown(server *http.Server, pidFile string) {
	// Set shutdown flag for health checks
	setShuttingDown(true)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop accepting new connections, wait for in-flight
	if err := httpShutdownFn(server, ctx); err != nil {
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
	osExitFn(0)
}

// shuttingDown is the global shutdown flag.
var shuttingDown bool

// setShuttingDown sets the global shutdown flag
func setShuttingDown(shutting bool) {
	shuttingDown = shutting
}

// closeDatabase closes database connections with timeout
func closeDatabase(_ time.Duration) {
	log.Println("closeDatabase: connections managed by store package")
}

// flushLogs flushes log buffers
func flushLogs(_ time.Duration) {
	log.Println("flushLogs: standard log package is unbuffered")
}

// reopenLogs reopens log files (for log rotation)
func reopenLogs() {
	log.Println("Log reopen requested (SIGHUP)")
}

// dumpStatus dumps current status to logs
func dumpStatus() {
	log.Printf("status: shutting_down=%v", shuttingDown)
}

// getChildPIDsFn is the function used to retrieve child PIDs.
// Tests may replace it to inject a fake PID list.
var getChildPIDsFn = func() []int { return []int{} }

// getChildPIDs returns list of child process PIDs
func getChildPIDs() []int {
	return getChildPIDsFn()
}
