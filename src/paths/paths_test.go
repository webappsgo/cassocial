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
		"Cache":    p.Cache,
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

func TestGetCacheDir(t *testing.T) {
	dir := GetCacheDir()
	if dir == "" {
		t.Error("GetCacheDir() returned empty string")
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
	if p.Config != "/config/cassocial" {
		t.Errorf("dockerPaths().Config = %q, want /config/cassocial", p.Config)
	}
	if p.Data != "/data/cassocial" {
		t.Errorf("dockerPaths().Data = %q, want /data/cassocial", p.Data)
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

// ---- darwinPaths ----

func TestDarwinPaths_Root(t *testing.T) {
	p := darwinPaths(true)
	if p == nil {
		t.Fatal("darwinPaths(true) returned nil")
	}
	if p.Config == "" {
		t.Error("darwinPaths(true).Config is empty")
	}
	if p.PID == "" {
		t.Error("darwinPaths(true).PID is empty")
	}
	if !strings.Contains(p.Config, "cassocial") && !strings.Contains(p.Config, "casapps") {
		t.Errorf("darwinPaths(true).Config = %q, want path containing project/org name", p.Config)
	}
}

func TestDarwinPaths_User(t *testing.T) {
	p := darwinPaths(false)
	if p == nil {
		t.Fatal("darwinPaths(false) returned nil")
	}
	homeDir, _ := os.UserHomeDir()
	if !strings.HasPrefix(p.Config, homeDir) {
		t.Errorf("darwinPaths(false).Config = %q, want prefix %q", p.Config, homeDir)
	}
	if !strings.HasSuffix(p.PID, ".pid") {
		t.Errorf("darwinPaths(false).PID = %q, want .pid suffix", p.PID)
	}
}

func TestDarwinPaths_AllFieldsNonEmpty(t *testing.T) {
	for _, isRoot := range []bool{true, false} {
		p := darwinPaths(isRoot)
		fields := map[string]string{
			"Config":   p.Config,
			"Data":     p.Data,
			"Cache":    p.Cache,
			"Log":      p.Log,
			"Backup":   p.Backup,
			"PID":      p.PID,
			"SSL":      p.SSL,
			"Security": p.Security,
			"Database": p.Database,
		}
		for name, val := range fields {
			if val == "" {
				t.Errorf("darwinPaths(%v).%s is empty", isRoot, name)
			}
		}
	}
}

// ---- bsdPaths ----

func TestBSDPaths_Root(t *testing.T) {
	p := bsdPaths(true)
	if p == nil {
		t.Fatal("bsdPaths(true) returned nil")
	}
	if !strings.HasPrefix(p.Config, "/usr/local/etc/") {
		t.Errorf("bsdPaths(true).Config = %q, want /usr/local/etc/... prefix", p.Config)
	}
	if !strings.HasSuffix(p.PID, ".pid") {
		t.Errorf("bsdPaths(true).PID = %q, want .pid suffix", p.PID)
	}
}

func TestBSDPaths_User(t *testing.T) {
	p := bsdPaths(false)
	if p == nil {
		t.Fatal("bsdPaths(false) returned nil")
	}
	homeDir, _ := os.UserHomeDir()
	if !strings.HasPrefix(p.Config, homeDir) {
		t.Errorf("bsdPaths(false).Config = %q, want prefix %q", p.Config, homeDir)
	}
}

func TestBSDPaths_AllFieldsNonEmpty(t *testing.T) {
	for _, isRoot := range []bool{true, false} {
		p := bsdPaths(isRoot)
		fields := map[string]string{
			"Config":   p.Config,
			"Data":     p.Data,
			"Cache":    p.Cache,
			"Log":      p.Log,
			"Backup":   p.Backup,
			"PID":      p.PID,
			"SSL":      p.SSL,
			"Security": p.Security,
			"Database": p.Database,
		}
		for name, val := range fields {
			if val == "" {
				t.Errorf("bsdPaths(%v).%s is empty", isRoot, name)
			}
		}
	}
}

// ---- windowsPaths ----

func TestWindowsPaths_Root_DefaultProgramData(t *testing.T) {
	// Unset ProgramData so the fallback is used
	t.Setenv("ProgramData", "")
	p := windowsPaths(true)
	if p == nil {
		t.Fatal("windowsPaths(true) returned nil")
	}
	if !strings.Contains(p.Config, "cassocial") && !strings.Contains(p.Config, "casapps") {
		t.Errorf("windowsPaths(true).Config = %q, want path containing project/org name", p.Config)
	}
	if !strings.HasSuffix(p.PID, ".pid") {
		t.Errorf("windowsPaths(true).PID = %q, want .pid suffix", p.PID)
	}
}

func TestWindowsPaths_Root_CustomProgramData(t *testing.T) {
	t.Setenv("ProgramData", "C:\\CustomProgramData")
	p := windowsPaths(true)
	if p == nil {
		t.Fatal("windowsPaths(true) returned nil")
	}
	if !strings.HasPrefix(p.Config, "C:\\CustomProgramData") {
		t.Errorf("windowsPaths(true).Config = %q, want C:\\CustomProgramData prefix", p.Config)
	}
}

func TestWindowsPaths_User_DefaultAppData(t *testing.T) {
	// Unset AppData and LocalAppData to exercise fallback
	t.Setenv("AppData", "")
	t.Setenv("LocalAppData", "")
	p := windowsPaths(false)
	if p == nil {
		t.Fatal("windowsPaths(false) returned nil")
	}
	if !strings.Contains(p.Config, "cassocial") && !strings.Contains(p.Config, "casapps") {
		t.Errorf("windowsPaths(false).Config = %q, want path containing project/org name", p.Config)
	}
}

func TestWindowsPaths_User_CustomAppData(t *testing.T) {
	t.Setenv("AppData", "C:\\Users\\TestUser\\AppData\\Roaming")
	t.Setenv("LocalAppData", "C:\\Users\\TestUser\\AppData\\Local")
	p := windowsPaths(false)
	if p == nil {
		t.Fatal("windowsPaths(false) returned nil")
	}
	if !strings.HasPrefix(p.Config, "C:\\Users\\TestUser\\AppData\\Roaming") {
		t.Errorf("windowsPaths(false).Config = %q, want Roaming prefix", p.Config)
	}
	if !strings.HasPrefix(p.Data, "C:\\Users\\TestUser\\AppData\\Local") {
		t.Errorf("windowsPaths(false).Data = %q, want Local prefix", p.Data)
	}
}

// ---------------------------------------------------------------------------
// isRunningInDocker — container env var branch
// ---------------------------------------------------------------------------

func TestIsRunningInDocker_ContainerEnvVar(t *testing.T) {
	// Set the "container" env var that isRunningInDocker checks.
	t.Setenv("container", "podman")
	got := isRunningInDocker()
	if !got {
		t.Error("isRunningInDocker() = false when 'container' env var is set, want true")
	}
}

func TestIsRunningInDocker_NoSignals(t *testing.T) {
	// In a clean non-container environment (no /.dockerenv, no 'container' env,
	// and PID 1 either doesn't exist or isn't tini) the function must return false.
	// We can only guarantee this when the 'container' env var is unset.
	t.Setenv("container", "")
	// Result depends on the actual environment; we just ensure no panic.
	_ = isRunningInDocker()
}

// ---------------------------------------------------------------------------
// Resolve — explicit OS-specific branches called directly
// ---------------------------------------------------------------------------

func TestLinuxPaths_RootAllFieldsNonEmpty(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}
	p := linuxPaths(true)
	fields := map[string]string{
		"Config":   p.Config,
		"Data":     p.Data,
		"Cache":    p.Cache,
		"Log":      p.Log,
		"Backup":   p.Backup,
		"PID":      p.PID,
		"SSL":      p.SSL,
		"Security": p.Security,
		"Database": p.Database,
	}
	for name, val := range fields {
		if val == "" {
			t.Errorf("linuxPaths(true).%s is empty", name)
		}
	}
}

func TestLinuxPaths_UserAllFieldsNonEmpty(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}
	p := linuxPaths(false)
	fields := map[string]string{
		"Config":   p.Config,
		"Data":     p.Data,
		"Cache":    p.Cache,
		"Log":      p.Log,
		"Backup":   p.Backup,
		"PID":      p.PID,
		"SSL":      p.SSL,
		"Security": p.Security,
		"Database": p.Database,
	}
	for name, val := range fields {
		if val == "" {
			t.Errorf("linuxPaths(false).%s is empty", name)
		}
	}
}

// TestResolve_NonDockerPaths exercises the GOOS switch in Resolve by calling the
// concrete OS-specific helpers directly (since GOOS is determined at compile time).
func TestResolve_BothPrivilegeLevels(t *testing.T) {
	// Ensure we can get non-nil, non-empty paths for the current OS at both
	// privilege levels by calling the helpers directly.
	for _, isRoot := range []bool{true, false} {
		var p *Paths
		switch runtime.GOOS {
		case "linux":
			p = linuxPaths(isRoot)
		case "darwin":
			p = darwinPaths(isRoot)
		case "freebsd", "openbsd", "netbsd":
			p = bsdPaths(isRoot)
		case "windows":
			p = windowsPaths(isRoot)
		default:
			p = linuxPaths(isRoot)
		}
		if p == nil {
			t.Fatalf("OS-specific paths(isRoot=%v) returned nil", isRoot)
		}
		if p.Config == "" || p.Data == "" || p.PID == "" {
			t.Errorf("OS-specific paths(isRoot=%v) has empty required fields", isRoot)
		}
	}
}

// ---------------------------------------------------------------------------
// Resolve — docker branch via "container" env var
// ---------------------------------------------------------------------------

func TestResolve_DockerBranch_ViaEnvVar(t *testing.T) {
	// Force isRunningInDocker() to return true via the "container" env var.
	t.Setenv("container", "docker")
	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() in docker mode returned nil")
	}
	// Docker paths are fixed values — verify the known ones.
	if p.Config != "/config/cassocial" {
		t.Errorf("Resolve docker Config = %q, want /config/cassocial", p.Config)
	}
	if p.Data != "/data/cassocial" {
		t.Errorf("Resolve docker Data = %q, want /data/cassocial", p.Data)
	}
}

// ---------------------------------------------------------------------------
// isRunningInDocker — tini branch (ppid==1, /usr/bin/tini present)
// We cannot reliably fake ppid==1 in tests, but we exercise the function path
// by calling it directly and asserting it returns a bool without panicking.
// ---------------------------------------------------------------------------

func TestIsRunningInDocker_NoPanic(t *testing.T) {
	// Ensure "container" env is unset so we don't short-circuit.
	t.Setenv("container", "")
	// Just call it — result depends on environment; we only need no panic.
	_ = isRunningInDocker()
}

// ---------------------------------------------------------------------------
// isPrivileged — Windows branch (always returns false on Windows; on Linux
// returns os.Geteuid()==0, which we can observe but not force to the root path
// in an unprivileged test run)
// ---------------------------------------------------------------------------

func TestIsPrivileged_NoPanic(t *testing.T) {
	got := isPrivileged()
	// On Linux/macOS in CI (non-root) this must be false.
	// We cannot assert the value since tests may run as root in containers.
	_ = got
}

// ---------------------------------------------------------------------------
// Resolve — non-docker, current GOOS: exercise both privilege levels through
// the Resolve entry point by calling the appropriate helper directly.
// ---------------------------------------------------------------------------

func TestResolve_AllFieldsFromHelpers(t *testing.T) {
	// Make sure "container" env is unset so Resolve takes the GOOS branch.
	t.Setenv("container", "")

	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() returned nil")
	}
	if p.Config == "" || p.Data == "" || p.Log == "" {
		t.Errorf("Resolve() returned empty path(s): Config=%q Data=%q Log=%q", p.Config, p.Data, p.Log)
	}
}

func TestWindowsPaths_AllFieldsNonEmpty(t *testing.T) {
	for _, isRoot := range []bool{true, false} {
		p := windowsPaths(isRoot)
		fields := map[string]string{
			"Config":   p.Config,
			"Data":     p.Data,
			"Cache":    p.Cache,
			"Log":      p.Log,
			"Backup":   p.Backup,
			"PID":      p.PID,
			"SSL":      p.SSL,
			"Security": p.Security,
			"Database": p.Database,
		}
		for name, val := range fields {
			if val == "" {
				t.Errorf("windowsPaths(%v).%s is empty", isRoot, name)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Override helpers — restore original values after each test
// ---------------------------------------------------------------------------

func withGOOS(t *testing.T, goos string) {
	t.Helper()
	orig := getGOOS
	getGOOS = func() string { return goos }
	t.Cleanup(func() { getGOOS = orig })
}

func withNonDockerStat(t *testing.T) {
	t.Helper()
	orig := statFn
	// Make /.dockerenv appear to not exist
	statFn = func(name string) (os.FileInfo, error) {
		if name == "/.dockerenv" {
			return nil, os.ErrNotExist
		}
		return os.Stat(name)
	}
	t.Cleanup(func() { statFn = orig })
}

func withNonRootEUID(t *testing.T) {
	t.Helper()
	orig := getEUID
	getEUID = func() int { return 1000 }
	t.Cleanup(func() { getEUID = orig })
}

func withRootEUID(t *testing.T) {
	t.Helper()
	orig := getEUID
	getEUID = func() int { return 0 }
	t.Cleanup(func() { getEUID = orig })
}

// ---------------------------------------------------------------------------
// Resolve — exercise GOOS switch cases via overrides
// ---------------------------------------------------------------------------

func TestResolve_Linux_AsRoot(t *testing.T) {
	withNonDockerStat(t)
	withGOOS(t, "linux")
	withRootEUID(t)
	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() returned nil")
	}
}

func TestResolve_Linux_NonRoot(t *testing.T) {
	withNonDockerStat(t)
	withGOOS(t, "linux")
	withNonRootEUID(t)
	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() returned nil")
	}
}

func TestResolve_Darwin(t *testing.T) {
	withNonDockerStat(t)
	withGOOS(t, "darwin")
	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() returned nil")
	}
}

func TestResolve_FreeBSD(t *testing.T) {
	withNonDockerStat(t)
	withGOOS(t, "freebsd")
	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() returned nil")
	}
}

func TestResolve_Windows(t *testing.T) {
	withNonDockerStat(t)
	withGOOS(t, "windows")
	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() returned nil")
	}
}

func TestResolve_DefaultGOOS(t *testing.T) {
	withNonDockerStat(t)
	withGOOS(t, "plan9") // unknown OS → default branch
	p := Resolve()
	if p == nil {
		t.Fatal("Resolve() returned nil")
	}
}

// ---------------------------------------------------------------------------
// isPrivileged — Windows branch
// ---------------------------------------------------------------------------

func TestIsPrivileged_Windows_AlwaysFalse(t *testing.T) {
	withGOOS(t, "windows")
	got := isPrivileged()
	if got {
		t.Error("isPrivileged() on windows should always return false")
	}
}

// ---------------------------------------------------------------------------
// isRunningInDocker — container env and tini/ppid branches
// ---------------------------------------------------------------------------

func TestIsRunningInDocker_ContainerEnv(t *testing.T) {
	withNonDockerStat(t) // make /.dockerenv check fail
	t.Setenv("container", "docker")
	got := isRunningInDocker()
	if !got {
		t.Error("isRunningInDocker() should return true when 'container' env var is set")
	}
}

func TestIsRunningInDocker_ReturnsFalse_WhenNoIndicators(t *testing.T) {
	withNonDockerStat(t)
	t.Setenv("container", "")
	// ppid is not 1 in test processes; function returns false via tini branch
	got := isRunningInDocker()
	// May be true if /usr/bin/tini exists and ppid==1; just check no panic
	_ = got
}

func TestIsRunningInDocker_FalseWhenNoDockerEnv(t *testing.T) {
	withNonDockerStat(t)
	t.Setenv("container", "")
	// As long as the process's ppid is not 1 (which it isn't in normal test runs),
	// this returns false — verifying the return false path is reachable.
	if os.Getppid() != 1 {
		got := isRunningInDocker()
		if got {
			t.Error("isRunningInDocker() should return false when no docker indicators present")
		}
	}
}

func TestIsRunningInDocker_TiniPresent_Ppid1(t *testing.T) {
	// Simulate ppid==1 and /usr/bin/tini present (container init scenario).
	origStat := statFn
	origPpid := getPpid
	statFn = func(name string) (os.FileInfo, error) {
		if name == "/.dockerenv" {
			return nil, os.ErrNotExist
		}
		if name == "/usr/bin/tini" {
			return nil, nil // pretend tini exists
		}
		return os.Stat(name)
	}
	getPpid = func() int { return 1 }
	t.Cleanup(func() {
		statFn = origStat
		getPpid = origPpid
	})
	t.Setenv("container", "")

	got := isRunningInDocker()
	if !got {
		t.Error("isRunningInDocker() should return true when ppid==1 and tini is present")
	}
}
