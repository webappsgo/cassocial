package paths

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestResolve_ReturnsNonNilPaths(t *testing.T) {
	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() returned nil")
	}
}

func TestResolve_AllFieldsNonEmpty(t *testing.T) {
	p := Resolve()
	fields := map[string]string{
		"Config":   p.Config,
		"Data":     p.Data,
		"Log":      p.Log,
		"Backup":   p.Backup,
		"PID":      p.PID,
		"SSL":      p.SSL,
		"Security": p.Security,
		"Database": p.Database,
	}
	for name, val := range fields {
		if val == "" {
			t.Errorf("Resolve().%s is empty, want non-empty path", name)
		}
	}
}

func TestResolve_ContainsProjectName(t *testing.T) {
	// When running in Docker, paths are fixed (/config, /data, etc.) and do not
	// contain the project name — that is correct Docker behavior. Skip the check
	// if we detect we're in a container.
	if isRunningInDocker() {
		t.Skip("Docker paths do not embed the project name by design")
	}
	p := Resolve()
	// At least one path should contain the project name or org
	paths := []string{p.Config, p.Data, p.Log}
	found := false
	for _, path := range paths {
		if strings.Contains(path, "cassocial") || strings.Contains(path, "casapps") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Resolve() paths do not contain project name or org: Config=%s Data=%s Log=%s",
			p.Config, p.Data, p.Log)
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := GetConfigDir()
	if dir == "" {
		t.Error("GetConfigDir() returned empty string")
	}
}

func TestGetDataDir(t *testing.T) {
	dir := GetDataDir()
	if dir == "" {
		t.Error("GetDataDir() returned empty string")
	}
}

func TestGetLogDir(t *testing.T) {
	dir := GetLogDir()
	if dir == "" {
		t.Error("GetLogDir() returned empty string")
	}
}

func TestGetBackupDir(t *testing.T) {
	dir := GetBackupDir()
	if dir == "" {
		t.Error("GetBackupDir() returned empty string")
	}
}

func TestGetPIDFile(t *testing.T) {
	f := GetPIDFile()
	if f == "" {
		t.Error("GetPIDFile() returned empty string")
	}
	if !strings.HasSuffix(f, ".pid") {
		t.Errorf("GetPIDFile() = %q, want path ending in .pid", f)
	}
}

func TestGetSSLDir(t *testing.T) {
	dir := GetSSLDir()
	if dir == "" {
		t.Error("GetSSLDir() returned empty string")
	}
}

func TestGetSecurityDir(t *testing.T) {
	dir := GetSecurityDir()
	if dir == "" {
		t.Error("GetSecurityDir() returned empty string")
	}
}

func TestGetDatabaseDir(t *testing.T) {
	dir := GetDatabaseDir()
	if dir == "" {
		t.Error("GetDatabaseDir() returned empty string")
	}
}

func TestLinuxPaths_Root(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}
	p := linuxPaths(true)
	if !strings.HasPrefix(p.Config, "/etc/") {
		t.Errorf("linuxPaths(root).Config = %q, want /etc/... prefix", p.Config)
	}
	if !strings.HasPrefix(p.Log, "/var/log/") {
		t.Errorf("linuxPaths(root).Log = %q, want /var/log/... prefix", p.Log)
	}
	if !strings.HasSuffix(p.PID, ".pid") {
		t.Errorf("linuxPaths(root).PID = %q, want .pid suffix", p.PID)
	}
}

func TestLinuxPaths_User(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}
	p := linuxPaths(false)
	homeDir, _ := os.UserHomeDir()
	if !strings.HasPrefix(p.Config, homeDir) {
		t.Errorf("linuxPaths(user).Config = %q, want prefix %q", p.Config, homeDir)
	}
}

func TestDockerPaths(t *testing.T) {
	p := dockerPaths()
	if p.Config != "/config" {
		t.Errorf("dockerPaths().Config = %q, want /config", p.Config)
	}
	if p.Data != "/data" {
		t.Errorf("dockerPaths().Data = %q, want /data", p.Data)
	}
}

func TestIsRunningInDocker(t *testing.T) {
	// In the test environment, result depends on whether running in Docker.
	// Just ensure the function returns a bool without panicking.
	_ = isRunningInDocker()
}

func TestIsPrivileged(t *testing.T) {
	// Just ensure the function runs without panic.
	_ = isPrivileged()
}
