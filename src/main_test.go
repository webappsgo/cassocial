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
