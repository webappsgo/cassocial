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

// TestRun_Daemon verifies that --daemon exits 1 (not yet implemented).
func TestRun_Daemon(t *testing.T) {
	// run() will try to load config before reaching --daemon; supply a
	// temp config dir so Load succeeds, then expect 1 from the daemon stub.
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	if code := run([]string{"--config", tmp + "/cfg", "--data", tmp + "/data", "--log", tmp + "/log", "--daemon"}); code != 1 {
		t.Errorf("run(--daemon) = %d, want 1", code)
	}
}

// TestRun_Service verifies that --service exits 1 (not yet implemented).
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

// TestRun_Maintenance verifies that --maintenance exits 1 (not yet implemented).
func TestRun_Maintenance(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	if code := run([]string{"--config", tmp + "/cfg", "--data", tmp + "/data", "--log", tmp + "/log", "--maintenance", "backup"}); code != 1 {
		t.Errorf("run(--maintenance backup) = %d, want 1", code)
	}
}

// TestRun_Update verifies that --update exits 1 (not yet implemented).
func TestRun_Update(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CASSOCIAL_PORT", "")
	t.Setenv("CASSOCIAL_MODE", "")
	t.Setenv("CASSOCIAL_ADDRESS", "")
	t.Setenv("CASSOCIAL_DB_DRIVER", "")
	if code := run([]string{"--config", tmp + "/cfg", "--data", tmp + "/data", "--log", tmp + "/log", "--update", "check"}); code != 1 {
		t.Errorf("run(--update check) = %d, want 1", code)
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
