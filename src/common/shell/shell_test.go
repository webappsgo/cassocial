package shell

import (
	"bytes"
	"strings"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		want  string
	}{
		{"empty falls back to bash", "", "bash"},
		{"unknown falls back to bash", "/usr/bin/tcsh", "bash"},
		{"bash", "/bin/bash", "bash"},
		{"zsh", "/usr/bin/zsh", "zsh"},
		{"fish", "/usr/local/bin/fish", "fish"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.shell)
			if got := Detect(); got != tt.want {
				t.Errorf("Detect() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	for _, sh := range Supported {
		if !isSupported(sh) {
			t.Errorf("isSupported(%q) = false, want true", sh)
		}
	}
	if isSupported("tcsh") {
		t.Error("isSupported(\"tcsh\") = true, want false")
	}
}

func TestHandle(t *testing.T) {
	flags := []string{"--help", "--version", "--shell", "--debug", "--color", "--lang"}

	t.Run("completions with explicit shell", func(t *testing.T) {
		var buf bytes.Buffer
		code := Handle("myapp", "completions", []string{"zsh"}, flags, &buf)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
		if !strings.Contains(buf.String(), "#compdef myapp") {
			t.Errorf("output missing zsh compdef header: %q", buf.String())
		}
	})

	t.Run("init with explicit shell", func(t *testing.T) {
		var buf bytes.Buffer
		code := Handle("myapp", "init", []string{"fish"}, flags, &buf)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
		if !strings.Contains(buf.String(), "myapp --shell completions fish | source") {
			t.Errorf("unexpected fish init output: %q", buf.String())
		}
	})

	t.Run("auto-detect shell", func(t *testing.T) {
		t.Setenv("SHELL", "/bin/bash")
		var buf bytes.Buffer
		code := Handle("myapp", "completions", nil, flags, &buf)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
		if !strings.Contains(buf.String(), "bash completion for myapp") {
			t.Errorf("unexpected bash completions output: %q", buf.String())
		}
	})

	t.Run("unsupported shell", func(t *testing.T) {
		var buf bytes.Buffer
		code := Handle("myapp", "completions", []string{"tcsh"}, flags, &buf)
		if code != 1 {
			t.Fatalf("exit code = %d, want 1", code)
		}
		if !strings.Contains(buf.String(), "unsupported shell") {
			t.Errorf("unexpected error output: %q", buf.String())
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		var buf bytes.Buffer
		code := Handle("myapp", "bogus", []string{"bash"}, flags, &buf)
		if code != 1 {
			t.Fatalf("exit code = %d, want 1", code)
		}
		if !strings.Contains(buf.String(), "unknown --shell command") {
			t.Errorf("unexpected error output: %q", buf.String())
		}
	})
}

func TestCompletions(t *testing.T) {
	flags := []string{"--help", "--debug"}

	tests := []struct {
		shell string
		want  string
	}{
		{"bash", "_myapp_completions"},
		{"zsh", "#compdef myapp"},
		{"fish", "complete -c myapp -l help"},
		{"sh", "MYAPP_OPTS="},
		{"dash", "MYAPP_OPTS="},
		{"ksh", "MYAPP_OPTS="},
		{"powershell", "Register-ArgumentCompleter"},
		{"pwsh", "Register-ArgumentCompleter"},
		{"unknown-shell", "_myapp_completions"},
	}
	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got := Completions("myapp", tt.shell, flags)
			if !strings.Contains(got, tt.want) {
				t.Errorf("Completions(%q) = %q, want substring %q", tt.shell, got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		shell string
		want  string
	}{
		{"bash", "source <(myapp --shell completions bash)\n"},
		{"zsh", "source <(myapp --shell completions zsh)\n"},
		{"fish", "myapp --shell completions fish | source\n"},
		{"powershell", "Invoke-Expression (& myapp --shell completions powershell)\n"},
		{"pwsh", "Invoke-Expression (& myapp --shell completions pwsh)\n"},
		{"sh", "eval \"$(myapp --shell completions sh)\"\n"},
		{"dash", "eval \"$(myapp --shell completions dash)\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			if got := Init("myapp", tt.shell); got != tt.want {
				t.Errorf("Init(%q) = %q, want %q", tt.shell, got, tt.want)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	if got := sanitize("cas-social.cli"); got != "cas_social_cli" {
		t.Errorf("sanitize() = %q, want %q", got, "cas_social_cli")
	}
}

func TestQuotedList(t *testing.T) {
	got := quotedList([]string{"--help", "--debug"})
	want := "'--help', '--debug'"
	if got != want {
		t.Errorf("quotedList() = %q, want %q", got, want)
	}
}
