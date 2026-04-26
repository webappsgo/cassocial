//go:build windows
// +build windows

package signal

import (
	"log"
	"net/http"
	"os"
	"os/signal"
)

// setupSignalHandler configures graceful shutdown (Windows)
// Windows only supports os.Interrupt (Ctrl+C, Ctrl+Break)
func setupSignalHandler(server *http.Server, pidFile string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		for sig := range sigChan {
			log.Printf("Received %v, starting graceful shutdown...", sig)
			gracefulShutdown(server, pidFile)
		}
	}()
}

// killProcess terminates process (Windows)
// Windows doesn't have graceful signals - uses TerminateProcess
func killProcess(pid int, graceful bool) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	// Windows: Kill() calls TerminateProcess - no graceful option
	return process.Kill()
}
