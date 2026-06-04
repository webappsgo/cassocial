package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/service"
	"github.com/casapps/cassocial/src/server/store"
)

// systemdUnitTemplate is the systemd unit file for cassocial
const systemdUnitTemplate = `[Unit]
Description=Cassocial - Self-hosted link aggregator and social profile
Documentation=https://github.com/casapps/cassocial
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
User=cassocial
Group=cassocial
ExecStart={binary_path} --mode production
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=cassocial

[Install]
WantedBy=multi-user.target
`

// systemdUnitPath is the path for the systemd unit file
const systemdUnitPath = "/etc/systemd/system/cassocial.service"

// handleStatus prints the current server status by reading the PID file and probing health.
func handleStatus(cfg *config.Config, pidFile string) {
	if pidFile == "" && cfg == nil {
		fmt.Println("Status: no PID file configured")
		return
	}

	running, pid, err := server.CheckPIDFile(pidFile)
	if err != nil {
		fmt.Printf("Status: error reading PID file: %v\n", err)
		return
	}

	if !running {
		fmt.Println("Status: stopped")
		return
	}

	fmt.Printf("Status: running (PID %d)\n", pid)

	if cfg == nil {
		return
	}

	fmt.Printf("Listen: %s:%d\n", cfg.Server.Address, cfg.Server.Port)

	// Try health endpoint
	url := fmt.Sprintf("http://%s:%d/server/healthz", cfg.Server.Address, cfg.Server.Port)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Health: unreachable (%v)\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Health: %s\n", resp.Status)
}

// handleDaemon re-executes the current binary in the background without the --daemon flag.
// Uses the daemon-env guard CASSOCIAL_DAEMONIZED=1 to avoid infinite re-exec loops.
func handleDaemon(args []string) int {
	if os.Getenv("CASSOCIAL_DAEMONIZED") == "1" {
		// Already in daemon mode — just run normally
		return -1
	}

	// Build arg list without --daemon
	filtered := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--daemon" {
			continue
		}
		filtered = append(filtered, a)
	}

	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot determine executable path:", err)
		return 1
	}

	cmd := exec.Command(self, filtered...)
	cmd.Env = append(os.Environ(), "CASSOCIAL_DAEMONIZED=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "error: failed to start daemon:", err)
		return 1
	}

	fmt.Printf("Cassocial started as daemon (PID %d)\n", cmd.Process.Pid)
	return 0
}

// handleService manages the systemd service unit.
func handleService(subcmd string) int {
	switch subcmd {
	case "--install":
		return installService()
	case "--uninstall":
		return uninstallService()
	case "--disable":
		return runSystemctl("disable", "--now", "cassocial")
	case "start":
		return runSystemctl("start", "cassocial")
	case "stop":
		return runSystemctl("stop", "cassocial")
	case "restart":
		return runSystemctl("restart", "cassocial")
	case "reload":
		return runSystemctl("reload-or-restart", "cassocial")
	case "--help":
		fmt.Println("Service commands: start, stop, restart, reload, --install, --uninstall, --disable")
		return 0
	default:
		fmt.Fprintf(os.Stderr, "error: unknown service command: %s\n", subcmd)
		fmt.Fprintln(os.Stderr, "Available: start, stop, restart, reload, --install, --uninstall, --disable, --help")
		return 1
	}
}

// installService writes the systemd unit file and enables the service.
func installService() int {
	if runtime.GOOS != "linux" {
		fmt.Fprintln(os.Stderr, "error: --service --install is only supported on Linux (systemd)")
		return 1
	}

	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot determine executable path:", err)
		return 1
	}

	// Resolve symlinks for the real binary path
	realSelf, err := filepath.EvalSymlinks(self)
	if err != nil {
		realSelf = self
	}

	unitContent := strings.ReplaceAll(systemdUnitTemplate, "{binary_path}", realSelf)

	// Write unit file (requires root)
	if err := os.WriteFile(systemdUnitPath, []byte(unitContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write unit file %s: %v\n", systemdUnitPath, err)
		fmt.Fprintln(os.Stderr, "hint: run with sudo or as root")
		return 1
	}

	fmt.Printf("Wrote systemd unit: %s\n", systemdUnitPath)

	if rc := runSystemctl("daemon-reload"); rc != 0 {
		return rc
	}
	if rc := runSystemctl("enable", "--now", "cassocial"); rc != 0 {
		return rc
	}

	fmt.Println("Service installed and started successfully")
	return 0
}

// uninstallService stops, disables, and removes the systemd unit file.
func uninstallService() int {
	if runtime.GOOS != "linux" {
		fmt.Fprintln(os.Stderr, "error: --service --uninstall is only supported on Linux (systemd)")
		return 1
	}

	runSystemctl("stop", "cassocial")
	runSystemctl("disable", "cassocial")

	if err := os.Remove(systemdUnitPath); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: failed to remove unit file: %v\n", err)
		return 1
	}

	runSystemctl("daemon-reload")
	fmt.Println("Service uninstalled successfully")
	return 0
}

// runSystemctl runs systemctl with the given arguments.
func runSystemctl(args ...string) int {
	cmd := exec.Command("systemctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: systemctl %s failed: %v\n", strings.Join(args, " "), err)
		return 1
	}
	return 0
}

// handleMaintenance dispatches maintenance sub-commands.
// Connects to the DB when required.
func handleMaintenance(subcmd string, cfg *config.Config, remainingArgs []string) int {
	switch subcmd {
	case "--help":
		fmt.Println("Maintenance commands: backup, restore, update, mode, setup, --help")
		fmt.Println("  backup            Create a backup now")
		fmt.Println("  restore {file}    Restore from backup file")
		fmt.Println("  mode enable|disable  Toggle maintenance mode")
		fmt.Println("  setup             Re-run first-run setup wizard")
		return 0

	case "backup":
		return runMaintenanceBackup(cfg)

	case "restore":
		if len(remainingArgs) == 0 {
			fmt.Fprintln(os.Stderr, "error: --maintenance restore requires a backup filename argument")
			return 1
		}
		return runMaintenanceRestore(cfg, remainingArgs[0])

	case "mode":
		if len(remainingArgs) == 0 {
			fmt.Fprintln(os.Stderr, "error: --maintenance mode requires enable or disable")
			return 1
		}
		return runMaintenanceMode(cfg, remainingArgs[0])

	case "setup":
		fmt.Println("To re-run setup, visit /setup in your browser after the server is running.")
		return 0

	case "update":
		fmt.Println("Use --update to manage software updates.")
		return 0

	default:
		fmt.Fprintf(os.Stderr, "error: unknown maintenance command: %s\n", subcmd)
		fmt.Fprintln(os.Stderr, "Available: backup, restore, update, mode, setup, --help")
		return 1
	}
}

// runMaintenanceBackup creates a backup of the database and uploaded files.
func runMaintenanceBackup(cfg *config.Config) int {
	db, err := store.Connect(cfg.Database.Driver, cfg.Database.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot connect to database: %v\n", err)
		return 1
	}
	defer db.Close()

	svc := service.NewBackupService(cfg, db)
	backup, err := svc.CreateBackup("manual")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: backup failed: %v\n", err)
		return 1
	}

	fmt.Printf("Backup created: %s (%d bytes)\n", backup.Filename, backup.Size)
	return 0
}

// runMaintenanceRestore restores from a named backup file.
func runMaintenanceRestore(cfg *config.Config, filename string) int {
	db, err := store.Connect(cfg.Database.Driver, cfg.Database.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot connect to database: %v\n", err)
		return 1
	}
	defer db.Close()

	svc := service.NewBackupService(cfg, db)
	if err := svc.RestoreBackup(filename); err != nil {
		fmt.Fprintf(os.Stderr, "error: restore failed: %v\n", err)
		return 1
	}

	fmt.Printf("Restored from backup: %s\n", filename)
	return 0
}

// runMaintenanceMode enables or disables maintenance mode.
func runMaintenanceMode(cfg *config.Config, action string) int {
	db, err := store.Connect(cfg.Database.Driver, cfg.Database.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot connect to database: %v\n", err)
		return 1
	}
	defer db.Close()

	mm := server.NewMaintenanceMode(db)

	switch action {
	case "enable", "on", "true":
		if err := mm.Enable(""); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to enable maintenance mode: %v\n", err)
			return 1
		}
		fmt.Println("Maintenance mode enabled")
	case "disable", "off", "false":
		if err := mm.Disable(); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to disable maintenance mode: %v\n", err)
			return 1
		}
		fmt.Println("Maintenance mode disabled")
	default:
		fmt.Fprintf(os.Stderr, "error: unknown mode action: %s (use enable or disable)\n", action)
		return 1
	}

	return 0
}

// githubRelease represents the GitHub API release response.
type githubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	HTMLURL    string `json:"html_url"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// handleUpdate dispatches update sub-commands.
func handleUpdate(subcmd string, remainingArgs []string) int {
	switch subcmd {
	case "check":
		return runUpdateCheck()
	case "yes":
		return runUpdateInstall()
	case "branch":
		if len(remainingArgs) == 0 {
			fmt.Fprintln(os.Stderr, "error: --update branch requires a channel: stable, beta, or daily")
			return 1
		}
		return runUpdateBranch(remainingArgs[0])
	case "":
		return runUpdateCheck()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown update command: %s\n", subcmd)
		fmt.Fprintln(os.Stderr, "Available: check, yes, branch {stable|beta|daily}")
		return 1
	}
}

// runUpdateCheck checks GitHub for the latest release.
func runUpdateCheck() int {
	release, err := fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not check for updates: %v\n", err)
		return 1
	}

	if release.TagName == "" {
		fmt.Println("No releases found")
		return 0
	}

	current := Version
	latest := strings.TrimPrefix(release.TagName, "v")

	if current == latest || current == "dev" {
		fmt.Printf("Current version: %s\n", current)
		fmt.Printf("Latest release:  %s\n", latest)
		if current == latest {
			fmt.Println("You are running the latest version")
		}
	} else {
		fmt.Printf("Update available: %s → %s\n", current, latest)
		fmt.Printf("Release notes: %s\n", release.HTMLURL)
		fmt.Println("Run --update yes to install the update")
	}

	return 0
}

// runUpdateInstall downloads and replaces the current binary with the latest release.
func runUpdateInstall() int {
	release, err := fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not fetch release info: %v\n", err)
		return 1
	}

	// Build expected asset name for this platform
	assetName := fmt.Sprintf("cassocial-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		fmt.Fprintf(os.Stderr, "error: no release asset found for %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return 1
	}

	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot determine executable path:", err)
		return 1
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Download to temp file
	tmpFile := self + ".new"
	if err := downloadFile(downloadURL, tmpFile); err != nil {
		fmt.Fprintf(os.Stderr, "error: download failed: %v\n", err)
		return 1
	}

	// Make executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "error: failed to make binary executable: %v\n", err)
		return 1
	}

	// Replace binary atomically
	if err := os.Rename(tmpFile, self); err != nil {
		os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "error: failed to replace binary: %v\n", err)
		return 1
	}

	fmt.Printf("Updated to %s. Restart the service to apply.\n", release.TagName)
	return 0
}

// runUpdateBranch sets the update channel preference.
func runUpdateBranch(channel string) int {
	switch channel {
	case "stable", "beta", "daily":
		fmt.Printf("Update channel set to: %s\n", channel)
		fmt.Println("Note: this setting is informational; use --update yes to install the latest release")
		return 0
	default:
		fmt.Fprintf(os.Stderr, "error: unknown channel: %s (use stable, beta, or daily)\n", channel)
		return 1
	}
}

// fetchLatestRelease fetches the latest release info from GitHub.
func fetchLatestRelease() (*githubRelease, error) {
	const apiURL = "https://api.github.com/repos/casapps/cassocial/releases/latest"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "cassocial/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadFile downloads a URL to a local file.
func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download request returned %s", resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

