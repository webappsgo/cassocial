package server

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// TorService manages Tor hidden service
type TorService struct {
	dataDir    string
	torDataDir string
	torrcPath  string
	cmd        *exec.Cmd
	onionAddr  string
	enabled    bool
}

// NewTorService creates a new Tor service
func NewTorService(dataDir string) *TorService {
	torDataDir := filepath.Join(dataDir, "tor")
	torrcPath := filepath.Join(dataDir, "tor", "torrc")

	// Check if tor binary is installed
	enabled := false
	if _, err := exec.LookPath("tor"); err == nil {
		enabled = true
	}

	return &TorService{
		dataDir:    dataDir,
		torDataDir: torDataDir,
		torrcPath:  torrcPath,
		enabled:    enabled,
	}
}

// Start starts the Tor hidden service
func (t *TorService) Start() error {
	if !t.enabled {
		log.Println("Tor not installed, skipping hidden service")
		return nil
	}

	log.Println("Starting Tor hidden service...")

	// Ensure directory exists
	if err := os.MkdirAll(t.torDataDir, 0700); err != nil {
		return fmt.Errorf("failed to create tor directory: %w", err)
	}

	// Create torrc if not exists
	if err := t.createTorrc(); err != nil {
		return fmt.Errorf("failed to create torrc: %w", err)
	}

	// Start Tor process
	t.cmd = exec.Command("tor", "-f", t.torrcPath)

	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tor: %w", err)
	}

	log.Printf("Tor started (PID: %d)", t.cmd.Process.Pid)

	// Wait for .onion address to be generated
	if err := t.waitForOnionAddress(); err != nil {
		return fmt.Errorf("failed to get .onion address: %w", err)
	}

	log.Printf("Tor hidden service: %s", t.onionAddr)

	return nil
}

// Stop stops the Tor service
func (t *TorService) Stop() error {
	if t.cmd == nil || t.cmd.Process == nil {
		return nil
	}

	log.Println("Stopping Tor service...")

	// Send SIGTERM for graceful shutdown
	if err := t.cmd.Process.Signal(os.Interrupt); err != nil {
		// Force kill if graceful fails
		t.cmd.Process.Kill()
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- t.cmd.Wait()
	}()

	select {
	case <-done:
		log.Println("Tor stopped")
	case <-time.After(10 * time.Second):
		log.Println("Tor shutdown timeout, force killing")
		t.cmd.Process.Kill()
	}

	return nil
}

// GetOnionAddress returns the .onion address
func (t *TorService) GetOnionAddress() string {
	return t.onionAddr
}

// IsEnabled returns whether Tor is enabled
func (t *TorService) IsEnabled() bool {
	return t.enabled && t.onionAddr != ""
}

// createTorrc creates the torrc configuration file
func (t *TorService) createTorrc() error {
	hiddenServiceDir := filepath.Join(t.torDataDir, "hidden_service")
	logFile := filepath.Join(t.torDataDir, "tor.log")

	torrc := fmt.Sprintf(`# Tor configuration for Cassocial
DataDirectory %s
HiddenServiceDir %s
HiddenServicePort 80 127.0.0.1:80
Log notice file %s
`, t.torDataDir, hiddenServiceDir, logFile)

	return os.WriteFile(t.torrcPath, []byte(torrc), 0600)
}

// waitForOnionAddress waits for Tor to generate .onion address
func (t *TorService) waitForOnionAddress() error {
	hostnameFile := filepath.Join(t.torDataDir, "hidden_service", "hostname")

	// Wait up to 60 seconds
	for i := 0; i < 60; i++ {
		if data, err := os.ReadFile(hostnameFile); err == nil {
			t.onionAddr = string(data)
			t.onionAddr = filepath.Base(t.onionAddr) // Trim whitespace
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for .onion address")
}

// GenerateVanityAddress generates a vanity .onion address
func (t *TorService) GenerateVanityAddress(prefix string) error {
	// TODO: Implement vanity address generation
	// This requires generating multiple keypairs until finding one with desired prefix
	log.Printf("Vanity address generation for prefix '%s' not yet implemented", prefix)
	return fmt.Errorf("vanity address generation not implemented")
}

// GetStatus returns Tor service status
func (t *TorService) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":      t.enabled,
		"running":      t.cmd != nil && t.cmd.Process != nil,
		"onion_address": t.onionAddr,
	}

	if t.cmd != nil && t.cmd.Process != nil {
		status["pid"] = t.cmd.Process.Pid
	}

	return status
}
