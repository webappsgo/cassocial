package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete server configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
	SSL      SSLConfig      `yaml:"ssl"`
	Email    EmailConfig    `yaml:"email"`

	// Cassocial-specific configuration
	Cassocial CassocialConfig `yaml:"cassocial"`

	// Scheduler manages all background tasks (geoip/blocklist updates, backups, etc.)
	Scheduler SchedulerConfig `yaml:"scheduler"`

	// RateLimit controls per-IP request throttling
	RateLimit RateLimitConfig `yaml:"rate_limit"`

	// Web is the frontend configuration (theme, CORS)
	Web WebConfig `yaml:"web"`

	// Internal fields (not in YAML)
	ConfigDir string `yaml:"-"`
	DataDir   string `yaml:"-"`
	LogDir    string `yaml:"-"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
	FQDN    string `yaml:"fqdn"` // Auto-detected from host, overridable via DOMAIN env var
	Mode    string `yaml:"mode"` // production or development
	Debug   bool   `yaml:"debug"`

	// AdminPath is the admin panel URL path segment (default: admin) - see PART 17
	AdminPath string `yaml:"admin_path"`
	// APIVersion is the API version prefix used in /api/{api_version}/ routes
	APIVersion string `yaml:"api_version"`

	Healthz HealthzConfig `yaml:"healthz"`

	// Branding & SEO - see PART 16 for full details
	Branding BrandingConfig `yaml:"branding"`
	SEO      SEOConfig      `yaml:"seo"`

	// System user/group the server runs as
	User  string `yaml:"user"`
	Group string `yaml:"group"`

	// PIDFile enables writing a PID file on start
	PIDFile bool `yaml:"pidfile"`

	// Daemonize detaches from the terminal on start (default: false)
	Daemonize bool `yaml:"daemonize"`

	Admin AdminPanelConfig `yaml:"admin"`
}

// HealthzConfig configures the /server/healthz endpoint family.
type HealthzConfig struct {
	Root HealthzRootConfig `yaml:"root"`
}

// HealthzRootConfig controls the optional /healthz root compatibility alias.
type HealthzRootConfig struct {
	// Enabled mounts /healthz to the SAME handler as /server/healthz (never redirect)
	Enabled bool `yaml:"enabled"`
}

// BrandingConfig holds site branding shown in the UI and page metadata.
type BrandingConfig struct {
	Title       string `yaml:"title"`
	Tagline     string `yaml:"tagline"`
	Description string `yaml:"description"`
}

// SEOConfig holds search-engine metadata.
type SEOConfig struct {
	Keywords []string `yaml:"keywords"`
}

// AdminPanelConfig holds admin-panel settings that live in server.yml.
// Username, password, and token are stored in the database (admins table),
// never in this config file.
type AdminPanelConfig struct {
	Email string `yaml:"email"`
}

// SchedulerConfig controls the built-in background task scheduler.
type SchedulerConfig struct {
	Enabled bool                           `yaml:"enabled"`
	Tasks   map[string]SchedulerTaskConfig `yaml:"tasks"`
}

// SchedulerTaskConfig configures a single scheduled task.
type SchedulerTaskConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Schedule     string `yaml:"schedule"`
	RetryOnFail  bool   `yaml:"retry_on_fail,omitempty"`
	RetryDelay   string `yaml:"retry_delay,omitempty"`
	Retention    int    `yaml:"retention,omitempty"`
	RenewBefore  string `yaml:"renew_before,omitempty"`
}

// RateLimitConfig controls per-IP request throttling.
type RateLimitConfig struct {
	Enabled bool             `yaml:"enabled"`
	Read    RateLimitRule    `yaml:"read"`
	Write   RateLimitRule    `yaml:"write"`
	Health  RateLimitRule    `yaml:"health"`
	// GlobalBurst is the absolute per-IP ceiling across all endpoint types (per minute)
	GlobalBurst int                 `yaml:"global_burst"`
	Auth        RateLimitAuthConfig `yaml:"auth"`
}

// RateLimitRule is a requests-per-window throttle.
type RateLimitRule struct {
	Requests int `yaml:"requests"`
	Window   int `yaml:"window"` // seconds
}

// RateLimitAuthConfig holds stricter limits for authentication endpoints,
// applied independently of the general limits above.
type RateLimitAuthConfig struct {
	Login          RateLimitRule `yaml:"login"`
	PasswordReset  RateLimitRule `yaml:"password_reset"`
	Registration   RateLimitRule `yaml:"registration"`
}

// WebConfig is the frontend configuration.
type WebConfig struct {
	UI   WebUIConfig `yaml:"ui"`
	CORS string      `yaml:"cors"`
}

// WebUIConfig holds frontend UI preferences.
type WebUIConfig struct {
	Theme string `yaml:"theme"` // dark, light, or auto
}

type DatabaseConfig struct {
	Driver      string `yaml:"driver"`       // sqlite, postgres, mysql
	URL         string `yaml:"url"`          // connection string (overrides host/port/name/user/password when set)
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Name        string `yaml:"name"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	SSLMode     string `yaml:"ssl_mode"`
	MaxConns    int    `yaml:"max_connections"`
	MaxIdleConns int   `yaml:"max_idle_connections"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`   // debug, info, warn, error
	Format string `yaml:"format"`  // text, json
}

type SSLConfig struct {
	Enabled     bool   `yaml:"enabled"`
	CertFile    string `yaml:"cert_file"`
	KeyFile     string `yaml:"key_file"`
	LetsEncrypt bool   `yaml:"letsencrypt"`
	Domain      string `yaml:"domain"`
}

type EmailConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	// FromName is the sender display name (default: app title)
	FromName string `yaml:"from_name"`
	From     string `yaml:"from"`
	// TLS enables/disables SMTP transport security; derived from TLSMode
	// (false only when TLSMode is "none")
	TLS bool `yaml:"tls"`
	// TLSMode is the SMTP transport security mode: auto, starttls, tls, none
	TLSMode string `yaml:"tls_mode"`
}

// CassocialConfig contains Cassocial-specific settings
type CassocialConfig struct {
	SiteName        string `yaml:"site_name"`
	SiteDescription string `yaml:"site_description"`
	AllowRegistration bool `yaml:"allow_registration"`
	MaxProfilesPerUser int `yaml:"max_profiles_per_user"`
	MaxLinksPerProfile int `yaml:"max_links_per_profile"`
}

// getEUID is the function used to obtain the effective UID. It defaults to
// os.Geteuid() and can be overridden in tests to simulate non-root environments.
var getEUID = os.Geteuid

// writeFileFn is used by Save to write the config file. Overridable in tests.
var writeFileFn = os.WriteFile

// Load loads configuration from file, environment variables, and CLI flags
func Load(configDir, dataDir, logDir string) (*Config, error) {
	cfg := &Config{}

	// Determine directories
	cfg.ConfigDir = determineConfigDir(configDir)
	cfg.DataDir = determineDataDir(dataDir)
	cfg.LogDir = determineLogDir(logDir)

	// Ensure directories exist
	if err := cfg.ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Migrate legacy server.yaml to server.yml if present
	configFile := filepath.Join(cfg.ConfigDir, "server.yml")
	legacyConfigFile := filepath.Join(cfg.ConfigDir, "server.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if _, legacyErr := os.Stat(legacyConfigFile); legacyErr == nil {
			if err := os.Rename(legacyConfigFile, configFile); err != nil {
				return nil, fmt.Errorf("failed to migrate server.yaml to server.yml: %w", err)
			}
		}
	}

	// Load from file if exists
	if _, err := os.Stat(configFile); err == nil {
		if err := loadFromFile(configFile, cfg); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	} else {
		// Create config file with defaults
		if err := cfg.setDefaults(); err != nil {
			return nil, err
		}
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	// Override with environment variables
	cfg.loadFromEnv()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// ensureDirectories creates all required directories
func (cfg *Config) ensureDirectories() error {
	dirs := []string{
		cfg.ConfigDir,
		cfg.DataDir,
		cfg.LogDir,
		filepath.Join(cfg.DataDir, "db"),
		filepath.Join(cfg.DataDir, "backup"),
		filepath.Join(cfg.ConfigDir, "ssl"),
		filepath.Join(cfg.ConfigDir, "security"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Validate validates the configuration
func (cfg *Config) Validate() error {
	// Validate server config
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", cfg.Server.Port)
	}

	if cfg.Server.Mode != "production" && cfg.Server.Mode != "development" {
		return fmt.Errorf("invalid mode: %s (must be 'production' or 'development')", cfg.Server.Mode)
	}

	// Validate database config
	validDrivers := map[string]bool{"sqlite": true, "pgx": true, "postgres": true, "mysql": true}
	if !validDrivers[cfg.Database.Driver] {
		return fmt.Errorf("invalid database driver: %s (must be sqlite, pgx, postgres, or mysql)", cfg.Database.Driver)
	}

	// Validate logging config
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", cfg.Logging.Level)
	}

	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[cfg.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be text or json)", cfg.Logging.Format)
	}

	return nil
}

func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}

func (cfg *Config) setDefaults() error {
	cfg.Server.Address = "0.0.0.0"
	cfg.Server.Port = 8080
	cfg.Server.Mode = "production"
	cfg.Server.Debug = false

	cfg.Server.FQDN = detectFQDN()
	cfg.Server.AdminPath = "admin"
	cfg.Server.APIVersion = "v1"
	cfg.Server.Healthz.Root.Enabled = false

	cfg.Server.Branding.Title = "Cassocial"
	cfg.Server.Branding.Tagline = ""
	cfg.Server.Branding.Description = ""
	cfg.Server.SEO.Keywords = []string{}

	cfg.Server.User = currentUsername()
	cfg.Server.Group = currentGroupname()
	cfg.Server.PIDFile = true
	cfg.Server.Daemonize = false
	cfg.Server.Admin.Email = "admin@" + cfg.Server.FQDN

	cfg.Database.Driver = "sqlite"
	cfg.Database.Name = filepath.Join(cfg.DataDir, "db", "cassocial.db")
	cfg.Database.MaxConns = 10
	cfg.Database.MaxIdleConns = 5

	cfg.Logging.Level = "info"
	cfg.Logging.Format = "text"

	cfg.SSL.Enabled = false
	cfg.SSL.LetsEncrypt = false

	cfg.Email.Enabled = false
	cfg.Email.Port = 587
	cfg.Email.TLS = true
	cfg.Email.TLSMode = "auto"
	cfg.Email.FromName = cfg.Server.Branding.Title

	cfg.Scheduler.Enabled = true
	cfg.Scheduler.Tasks = defaultSchedulerTasks()

	cfg.RateLimit.Enabled = true
	cfg.RateLimit.Read = RateLimitRule{Requests: 120, Window: 60}
	cfg.RateLimit.Write = RateLimitRule{Requests: 10, Window: 60}
	cfg.RateLimit.Health = RateLimitRule{Requests: 120, Window: 60}
	cfg.RateLimit.GlobalBurst = 240
	cfg.RateLimit.Auth = RateLimitAuthConfig{
		Login:         RateLimitRule{Requests: 5, Window: 900},
		PasswordReset: RateLimitRule{Requests: 3, Window: 3600},
		Registration:  RateLimitRule{Requests: 5, Window: 3600},
	}

	cfg.Web.UI.Theme = "dark"
	cfg.Web.CORS = "*"

	cfg.Cassocial.SiteName = "Cassocial"
	cfg.Cassocial.SiteDescription = "Self-hosted link aggregator and social profile"
	cfg.Cassocial.AllowRegistration = true
	cfg.Cassocial.MaxProfilesPerUser = 5
	cfg.Cassocial.MaxLinksPerProfile = 100

	return nil
}

// defaultSchedulerTasks returns the built-in scheduled tasks with sane
// defaults, all enabled, per AI.md PART 5.
func defaultSchedulerTasks() map[string]SchedulerTaskConfig {
	return map[string]SchedulerTaskConfig{
		"geoip_update":     {Enabled: true, Schedule: "0 3 * * 0", RetryOnFail: true, RetryDelay: "1h"},
		"blocklist_update": {Enabled: true, Schedule: "0 4 * * *", RetryOnFail: true, RetryDelay: "1h"},
		"cve_update":       {Enabled: true, Schedule: "0 5 * * *", RetryOnFail: true, RetryDelay: "1h"},
		"log_rotation":     {Enabled: true, Schedule: "0 0 * * *"},
		"session_cleanup":  {Enabled: true, Schedule: "@hourly"},
		"backup":           {Enabled: true, Schedule: "0 2 * * *", Retention: 4},
		"ssl_renewal":      {Enabled: true, Schedule: "0 3 * * *", RenewBefore: "7d"},
		"health_check":     {Enabled: true, Schedule: "*/5 * * * *"},
		"tor_health":       {Enabled: true, Schedule: "*/10 * * * *"},
	}
}

// detectFQDN resolves the server FQDN: DOMAIN env var > os.Hostname() > localhost.
func detectFQDN() string {
	if domain := os.Getenv("DOMAIN"); domain != "" {
		return domain
	}
	if host, err := os.Hostname(); err == nil && host != "" {
		return host
	}
	return "localhost"
}

// currentUsername returns the current OS username, or "" if it cannot be determined.
func currentUsername() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return ""
}

// currentGroupname returns the current OS primary group name, or "" if it cannot be determined.
func currentGroupname() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	g, err := user.LookupGroupId(u.Gid)
	if err != nil {
		return ""
	}
	return g.Name
}

func (cfg *Config) loadFromEnv() {
	// Runtime variables (always checked) — see AI.md PART 5 Environment Variables
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if listen := os.Getenv("LISTEN"); listen != "" {
		cfg.Server.Address = listen
	}
	// DOMAIN has the highest priority for hostname/FQDN resolution.
	if domain := os.Getenv("DOMAIN"); domain != "" {
		cfg.Server.FQDN = domain
	}
	if modeVal := os.Getenv("MODE"); modeVal != "" {
		if strings.EqualFold(modeVal, "debug") {
			cfg.Server.Mode = "development"
		} else {
			cfg.Server.Mode = modeVal
		}
	}
	// Debug: explicit DEBUG env var wins; otherwise MODE=debug expands to debug on.
	if debug := os.Getenv("DEBUG"); debug != "" {
		if val, err := ParseBool(debug, cfg.Server.Debug); err == nil {
			cfg.Server.Debug = val
		}
	} else if strings.EqualFold(os.Getenv("MODE"), "debug") {
		cfg.Server.Debug = true
	}

	// Database settings
	if driver := os.Getenv("DATABASE_DRIVER"); driver != "" {
		cfg.Database.Driver = driver
	}
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		cfg.Database.URL = dsn
	}

	// Email/SMTP settings
	if host := os.Getenv("SMTP_HOST"); host != "" {
		cfg.Email.Host = host
	}
	if port := os.Getenv("SMTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Email.Port = p
		}
	}
	if user := os.Getenv("SMTP_USERNAME"); user != "" {
		cfg.Email.Username = user
	}
	if password := os.Getenv("SMTP_PASSWORD"); password != "" {
		cfg.Email.Password = password
	}
	if fromName := os.Getenv("SMTP_FROM_NAME"); fromName != "" {
		cfg.Email.FromName = fromName
	}
	if from := os.Getenv("SMTP_FROM_EMAIL"); from != "" {
		cfg.Email.From = from
	}
	if tlsMode := os.Getenv("SMTP_TLS"); tlsMode != "" {
		cfg.Email.TLSMode = strings.ToLower(tlsMode)
		// Keep the legacy boolean in sync: only "none" disables TLS outright.
		cfg.Email.TLS = cfg.Email.TLSMode != "none"
	}
}

// Overrides holds CLI-flag values that take precedence over environment
// variables, the config file, and defaults (see AI.md PART 5: "Flags > env >
// file > defaults"). A nil field means "flag not provided" and leaves the
// underlying config value untouched, including when the flag can express an
// explicit false (see TriBoolFlag).
type Overrides struct {
	Address *string
	Port    *int
	Mode    *string
	Debug   *bool
}

// ApplyOverrides applies CLI-flag overrides on top of the already-loaded
// (file + env + defaults) configuration. This is the final, highest-priority
// layer in the precedence chain and must be called after Load returns.
func (cfg *Config) ApplyOverrides(ov Overrides) {
	if ov.Address != nil && *ov.Address != "" {
		cfg.Server.Address = *ov.Address
	}
	if ov.Port != nil && *ov.Port != 0 {
		cfg.Server.Port = *ov.Port
	}
	if ov.Mode != nil && *ov.Mode != "" {
		cfg.Server.Mode = *ov.Mode
	}
	if ov.Debug != nil {
		// Explicit --debug or --debug=false always wins, even over a
		// config-file/env value of true.
		cfg.Server.Debug = *ov.Debug
	}
}

// Save writes the configuration to server.yml
func (cfg *Config) Save() error {
	// Ensure config directory exists
	if err := os.MkdirAll(cfg.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := cfg.MarshalCommented()
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configFile := filepath.Join(cfg.ConfigDir, "server.yml")
	if err := writeFileFn(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// determineConfigDir resolves the configuration directory
func determineConfigDir(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return dir
	}

	// Check for portable mode
	if _, err := os.Stat("./config"); err == nil {
		return "./config"
	}

	// System or user install
	if getEUID() == 0 {
		return "/etc/casapps/cassocial"
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "casapps", "cassocial")
}

// determineDataDir resolves the data directory
func determineDataDir(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	if dir := os.Getenv("DATA_DIR"); dir != "" {
		return dir
	}

	// Check for portable mode
	if _, err := os.Stat("./data"); err == nil {
		return "./data"
	}

	// System or user install
	if getEUID() == 0 {
		return "/var/lib/casapps/cassocial"
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "casapps", "cassocial")
}

// determineLogDir resolves the log directory
func determineLogDir(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	if dir := os.Getenv("LOG_DIR"); dir != "" {
		return dir
	}

	// Check for portable mode
	if _, err := os.Stat("./logs"); err == nil {
		return "./logs"
	}

	// System or user install
	if getEUID() == 0 {
		return "/var/log/casapps/cassocial"
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "casapps", "cassocial", "logs")
}

// DeterminePIDFile resolves the PID file path
func DeterminePIDFile(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	if file := os.Getenv("CASSOCIAL_PID"); file != "" {
		return file
	}

	// Check for portable mode
	if _, err := os.Stat("./data"); err == nil {
		return "./cassocial.pid"
	}

	// System or user install
	if getEUID() == 0 {
		return "/var/run/casapps/cassocial.pid"
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "casapps", "cassocial", "cassocial.pid")
}
