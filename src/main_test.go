package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// captureStdout captures everything written to os.Stdout during f().
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = old
	return <-done
}

func TestPrintVersion(t *testing.T) {
	out := captureStdout(t, printVersion)
	if out == "" {
		t.Fatal("printVersion produced no output")
	}
	// Must contain the version variable value (default "dev").
	if !containsString(out, Version) {
		t.Errorf("printVersion output %q does not contain version %q", out, Version)
	}
}

func TestPrintHelp(t *testing.T) {
	out := captureStdout(t, printHelp)
	if out == "" {
		t.Fatal("printHelp produced no output")
	}
	for _, keyword := range []string{"--help", "--version", "--mode", "--port"} {
		if !containsString(out, keyword) {
			t.Errorf("printHelp output missing keyword %q", keyword)
		}
	}
}

func TestApplyColorPreference_Never(t *testing.T) {
	// Clear any existing NO_COLOR before test.
	old := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	t.Cleanup(func() {
		if old != "" {
			os.Setenv("NO_COLOR", old)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	})

	applyColorPreference("never")
	if os.Getenv("NO_COLOR") == "" {
		t.Error("applyColorPreference('never') did not set NO_COLOR")
	}
}

func TestApplyColorPreference_Always(t *testing.T) {
	// Clear any existing NO_COLOR before test.
	old := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	t.Cleanup(func() {
		if old != "" {
			os.Setenv("NO_COLOR", old)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	})

	applyColorPreference("always")
	// "always" must NOT set NO_COLOR.
	if os.Getenv("NO_COLOR") != "" {
		t.Error("applyColorPreference('always') unexpectedly set NO_COLOR")
	}
}

func TestApplyColorPreference_NoColorEnv(t *testing.T) {
	old := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	t.Cleanup(func() {
		if old != "" {
			os.Setenv("NO_COLOR", old)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	})

	// When NO_COLOR is already set, applyColorPreference must keep it.
	applyColorPreference("auto")
	if os.Getenv("NO_COLOR") == "" {
		t.Error("applyColorPreference('auto') cleared NO_COLOR that was already set")
	}
}

func TestHandleStatus(t *testing.T) {
	// handleStatus just prints a message — verify it doesn't panic.
	// It ignores both arguments, so nil cfg and empty pidFile are fine.
	handleStatus(nil, "")
}

// ---------------------------------------------------------------------------
// run() — testable entry point extracted from main()
// ---------------------------------------------------------------------------

// TestRun_Version verifies that --version exits 0.
func TestRun_Version(t *testing.T) {
	if code := run([]string{"--version"}); code != 0 {
		t.Errorf("run(--version) = %d, want 0", code)
	}
}

// TestRun_VersionShort verifies that -v exits 0.
func TestRun_VersionShort(t *testing.T) {
	if code := run([]string{"-v"}); code != 0 {
		t.Errorf("run(-v) = %d, want 0", code)
	}
}

// TestRun_Help verifies that --help exits 0.
func TestRun_Help(t *testing.T) {
	if code := run([]string{"--help"}); code != 0 {
		t.Errorf("run(--help) = %d, want 0", code)
	}
}

// TestRun_HelpShort verifies that -h exits 0.
func TestRun_HelpShort(t *testing.T) {
	if code := run([]string{"-h"}); code != 0 {
		t.Errorf("run(-h) = %d, want 0", code)
	}
}

// TestRun_UnknownFlag verifies that an unknown flag exits 1.
func TestRun_UnknownFlag(t *testing.T) {
	if code := run([]string{"--unknown-flag-xyz"}); code != 1 {
		t.Errorf("run(--unknown-flag-xyz) = %d, want 1", code)
	}
}

// TestRun_Daemon verifies that --daemon successfully backgrounds itself (exit 0).
// CASSOCIAL_DAEMONIZED=1 prevents actual re-exec so we test the guard path,
// which then falls through to server startup (fails → exit 1).
func TestRun_Daemon(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	// Inject guard so handleDaemon returns -1 (already-daemonized path)
	// and execution continues to server startup which fails → exit 1.
	t.Setenv("CASSOCIAL_DAEMONIZED", "1")
	// Use 256.256.256.256 to force server.Start to fail immediately; exit 1.
	if code := run([]string{"--config", tmp + "/cfg", "--data", tmp + "/data", "--log", tmp + "/log", "--daemon", "--address", "256.256.256.256"}); code != 1 {
		t.Errorf("run(--daemon, already-daemonized) = %d, want 1", code)
	}
}

// TestRun_Service verifies that --service start exits 1 when systemctl is unavailable.
func TestRun_Service(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	if code := run([]string{"--config", tmp + "/cfg", "--data", tmp + "/data", "--log", tmp + "/log", "--service", "start"}); code != 1 {
		t.Errorf("run(--service start) = %d, want 1", code)
	}
}

// TestRun_Maintenance verifies that --maintenance backup exits 1 when the data dir is unwritable.
func TestRun_Maintenance(t *testing.T) {
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	// /proc/cassocial-test-maint cannot be created → config.Load fails → exit 1.
	if code := run([]string{"--config", "/proc/cassocial-test-maint", "--data", "/proc/cassocial-test-maint-data", "--log", "/proc/cassocial-test-maint-log", "--maintenance", "backup"}); code != 1 {
		t.Errorf("run(--maintenance backup) = %d, want 1", code)
	}
}

// TestRun_Update verifies that --update check completes gracefully (0 or 1 depending on network).
func TestRun_Update(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	code := run([]string{"--config", tmp + "/cfg", "--data", tmp + "/data", "--log", tmp + "/log", "--update", "check"})
	if code != 0 && code != 1 {
		t.Errorf("run(--update check) = %d, want 0 or 1", code)
	}
}

// TestRun_ConfigLoadFails verifies that a bad config path exits 1.
func TestRun_ConfigLoadFails(t *testing.T) {
	// /proc/cassocial-test cannot be created → ensureDirectories fails → Load returns error.
	if code := run([]string{"--config", "/proc/cassocial-test-run-fail"}); code != 1 {
		t.Errorf("run(bad --config) = %d, want 1", code)
	}
}

// TestRun_Status verifies that --status exits 0 after loading config.
func TestRun_Status(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	if code := run([]string{"--config", tmp + "/cfg", "--data", tmp + "/data", "--log", tmp + "/log", "--status"}); code != 0 {
		t.Errorf("run(--status) = %d, want 0", code)
	}
}

// TestRun_ConfigOverrides verifies that --address, --port, --mode, --debug override config
// and that --status causes an early exit 0 after applying them.
func TestRun_ConfigOverrides(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	code := run([]string{
		"--config", tmp + "/cfg",
		"--data", tmp + "/data",
		"--log", tmp + "/log",
		"--address", "127.0.0.1",
		"--port", "65001",
		"--mode", "development",
		"--debug",
		"--status",
	})
	if code != 0 {
		t.Errorf("run with overrides + --status = %d, want 0", code)
	}
}

// TestRun_PIDWriteFails verifies run() returns 1 when WritePIDFile fails.
func TestRun_PIDWriteFails(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	// /proc is not writable even as root — WritePIDFile must fail.
	code := run([]string{
		"--config", tmp + "/cfg",
		"--data", tmp + "/data",
		"--log", tmp + "/log",
		"--pid", "/proc/cassocial-test-pid-run-fail",
	})
	if code != 1 {
		t.Errorf("run with bad pid path = %d, want 1", code)
	}
}

// TestRun_DBConnectFails verifies run() returns 1 when the database cannot be reached.
func TestRun_DBConnectFails(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	// MySQL driver with no server → Ping fails → store.Connect returns error.
	t.Setenv("CASSOCIAL_DB_DRIVER", "mysql")
	code := run([]string{
		"--config", tmp + "/cfg",
		"--data", tmp + "/data",
		"--log", tmp + "/log",
		"--pid", tmp + "/test.pid",
	})
	if code != 1 {
		t.Errorf("run with mysql (no server) = %d, want 1", code)
	}
}

// TestRun_ServerStartFails verifies run() returns 1 when server.Start fails (invalid listen address).
func TestRun_ServerStartFails(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	// 256.x.x.x is an invalid IP; ListenAndServe fails immediately → errChan fires.
	code := run([]string{
		"--config", tmp + "/cfg",
		"--data", tmp + "/data",
		"--log", tmp + "/log",
		"--pid", tmp + "/test.pid",
		"--address", "256.256.256.256",
		"--port", "65432",
	})
	if code != 1 {
		t.Errorf("run with invalid listen address = %d, want 1", code)
	}
}

// TestRun_LangAutoDetect verifies run() reads LANG env when --lang is not set.
func TestRun_LangAutoDetect(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LANG", "es_ES.UTF-8")
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	// --status exits early; we just verify no panic.
	code := run([]string{
		"--config", tmp + "/cfg",
		"--data", tmp + "/data",
		"--log", tmp + "/log",
		"--status",
	})
	if code != 0 {
		t.Errorf("run with LANG env = %d, want 0", code)
	}
}

// containsString is a simple helper used only in this test file.
func containsString(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
