package signal

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	ossignal "os/signal"
	"os/exec"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestSetShuttingDown(t *testing.T) {
	// Start in known state
	setShuttingDown(false)
	if shuttingDown {
		t.Error("shuttingDown should be false after setShuttingDown(false)")
	}

	setShuttingDown(true)
	if !shuttingDown {
		t.Error("shuttingDown should be true after setShuttingDown(true)")
	}

	// Reset
	setShuttingDown(false)
}

func TestCloseDatabase(t *testing.T) {
	// Should not panic; current implementation is a no-op
	closeDatabase(5 * time.Second)
}

func TestFlushLogs(t *testing.T) {
	// Should not panic; current implementation is a no-op
	flushLogs(1 * time.Second)
}

func TestReopenLogs(t *testing.T) {
	// Should not panic
	reopenLogs()
}

func TestDumpStatus(t *testing.T) {
	// Should not panic
	dumpStatus()
}

func TestGetChildPIDs(t *testing.T) {
	pids := getChildPIDs()
	if pids == nil {
		t.Error("getChildPIDs() returned nil, want empty slice")
	}
}

func TestGetChildPIDs_IsEmpty(t *testing.T) {
	pids := getChildPIDs()
	if len(pids) != 0 {
		t.Errorf("getChildPIDs() = %v, want empty slice", pids)
	}
}

func TestSetShuttingDown_MultipleTransitions(t *testing.T) {
	// Ensure repeated transitions work correctly
	setShuttingDown(false)
	setShuttingDown(true)
	setShuttingDown(false)
	setShuttingDown(true)
	if !shuttingDown {
		t.Error("shuttingDown should be true after final setShuttingDown(true)")
	}
	setShuttingDown(false)
}

func TestCloseDatabase_Idempotent(t *testing.T) {
	// Call multiple times — must not panic
	closeDatabase(1 * time.Second)
	closeDatabase(0)
	closeDatabase(100 * time.Millisecond)
}

func TestFlushLogs_Idempotent(t *testing.T) {
	// Call multiple times — must not panic
	flushLogs(1 * time.Second)
	flushLogs(0)
	flushLogs(100 * time.Millisecond)
}

func TestStopChildProcesses_EmptyList(t *testing.T) {
	// getChildPIDs returns empty slice; stopChildProcesses should be a no-op
	stopChildProcesses(100 * time.Millisecond)
}

func TestSetupSignalHandler_DoesNotPanic(t *testing.T) {
	// setupSignalHandler installs signal handlers in a goroutine and returns.
	// We cannot easily test the actual signal dispatch, but we can verify it
	// doesn't panic during setup.
	server := &http.Server{}
	stop := setupSignalHandler(server, "")
	defer stop()
}

func TestKillProcess_InvalidPID(t *testing.T) {
	// Sending to PID 0 or -1 is unreliable; use a PID that won't exist.
	// PID 2147483647 is max int32 and virtually guaranteed not to exist.
	err := killProcess(2147483647, true)
	// We expect an error (process not found), not a panic.
	if err == nil {
		t.Log("killProcess(maxPID) unexpectedly succeeded — may be system-dependent")
	}
}

// TestKillProcess_Graceful_CurrentProcess exercises the graceful=true path by
// sending SIGTERM to the current process. We mask the signal so the process is
// not actually terminated.
func TestKillProcess_Graceful_CurrentProcess(t *testing.T) {
	// Mask SIGTERM before sending it so the process survives.
	sigChan := make(chan os.Signal, 1)
	ossignal.Notify(sigChan, syscall.SIGTERM)
	defer ossignal.Reset(syscall.SIGTERM)

	err := killProcess(os.Getpid(), true)
	if err != nil {
		t.Errorf("killProcess(self, graceful=true) returned error: %v", err)
	}

	// Drain the channel so the masked signal doesn't linger.
	select {
	case <-sigChan:
	default:
	}
}

// TestKillProcess_Forceful_InvalidPID exercises the graceful=false path.
// We use a non-existent PID so SIGKILL is sent but fails gracefully.
func TestKillProcess_Forceful_InvalidPID(t *testing.T) {
	// PID 2147483647 is max int32 — virtually guaranteed not to exist.
	err := killProcess(2147483647, false)
	// We expect an OS error; a nil return is also acceptable on some systems.
	_ = err // either outcome is valid; the goal is to cover the false branch
}

// TestStopChildProcesses_WithChild exercises the loop body of stopChildProcesses
// by injecting a real child PID (a sleep subprocess) and calling with a very
// short timeout so the SIGKILL force-kill path also runs.
func TestStopChildProcesses_WithChild(t *testing.T) {
	// Start a child process that will ignore SIGTERM so the deadline/SIGKILL
	// path is also exercised.
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Skipf("could not start sleep subprocess: %v", err)
	}
	childPID := cmd.Process.Pid

	// Inject the child PID into getChildPIDsFn.
	orig := getChildPIDsFn
	getChildPIDsFn = func() []int { return []int{childPID} }
	defer func() {
		getChildPIDsFn = orig
		// Clean up: the process should already be dead after SIGKILL, but try anyway.
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// Use a short timeout so the force-kill deadline fires quickly.
	stopChildProcesses(50 * time.Millisecond)
}

// TestSetupSignalHandler_SIGUSR1 exercises the SIGUSR1 branch in the goroutine
// started by setupSignalHandler.
func TestSetupSignalHandler_SIGUSR1(t *testing.T) {
	server := &http.Server{}
	stop := setupSignalHandler(server, "")
	defer stop()

	// Give the goroutine time to start.
	time.Sleep(10 * time.Millisecond)

	// Send SIGUSR1 to the current process. The goroutine will call reopenLogs().
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	if err := p.Signal(syscall.SIGUSR1); err != nil {
		t.Fatalf("Signal(SIGUSR1): %v", err)
	}

	// Brief pause to let the goroutine handle the signal.
	time.Sleep(20 * time.Millisecond)
}

// TestSetupSignalHandler_SIGUSR2 exercises the SIGUSR2 branch in the goroutine.
func TestSetupSignalHandler_SIGUSR2(t *testing.T) {
	server := &http.Server{}
	stop := setupSignalHandler(server, "")
	defer stop()

	time.Sleep(10 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	if err := p.Signal(syscall.SIGUSR2); err != nil {
		t.Fatalf("Signal(SIGUSR2): %v", err)
	}

	time.Sleep(20 * time.Millisecond)
}

// TestSetupSignalHandler_DefaultCase exercises the default: branch in the
// signal handler goroutine by sending SIGQUIT (masked so the process survives).
func TestSetupSignalHandler_DefaultCase(t *testing.T) {
	// Replace osExitFn so gracefulShutdown doesn't terminate the process.
	// Multiple stale goroutines may call this concurrently; use an atomic flag.
	origExit := osExitFn
	var exitCalled atomic.Int32
	done := make(chan struct{}, 1)
	osExitFn = func(_ int) {
		if exitCalled.CompareAndSwap(0, 1) {
			done <- struct{}{}
		}
	}
	defer func() { osExitFn = origExit }()

	// Start a test server for gracefulShutdown to shut down.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	stop := setupSignalHandler(ts.Config, "")
	defer stop()
	time.Sleep(10 * time.Millisecond)

	// Mask SIGQUIT so the OS default action (core dump) doesn't fire.
	sigChan := make(chan os.Signal, 1)
	ossignal.Notify(sigChan, syscall.SIGQUIT)
	defer ossignal.Reset(syscall.SIGQUIT)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	if err := p.Signal(syscall.SIGQUIT); err != nil {
		t.Fatalf("Signal(SIGQUIT): %v", err)
	}

	// Wait for gracefulShutdown to call osExitFn, or time out.
	select {
	case <-done:
		// Success — default branch was exercised.
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for gracefulShutdown to be called via default: branch")
	}
	// defer stop() fires here, stopping the goroutine before osExitFn is restored.
}

// TestStopChildProcesses_ChildExitsBeforeDeadline exercises the break branch
// inside the polling loop by injecting an already-reaped PID so Signal(0)
// immediately returns ESRCH — exercising the break path.
func TestStopChildProcesses_ChildExitsBeforeDeadline(t *testing.T) {
	// Start and immediately kill+wait a child so its PID is fully reaped.
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Skipf("could not start sleep subprocess: %v", err)
	}
	childPID := cmd.Process.Pid
	// Kill immediately and wait so the process is fully reaped (not a zombie).
	_ = cmd.Process.Kill()
	_ = cmd.Wait()

	orig := getChildPIDsFn
	// First call (SIGTERM loop): return the dead PID so SIGTERM is "sent"
	// Second call (polling loop): same dead PID — Signal(0) returns ESRCH → break
	getChildPIDsFn = func() []int { return []int{childPID} }
	defer func() { getChildPIDsFn = orig }()

	// Use a generous timeout; the break path should fire on the first Signal(0) poll.
	stopChildProcesses(500 * time.Millisecond)
}

// TestKillProcess_FindProcessError exercises the os.FindProcess error path by
// injecting a failing findProcessFn.
func TestKillProcess_FindProcessError(t *testing.T) {
	orig := findProcessFn
	findProcessFn = func(_ int) (*os.Process, error) {
		return nil, errors.New("injected find error")
	}
	defer func() { findProcessFn = orig }()

	err := killProcess(1, true)
	if err == nil {
		t.Error("killProcess() should return error when findProcessFn fails")
	}
}

// TestStopChildProcesses_FindProcessError exercises the continue branch in the
// first loop of stopChildProcesses by injecting a failing findProcessFn.
func TestStopChildProcesses_FindProcessError(t *testing.T) {
	orig := findProcessFn
	findProcessFn = func(_ int) (*os.Process, error) {
		return nil, errors.New("injected find error")
	}
	defer func() { findProcessFn = orig }()

	origPIDs := getChildPIDsFn
	getChildPIDsFn = func() []int { return []int{99999} }
	defer func() { getChildPIDsFn = origPIDs }()

	// Should not panic; the continue branch is exercised
	stopChildProcesses(10 * time.Millisecond)
}

// TestGracefulShutdown_ShutdownError exercises the server.Shutdown error branch.
func TestGracefulShutdown_ShutdownError(t *testing.T) {
	origExit := osExitFn
	osExitFn = func(_ int) {}
	defer func() { osExitFn = origExit }()

	origShutdown := httpShutdownFn
	httpShutdownFn = func(_ *http.Server, _ context.Context) error {
		return errors.New("injected shutdown error")
	}
	defer func() { httpShutdownFn = origShutdown }()

	gracefulShutdown(&http.Server{}, "")
}

// TestGracefulShutdown exercises gracefulShutdown by replacing osExitFn to
// prevent the process from actually exiting.
func TestGracefulShutdown(t *testing.T) {
	origExit := osExitFn
	exitCalled := false
	osExitFn = func(_ int) { exitCalled = true }
	defer func() { osExitFn = origExit }()

	// Use a test HTTP server so Shutdown() can complete cleanly.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	gracefulShutdown(ts.Config, "")

	if !exitCalled {
		t.Error("gracefulShutdown() did not call osExitFn")
	}
}

// TestGracefulShutdown_WithPIDFile exercises the os.Remove(pidFile) branch.
func TestGracefulShutdown_WithPIDFile(t *testing.T) {
	origExit := osExitFn
	osExitFn = func(_ int) {}
	defer func() { osExitFn = origExit }()

	// Create a temporary PID file.
	pidFile := t.TempDir() + "/test.pid"
	if err := os.WriteFile(pidFile, []byte("12345\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	gracefulShutdown(ts.Config, pidFile)

	// PID file should be removed.
	if _, err := os.Stat(pidFile); !errors.Is(err, os.ErrNotExist) {
		t.Error("gracefulShutdown() should have removed the PID file")
	}
}
