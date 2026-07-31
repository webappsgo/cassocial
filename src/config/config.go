package config

import (
	"fmt"
	"os"
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

	// Internal fields (not in YAML)
	ConfigDir string `yaml:"-"`
	DataDir   string `yaml:"-"`
	LogDir    string `yaml:"-"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
	Mode    string `yaml:"mode"`    // production or development
	Debug   bool   `yaml:"debug"`
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
	From     string `yaml:"from"`
	TLS      bool   `yaml:"tls"`
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

	cfg.Cassocial.SiteName = "Cassocial"
	cfg.Cassocial.SiteDescription = "Self-hosted link aggregator and social profile"
	cfg.Cassocial.AllowRegistration = true
	cfg.Cassocial.MaxProfilesPerUser = 5
	cfg.Cassocial.MaxLinksPerProfile = 100

	return nil
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
	if from := os.Getenv("SMTP_FROM_EMAIL"); from != "" {
		cfg.Email.From = from
	}
}

// Save writes the configuration to server.yml
func (cfg *Config) Save() error {
	// Ensure config directory exists
	if err := os.MkdirAll(cfg.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
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
