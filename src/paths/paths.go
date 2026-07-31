package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	projectName = "cassocial"
	projectOrg  = "casapps"
)

// statFn is used to check for /.dockerenv. Overridable in tests.
var statFn = os.Stat

// getGOOS returns the current OS. Overridable in tests to simulate other platforms.
var getGOOS = func() string { return runtime.GOOS }

// getEUID returns the effective UID. Overridable in tests to simulate non-root.
var getEUID = os.Geteuid

// getPpid returns the parent process ID. Overridable in tests.
var getPpid = os.Getppid

// Paths holds all application paths
type Paths struct {
	Config   string
	Data     string
	Cache    string
	Log      string
	Backup   string
	PID      string
	SSL      string
	Security string
	Database string
}

// Resolve returns OS-appropriate paths based on privilege level
func Resolve() *Paths {
	isRoot := isPrivileged()
	isDocker := isRunningInDocker()

	if isDocker {
		return dockerPaths()
	}

	switch getGOOS() {
	case "linux":
		return linuxPaths(isRoot)
	case "darwin":
		return darwinPaths(isRoot)
	case "freebsd", "openbsd", "netbsd":
		return bsdPaths(isRoot)
	case "windows":
		return windowsPaths(isRoot)
	default:
		return linuxPaths(isRoot)
	}
}

// isPrivileged checks if running with elevated privileges
func isPrivileged() bool {
	if getGOOS() == "windows" {
		// On Windows, check if running as Administrator
		// For now, assume non-privileged
		return false
	}
	return getEUID() == 0
}

// isRunningInDocker checks if running inside a Docker container
func isRunningInDocker() bool {
	// Check for .dockerenv file
	if _, err := statFn("/.dockerenv"); err == nil {
		return true
	}

	// Check for container env var
	if os.Getenv("container") != "" {
		return true
	}

	// Check if running under tini or other container init
	if getPpid() == 1 {
		if _, err := statFn("/usr/bin/tini"); err == nil {
			return true
		}
	}

	return false
}

// dockerPaths returns container paths
func dockerPaths() *Paths {
	return &Paths{
		Config:   "/config/" + projectName,
		Data:     "/data/" + projectName,
		Cache:    "/data/" + projectName + "/cache",
		Log:      "/data/log/" + projectName,
		Backup:   "/data/backups/" + projectName,
		PID:      "/data/" + projectName + "/" + projectName + ".pid",
		SSL:      "/config/" + projectName + "/ssl",
		Security: "/data/" + projectName + "/security",
		Database: "/data/db/sqlite",
	}
}

// linuxPaths returns Linux-specific paths
func linuxPaths(isRoot bool) *Paths {
	if isRoot {
		return &Paths{
			Config:   "/etc/" + projectOrg + "/" + projectName,
			Data:     "/var/lib/" + projectOrg + "/" + projectName,
			Cache:    "/var/cache/" + projectOrg + "/" + projectName,
			Log:      "/var/log/" + projectOrg + "/" + projectName,
			Backup:   "/mnt/Backups/" + projectOrg + "/" + projectName,
			PID:      "/var/run/" + projectOrg + "/" + projectName + ".pid",
			SSL:      "/etc/" + projectOrg + "/" + projectName + "/ssl",
			Security: "/var/lib/" + projectOrg + "/" + projectName + "/security",
			Database: "/var/lib/" + projectOrg + "/" + projectName + "/db",
		}
	}

	homeDir, _ := os.UserHomeDir()
	return &Paths{
		Config:   filepath.Join(homeDir, ".config", projectOrg, projectName),
		Data:     filepath.Join(homeDir, ".local", "share", projectOrg, projectName),
		Cache:    filepath.Join(homeDir, ".cache", projectOrg, projectName),
		Log:      filepath.Join(homeDir, ".local", "log", projectOrg, projectName),
		Backup:   filepath.Join(homeDir, ".local", "share", "Backups", projectOrg, projectName),
		PID:      filepath.Join(homeDir, ".local", "share", projectOrg, projectName, projectName+".pid"),
		SSL:      filepath.Join(homeDir, ".config", projectOrg, projectName, "ssl"),
		Security: filepath.Join(homeDir, ".local", "share", projectOrg, projectName, "security"),
		Database: filepath.Join(homeDir, ".local", "share", projectOrg, projectName, "db"),
	}
}

// darwinPaths returns macOS-specific paths
func darwinPaths(isRoot bool) *Paths {
	if isRoot {
		return &Paths{
			Config:   "/Library/Application Support/" + projectOrg + "/" + projectName,
			Data:     "/Library/Application Support/" + projectOrg + "/" + projectName + "/data",
			Cache:    "/Library/Caches/" + projectOrg + "/" + projectName,
			Log:      "/Library/Logs/" + projectOrg + "/" + projectName,
			Backup:   "/Library/Backups/" + projectOrg + "/" + projectName,
			PID:      "/var/run/" + projectOrg + "/" + projectName + ".pid",
			SSL:      "/Library/Application Support/" + projectOrg + "/" + projectName + "/ssl",
			Security: "/Library/Application Support/" + projectOrg + "/" + projectName + "/data/security",
			Database: "/Library/Application Support/" + projectOrg + "/" + projectName + "/db",
		}
	}

	homeDir, _ := os.UserHomeDir()
	return &Paths{
		Config:   filepath.Join(homeDir, "Library", "Application Support", projectOrg, projectName),
		Data:     filepath.Join(homeDir, "Library", "Application Support", projectOrg, projectName),
		Cache:    filepath.Join(homeDir, "Library", "Caches", projectOrg, projectName),
		Log:      filepath.Join(homeDir, "Library", "Logs", projectOrg, projectName),
		Backup:   filepath.Join(homeDir, "Library", "Backups", projectOrg, projectName),
		PID:      filepath.Join(homeDir, "Library", "Application Support", projectOrg, projectName, projectName+".pid"),
		SSL:      filepath.Join(homeDir, "Library", "Application Support", projectOrg, projectName, "ssl"),
		Security: filepath.Join(homeDir, "Library", "Application Support", projectOrg, projectName, "data", "security"),
		Database: filepath.Join(homeDir, "Library", "Application Support", projectOrg, projectName, "db"),
	}
}

// bsdPaths returns BSD-specific paths (FreeBSD, OpenBSD, NetBSD)
func bsdPaths(isRoot bool) *Paths {
	if isRoot {
		return &Paths{
			Config:   "/usr/local/etc/" + projectOrg + "/" + projectName,
			Data:     "/var/db/" + projectOrg + "/" + projectName,
			Cache:    "/var/cache/" + projectOrg + "/" + projectName,
			Log:      "/var/log/" + projectOrg + "/" + projectName,
			Backup:   "/var/backups/" + projectOrg + "/" + projectName,
			PID:      "/var/run/" + projectOrg + "/" + projectName + ".pid",
			SSL:      "/usr/local/etc/" + projectOrg + "/" + projectName + "/ssl",
			Security: "/var/db/" + projectOrg + "/" + projectName + "/security",
			Database: "/var/db/" + projectOrg + "/" + projectName + "/db",
		}
	}

	homeDir, _ := os.UserHomeDir()
	return &Paths{
		Config:   filepath.Join(homeDir, ".config", projectOrg, projectName),
		Data:     filepath.Join(homeDir, ".local", "share", projectOrg, projectName),
		Cache:    filepath.Join(homeDir, ".cache", projectOrg, projectName),
		Log:      filepath.Join(homeDir, ".local", "log", projectOrg, projectName),
		Backup:   filepath.Join(homeDir, ".local", "share", "Backups", projectOrg, projectName),
		PID:      filepath.Join(homeDir, ".local", "share", projectOrg, projectName, projectName+".pid"),
		SSL:      filepath.Join(homeDir, ".config", projectOrg, projectName, "ssl"),
		Security: filepath.Join(homeDir, ".local", "share", projectOrg, projectName, "security"),
		Database: filepath.Join(homeDir, ".local", "share", projectOrg, projectName, "db"),
	}
}

// windowsPaths returns Windows-specific paths
func windowsPaths(isRoot bool) *Paths {
	if isRoot {
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = "C:\\ProgramData"
		}

		return &Paths{
			Config:   filepath.Join(programData, projectOrg, projectName),
			Data:     filepath.Join(programData, projectOrg, projectName, "data"),
			Cache:    filepath.Join(programData, projectOrg, projectName, "cache"),
			Log:      filepath.Join(programData, projectOrg, projectName, "logs"),
			Backup:   filepath.Join(programData, "Backups", projectOrg, projectName),
			PID:      filepath.Join(programData, projectOrg, projectName, projectName+".pid"),
			SSL:      filepath.Join(programData, projectOrg, projectName, "ssl"),
			Security: filepath.Join(programData, projectOrg, projectName, "data", "security"),
			Database: filepath.Join(programData, projectOrg, projectName, "db"),
		}
	}

	appData := os.Getenv("AppData")
	localAppData := os.Getenv("LocalAppData")
	if appData == "" {
		homeDir, _ := os.UserHomeDir()
		appData = filepath.Join(homeDir, "AppData", "Roaming")
	}
	if localAppData == "" {
		homeDir, _ := os.UserHomeDir()
		localAppData = filepath.Join(homeDir, "AppData", "Local")
	}

	return &Paths{
		Config:   filepath.Join(appData, projectOrg, projectName),
		Data:     filepath.Join(localAppData, projectOrg, projectName),
		Cache:    filepath.Join(localAppData, projectOrg, projectName, "cache"),
		Log:      filepath.Join(localAppData, projectOrg, projectName, "logs"),
		Backup:   filepath.Join(localAppData, "Backups", projectOrg, projectName),
		PID:      filepath.Join(localAppData, projectOrg, projectName, projectName+".pid"),
		SSL:      filepath.Join(appData, projectOrg, projectName, "ssl"),
		Security: filepath.Join(localAppData, projectOrg, projectName, "security"),
		Database: filepath.Join(localAppData, projectOrg, projectName, "db"),
	}
}

// GetConfigDir returns the configuration directory
func GetConfigDir() string {
	return Resolve().Config
}

// GetDataDir returns the data directory
func GetDataDir() string {
	return Resolve().Data
}

// GetCacheDir returns the cache directory
func GetCacheDir() string {
	return Resolve().Cache
}

// GetLogDir returns the log directory
func GetLogDir() string {
	return Resolve().Log
}

// GetBackupDir returns the backup directory
func GetBackupDir() string {
	return Resolve().Backup
}

// GetPIDFile returns the PID file path
func GetPIDFile() string {
	return Resolve().PID
}

// GetSSLDir returns the SSL directory
func GetSSLDir() string {
	return Resolve().SSL
}

// GetSecurityDir returns the security directory
func GetSecurityDir() string {
	return Resolve().Security
}

// GetDatabaseDir returns the database directory
func GetDatabaseDir() string {
	return Resolve().Database
}
