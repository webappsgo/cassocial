package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBool(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Truthy — exact matches
		{"one", "1", true},
		{"yes lowercase", "yes", true},
		{"true lowercase", "true", true},
		{"on lowercase", "on", true},
		{"enable lowercase", "enable", true},
		{"enabled lowercase", "enabled", true},
		{"y", "y", true},
		{"t", "t", true},
		{"yep", "yep", true},
		{"yup", "yup", true},
		{"yeah", "yeah", true},
		{"aye", "aye", true},
		{"si", "si", true},
		{"oui", "oui", true},
		{"da", "da", true},
		{"hai", "hai", true},
		{"affirmative", "affirmative", true},
		{"accept", "accept", true},
		{"allow", "allow", true},
		{"totally", "totally", true},

		// Truthy — uppercase variants (ToLower normalization)
		{"TRUE uppercase", "TRUE", true},
		{"YES uppercase", "YES", true},
		{"ON uppercase", "ON", true},
		{"ENABLED uppercase", "ENABLED", true},
		{"Yes mixed", "Yes", true},
		{"Enable mixed", "Enable", true},
		{"Enabled mixed", "Enabled", true},
		{"Affirmative mixed", "Affirmative", true},

		// Truthy — with surrounding whitespace
		{"whitespace before", " yes", true},
		{"whitespace after", "yes ", true},
		{"whitespace both", " true ", true},
		{"tab before", "\tyes", true},
		{"newline before", "\nyes", true},

		// Falsy — not in the truthy map
		{"false", "false", false},
		{"zero", "0", false},
		{"no", "no", false},
		{"off", "off", false},
		{"disable", "disable", false},
		{"disabled", "disabled", false},
		{"n", "n", false},
		{"f", "f", false},
		{"nope", "nope", false},
		{"nah", "nah", false},
		{"FALSE uppercase", "FALSE", false},
		{"NO uppercase", "NO", false},
		{"OFF uppercase", "OFF", false},
		{"DISABLE uppercase", "DISABLE", false},
		{"DISABLED uppercase", "DISABLED", false},

		// Edge cases
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"random word", "maybe", false},
		{"number two", "2", false},
		{"negative", "-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseBool(tt.input)
			if got != tt.want {
				t.Errorf("ParseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---- loadFromFile ----

func TestLoadFromFile_ValidYAML(t *testing.T) {
	tmp := t.TempDir()
	yamlContent := `server:
  address: "127.0.0.1"
  port: 9090
  mode: "development"
  debug: true
database:
  driver: "sqlite"
  max_connections: 20
logging:
  level: "debug"
  format: "json"
`
	configFile := filepath.Join(tmp, "server.yml")
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg := &Config{}
	if err := loadFromFile(configFile, cfg); err != nil {
		t.Fatalf("loadFromFile() returned error: %v", err)
	}

	if cfg.Server.Address != "127.0.0.1" {
		t.Errorf("Server.Address = %q, want 127.0.0.1", cfg.Server.Address)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Mode != "development" {
		t.Errorf("Server.Mode = %q, want development", cfg.Server.Mode)
	}
	if !cfg.Server.Debug {
		t.Error("Server.Debug should be true")
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want debug", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want json", cfg.Logging.Format)
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	cfg := &Config{}
	err := loadFromFile("/nonexistent/path/server.yml", cfg)
	if err == nil {
		t.Error("loadFromFile() with missing file should return error")
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	configFile := filepath.Join(tmp, "server.yml")
	if err := os.WriteFile(configFile, []byte("{\nnot valid yaml: [unclosed"), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg := &Config{}
	err := loadFromFile(configFile, cfg)
	if err == nil {
		t.Error("loadFromFile() with invalid YAML should return error")
	}
}

// ---- loadFromEnv ----

func TestLoadFromEnv_ServerSettings(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}

	t.Setenv("CASSOCIAL_ADDRESS", "192.168.1.1")
	t.Setenv("CASSOCIAL_PORT", "7777")
	t.Setenv("CASSOCIAL_MODE", "development")
	t.Setenv("CASSOCIAL_DEBUG", "true")

	cfg.loadFromEnv()

	if cfg.Server.Address != "192.168.1.1" {
		t.Errorf("Server.Address = %q, want 192.168.1.1", cfg.Server.Address)
	}
	if cfg.Server.Port != 7777 {
		t.Errorf("Server.Port = %d, want 7777", cfg.Server.Port)
	}
	if cfg.Server.Mode != "development" {
		t.Errorf("Server.Mode = %q, want development", cfg.Server.Mode)
	}
	if !cfg.Server.Debug {
		t.Error("Server.Debug should be true")
	}
}

func TestLoadFromEnv_DatabaseSettings(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}

	t.Setenv("CASSOCIAL_DB_DRIVER", "pgx")
	t.Setenv("CASSOCIAL_DB_HOST", "db.example.com")
	t.Setenv("CASSOCIAL_DB_PORT", "5432")
	t.Setenv("CASSOCIAL_DB_NAME", "mydb")
	t.Setenv("CASSOCIAL_DB_USER", "myuser")
	t.Setenv("CASSOCIAL_DB_PASSWORD", "secret")

	cfg.loadFromEnv()

	if cfg.Database.Driver != "pgx" {
		t.Errorf("Database.Driver = %q, want pgx", cfg.Database.Driver)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Database.Host = %q, want db.example.com", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want 5432", cfg.Database.Port)
	}
	if cfg.Database.Name != "mydb" {
		t.Errorf("Database.Name = %q, want mydb", cfg.Database.Name)
	}
	if cfg.Database.User != "myuser" {
		t.Errorf("Database.User = %q, want myuser", cfg.Database.User)
	}
	if cfg.Database.Password != "secret" {
		t.Errorf("Database.Password not set correctly")
	}
}

func TestLoadFromEnv_InvalidPort_Ignored(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}

	originalPort := cfg.Server.Port
	t.Setenv("CASSOCIAL_PORT", "not-a-number")
	cfg.loadFromEnv()

	if cfg.Server.Port != originalPort {
		t.Errorf("Server.Port = %d, want %d (invalid env should be ignored)", cfg.Server.Port, originalPort)
	}
}

func TestLoadFromEnv_InvalidDBPort_Ignored(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}

	originalPort := cfg.Database.Port
	t.Setenv("CASSOCIAL_DB_PORT", "not-a-number")
	cfg.loadFromEnv()

	if cfg.Database.Port != originalPort {
		t.Errorf("Database.Port = %d, want %d (invalid env should be ignored)", cfg.Database.Port, originalPort)
	}
}

// ---- Save ----

func TestSave_WritesFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := &Config{ConfigDir: tmp}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	configFile := filepath.Join(tmp, "server.yml")
	if _, err := os.Stat(configFile); err != nil {
		t.Errorf("Save() did not create config file: %v", err)
	}

	// Re-read and verify round-trip
	cfg2 := &Config{}
	if err := loadFromFile(configFile, cfg2); err != nil {
		t.Fatalf("loadFromFile() after Save() failed: %v", err)
	}
	if cfg2.Server.Address != cfg.Server.Address {
		t.Errorf("round-trip Server.Address = %q, want %q", cfg2.Server.Address, cfg.Server.Address)
	}
	if cfg2.Server.Port != cfg.Server.Port {
		t.Errorf("round-trip Server.Port = %d, want %d", cfg2.Server.Port, cfg.Server.Port)
	}
}

func TestSave_InvalidDir(t *testing.T) {
	// Attempt to save to a file path that cannot be created (parent is a file, not a dir)
	tmp := t.TempDir()
	blockingFile := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blockingFile, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create blocking file: %v", err)
	}

	// Use a path that goes through blockingFile as if it were a directory
	cfg := &Config{ConfigDir: filepath.Join(blockingFile, "subdir")}
	err := cfg.Save()
	if err == nil {
		t.Error("Save() should return error when config dir cannot be created")
	}
}

// ---- Validate ----

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() on default config returned error: %v", err)
	}
}

func TestValidate_InvalidPort_Zero(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}
	cfg.Server.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject port 0")
	}
}

func TestValidate_InvalidPort_TooHigh(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}
	cfg.Server.Port = 65536
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject port 65536")
	}
}

func TestValidate_InvalidMode(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}
	cfg.Server.Mode = "staging"
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject mode 'staging'")
	}
}

func TestValidate_InvalidDriver(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}
	cfg.Database.Driver = "redis"
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject database driver 'redis'")
	}
}

func TestValidate_AllValidDrivers(t *testing.T) {
	for _, driver := range []string{"sqlite", "pgx", "postgres", "mysql"} {
		cfg := &Config{}
		if err := cfg.setDefaults(); err != nil {
			t.Fatalf("setDefaults() failed: %v", err)
		}
		cfg.Database.Driver = driver
		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate() rejected valid driver %q: %v", driver, err)
		}
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}
	cfg.Logging.Level = "verbose"
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject log level 'verbose'")
	}
}

func TestValidate_AllValidLogLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg := &Config{}
		if err := cfg.setDefaults(); err != nil {
			t.Fatalf("setDefaults() failed: %v", err)
		}
		cfg.Logging.Level = level
		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate() rejected valid log level %q: %v", level, err)
		}
	}
}

func TestValidate_InvalidLogFormat(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}
	cfg.Logging.Format = "xml"
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject log format 'xml'")
	}
}

// ---- determineConfigDir / DataDir / LogDir ----

func TestDetermineConfigDir_FlagValue(t *testing.T) {
	got := determineConfigDir("/my/custom/config")
	if got != "/my/custom/config" {
		t.Errorf("determineConfigDir(flagValue) = %q, want /my/custom/config", got)
	}
}

func TestDetermineConfigDir_EnvVar(t *testing.T) {
	t.Setenv("CASSOCIAL_CONFIG", "/env/config/dir")
	t.Setenv("CASSOCIAL_DATA", "")
	got := determineConfigDir("")
	if got != "/env/config/dir" {
		t.Errorf("determineConfigDir from env = %q, want /env/config/dir", got)
	}
}

func TestDetermineConfigDir_Fallback(t *testing.T) {
	t.Setenv("CASSOCIAL_CONFIG", "")
	got := determineConfigDir("")
	// Must be non-empty regardless of how the fallback resolves
	if got == "" {
		t.Error("determineConfigDir() fallback returned empty string")
	}
}

func TestDetermineDataDir_FlagValue(t *testing.T) {
	got := determineDataDir("/my/custom/data")
	if got != "/my/custom/data" {
		t.Errorf("determineDataDir(flagValue) = %q, want /my/custom/data", got)
	}
}

func TestDetermineDataDir_EnvVar(t *testing.T) {
	t.Setenv("CASSOCIAL_DATA", "/env/data/dir")
	got := determineDataDir("")
	if got != "/env/data/dir" {
		t.Errorf("determineDataDir from env = %q, want /env/data/dir", got)
	}
}

func TestDetermineDataDir_Fallback(t *testing.T) {
	t.Setenv("CASSOCIAL_DATA", "")
	got := determineDataDir("")
	if got == "" {
		t.Error("determineDataDir() fallback returned empty string")
	}
}

func TestDetermineLogDir_FlagValue(t *testing.T) {
	got := determineLogDir("/my/custom/logs")
	if got != "/my/custom/logs" {
		t.Errorf("determineLogDir(flagValue) = %q, want /my/custom/logs", got)
	}
}

func TestDetermineLogDir_EnvVar(t *testing.T) {
	t.Setenv("CASSOCIAL_LOG", "/env/log/dir")
	got := determineLogDir("")
	if got != "/env/log/dir" {
		t.Errorf("determineLogDir from env = %q, want /env/log/dir", got)
	}
}

func TestDetermineLogDir_Fallback(t *testing.T) {
	t.Setenv("CASSOCIAL_LOG", "")
	got := determineLogDir("")
	if got == "" {
		t.Error("determineLogDir() fallback returned empty string")
	}
}

// ---- DeterminePIDFile ----

func TestDeterminePIDFile_FlagValue(t *testing.T) {
	got := DeterminePIDFile("/my/run/app.pid")
	if got != "/my/run/app.pid" {
		t.Errorf("DeterminePIDFile(flagValue) = %q, want /my/run/app.pid", got)
	}
}

func TestDeterminePIDFile_EnvVar(t *testing.T) {
	t.Setenv("CASSOCIAL_PID", "/env/run/app.pid")
	got := DeterminePIDFile("")
	if got != "/env/run/app.pid" {
		t.Errorf("DeterminePIDFile from env = %q, want /env/run/app.pid", got)
	}
}

func TestDeterminePIDFile_Fallback(t *testing.T) {
	t.Setenv("CASSOCIAL_PID", "")
	got := DeterminePIDFile("")
	if got == "" {
		t.Error("DeterminePIDFile() fallback returned empty string")
	}
}

// ---- non-root homeDir fallback paths ----

// withNonRootEUID temporarily overrides getEUID to return 1000 (non-root) so
// the homeDir branch in the determine* functions is reachable in tests.
func withNonRootEUID(t *testing.T) {
	t.Helper()
	orig := getEUID
	getEUID = func() int { return 1000 }
	t.Cleanup(func() { getEUID = orig })
}

func TestDetermineConfigDir_NonRootFallback(t *testing.T) {
	t.Setenv("CASSOCIAL_CONFIG", "")
	withNonRootEUID(t)
	got := determineConfigDir("")
	if got == "" {
		t.Error("determineConfigDir() non-root fallback returned empty string")
	}
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, ".config", "casapps", "cassocial")
	// portable path may be returned if ./config exists; only check homeDir path
	// when portable mode is not active
	if _, err := os.Stat("./config"); os.IsNotExist(err) {
		if got != want {
			t.Errorf("determineConfigDir() non-root = %q, want %q", got, want)
		}
	}
}

func TestDetermineDataDir_NonRootFallback(t *testing.T) {
	t.Setenv("CASSOCIAL_DATA", "")
	withNonRootEUID(t)
	got := determineDataDir("")
	if got == "" {
		t.Error("determineDataDir() non-root fallback returned empty string")
	}
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, ".local", "share", "casapps", "cassocial")
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		if got != want {
			t.Errorf("determineDataDir() non-root = %q, want %q", got, want)
		}
	}
}

func TestDetermineLogDir_NonRootFallback(t *testing.T) {
	t.Setenv("CASSOCIAL_LOG", "")
	withNonRootEUID(t)
	got := determineLogDir("")
	if got == "" {
		t.Error("determineLogDir() non-root fallback returned empty string")
	}
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, ".local", "share", "casapps", "cassocial", "logs")
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		if got != want {
			t.Errorf("determineLogDir() non-root = %q, want %q", got, want)
		}
	}
}

func TestDeterminePIDFile_NonRootFallback(t *testing.T) {
	t.Setenv("CASSOCIAL_PID", "")
	withNonRootEUID(t)
	got := DeterminePIDFile("")
	if got == "" {
		t.Error("DeterminePIDFile() non-root fallback returned empty string")
	}
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, ".local", "share", "casapps", "cassocial", "cassocial.pid")
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		if got != want {
			t.Errorf("DeterminePIDFile() non-root = %q, want %q", got, want)
		}
	}
}

// ---- Load ----

func TestLoad_LoadsExistingFile(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	dataDir := filepath.Join(tmp, "data")
	logDir := filepath.Join(tmp, "logs")

	// Pre-create config dir and write a valid file
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	yamlContent := `server:
  address: "10.0.0.1"
  port: 8888
  mode: "production"
database:
  driver: "sqlite"
logging:
  level: "info"
  format: "text"
`
	if err := os.WriteFile(filepath.Join(configDir, "server.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write pre-existing config: %v", err)
	}

	// Clear env overrides
	for _, env := range []string{"CASSOCIAL_ADDRESS", "CASSOCIAL_PORT", "CASSOCIAL_MODE", "CASSOCIAL_DEBUG", "CASSOCIAL_DB_DRIVER", "CASSOCIAL_DB_HOST", "CASSOCIAL_DB_PORT", "CASSOCIAL_DB_NAME", "CASSOCIAL_DB_USER", "CASSOCIAL_DB_PASSWORD"} {
		t.Setenv(env, "")
	}

	cfg, err := Load(configDir, dataDir, logDir)
	if err != nil {
		t.Fatalf("Load() with existing file returned error: %v", err)
	}
	if cfg.Server.Address != "10.0.0.1" {
		t.Errorf("Server.Address = %q, want 10.0.0.1", cfg.Server.Address)
	}
	if cfg.Server.Port != 8888 {
		t.Errorf("Server.Port = %d, want 8888", cfg.Server.Port)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	dataDir := filepath.Join(tmp, "data")
	logDir := filepath.Join(tmp, "logs")

	for _, env := range []string{"CASSOCIAL_ADDRESS", "CASSOCIAL_PORT", "CASSOCIAL_MODE", "CASSOCIAL_DEBUG", "CASSOCIAL_DB_DRIVER", "CASSOCIAL_DB_HOST", "CASSOCIAL_DB_PORT", "CASSOCIAL_DB_NAME", "CASSOCIAL_DB_USER", "CASSOCIAL_DB_PASSWORD"} {
		t.Setenv(env, "")
	}

	// First load to create defaults
	_, err := Load(configDir, dataDir, logDir)
	if err != nil {
		t.Fatalf("first Load() failed: %v", err)
	}

	// Now set env override and reload
	t.Setenv("CASSOCIAL_ADDRESS", "10.1.2.3")
	t.Setenv("CASSOCIAL_PORT", "9999")

	cfg, err := Load(configDir, dataDir, logDir)
	if err != nil {
		t.Fatalf("second Load() failed: %v", err)
	}
	if cfg.Server.Address != "10.1.2.3" {
		t.Errorf("Server.Address = %q, want 10.1.2.3 (env override)", cfg.Server.Address)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("Server.Port = %d, want 9999 (env override)", cfg.Server.Port)
	}
}

// ---------------------------------------------------------------------------
// determineConfigDir / DataDir / LogDir / DeterminePIDFile — portable mode
// ---------------------------------------------------------------------------

// changeToTempDir changes the working directory to a fresh temp dir for the
// duration of the test, then restores the original. This lets us test the
// "portable mode" branch that checks for ./config, ./data, etc.
func changeToTempDir(t *testing.T) string {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("os.Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Logf("WARNING: could not restore cwd: %v", err)
		}
	})
	return tmp
}

func TestDetermineConfigDir_PortableMode(t *testing.T) {
	tmp := changeToTempDir(t)
	// Create ./config directory so portable mode triggers.
	if err := os.Mkdir(tmp+"/config", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	t.Setenv("CASSOCIAL_CONFIG", "")
	got := determineConfigDir("")
	if got != "./config" {
		t.Errorf("determineConfigDir portable mode = %q, want ./config", got)
	}
}

func TestDetermineDataDir_PortableMode(t *testing.T) {
	tmp := changeToTempDir(t)
	if err := os.Mkdir(tmp+"/data", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	t.Setenv("CASSOCIAL_DATA", "")
	got := determineDataDir("")
	if got != "./data" {
		t.Errorf("determineDataDir portable mode = %q, want ./data", got)
	}
}

func TestDetermineLogDir_PortableMode(t *testing.T) {
	tmp := changeToTempDir(t)
	if err := os.Mkdir(tmp+"/logs", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	t.Setenv("CASSOCIAL_LOG", "")
	got := determineLogDir("")
	if got != "./logs" {
		t.Errorf("determineLogDir portable mode = %q, want ./logs", got)
	}
}

func TestDeterminePIDFile_PortableMode(t *testing.T) {
	tmp := changeToTempDir(t)
	if err := os.Mkdir(tmp+"/data", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	t.Setenv("CASSOCIAL_PID", "")
	got := DeterminePIDFile("")
	if got != "./cassocial.pid" {
		t.Errorf("DeterminePIDFile portable mode = %q, want ./cassocial.pid", got)
	}
}

// ---------------------------------------------------------------------------
// Load — invalid YAML triggers error path
// ---------------------------------------------------------------------------

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	tmp := t.TempDir()
	configDir := tmp + "/config"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Write a structurally invalid YAML file.
	if err := os.WriteFile(configDir+"/server.yml", []byte("{\nnot: [valid yaml"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Load(configDir, tmp+"/data", tmp+"/logs")
	if err == nil {
		t.Error("Load() with invalid YAML should return an error")
	}
}

// ---------------------------------------------------------------------------
// ensureDirectories — error when a path component is a file, not a directory
// ---------------------------------------------------------------------------

func TestEnsureDirectories_Error(t *testing.T) {
	tmp := t.TempDir()
	// Create a regular file where a directory is expected so MkdirAll fails.
	blockingFile := tmp + "/blocker"
	if err := os.WriteFile(blockingFile, []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := &Config{
		ConfigDir: blockingFile + "/config", // cannot create: parent is a file
		DataDir:   tmp + "/data",
		LogDir:    tmp + "/logs",
	}
	if err := cfg.ensureDirectories(); err == nil {
		t.Error("ensureDirectories() should return an error when a path cannot be created")
	}
}

// ---------------------------------------------------------------------------
// Save — error on os.WriteFile (directory is read-only)
// ---------------------------------------------------------------------------

func TestSave_ReadOnlyDir_ReturnsError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test read-only directory as root")
	}
	tmp := t.TempDir()
	// Create the config dir and make it read-only so WriteFile fails.
	configDir := filepath.Join(tmp, "roconfig")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Chmod(configDir, 0555); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	defer os.Chmod(configDir, 0755) //nolint:errcheck

	cfg := &Config{ConfigDir: configDir}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults: %v", err)
	}
	if err := cfg.Save(); err == nil {
		t.Error("Save() to read-only dir should return error")
	}
}

// ---------------------------------------------------------------------------
// Load — Save fails during default config creation
// ---------------------------------------------------------------------------

func TestLoad_SaveFails_ReturnsError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test read-only directory as root")
	}
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "rocfg")
	dataDir := filepath.Join(tmp, "data")
	logDir := filepath.Join(tmp, "logs")

	// Pre-create configDir so ensureDirectories succeeds, then make it read-only
	// so Save() cannot write server.yml.
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll configDir: %v", err)
	}
	// Also create sub-dirs that ensureDirectories would create.
	for _, sub := range []string{"ssl", "security"} {
		if err := os.MkdirAll(filepath.Join(configDir, sub), 0755); err != nil {
			t.Fatalf("MkdirAll sub %s: %v", sub, err)
		}
	}
	for _, sub := range []string{"db", "backup"} {
		if err := os.MkdirAll(filepath.Join(dataDir, sub), 0755); err != nil {
			t.Fatalf("MkdirAll data sub %s: %v", sub, err)
		}
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("MkdirAll logDir: %v", err)
	}

	// Now lock the config dir so no server.yml can be written.
	if err := os.Chmod(configDir, 0555); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	defer os.Chmod(configDir, 0755) //nolint:errcheck

	_, err := Load(configDir, dataDir, logDir)
	if err == nil {
		t.Error("Load() should return error when Save() cannot write server.yml")
	}
}

// ---------------------------------------------------------------------------
// Load — Validate fails (env overrides produce invalid config)
// ---------------------------------------------------------------------------

func TestLoad_ValidateFails_ReturnsError(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	dataDir := filepath.Join(tmp, "data")
	logDir := filepath.Join(tmp, "logs")

	// Clear all env overrides first, then set an invalid port.
	for _, env := range []string{
		"CASSOCIAL_ADDRESS", "CASSOCIAL_MODE", "CASSOCIAL_DEBUG",
		"CASSOCIAL_DB_DRIVER", "CASSOCIAL_DB_HOST", "CASSOCIAL_DB_PORT",
		"CASSOCIAL_DB_NAME", "CASSOCIAL_DB_USER", "CASSOCIAL_DB_PASSWORD",
	} {
		t.Setenv(env, "")
	}
	t.Setenv("CASSOCIAL_PORT", "99999") // invalid port — triggers Validate error

	_, err := Load(configDir, dataDir, logDir)
	if err == nil {
		t.Error("Load() should return error when Validate fails")
	}
}

func TestLoadDefaults(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	dataDir := filepath.Join(tmp, "data")
	logDir := filepath.Join(tmp, "logs")

	// Clear any environment overrides that could interfere
	for _, env := range []string{
		"CASSOCIAL_CONFIG", "CASSOCIAL_DATA", "CASSOCIAL_LOG",
		"CASSOCIAL_ADDRESS", "CASSOCIAL_PORT", "CASSOCIAL_MODE",
		"CASSOCIAL_DB_DRIVER",
	} {
		t.Setenv(env, "")
	}
	// Restore after test — t.Setenv does this automatically via Cleanup.

	cfg, err := Load(configDir, dataDir, logDir)
	if err != nil {
		t.Fatalf("Load(%q, %q, %q) returned error: %v", configDir, dataDir, logDir, err)
	}

	// Server address must be non-empty
	if cfg.Server.Address == "" {
		t.Errorf("Server.Address is empty, want non-empty default")
	}

	// Port must be in the valid 1–65535 range
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		t.Errorf("Server.Port = %d, want a value in 1–65535", cfg.Server.Port)
	}

	// Database driver must be one of the accepted values
	validDrivers := map[string]bool{"sqlite": true, "pgx": true, "postgres": true, "mysql": true}
	if !validDrivers[cfg.Database.Driver] {
		t.Errorf("Database.Driver = %q, want one of sqlite|pgx|postgres|mysql", cfg.Database.Driver)
	}

	// Logging level must be set
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		t.Errorf("Logging.Level = %q, want one of debug|info|warn|error", cfg.Logging.Level)
	}

	// Verify config file was actually written
	configFile := filepath.Join(configDir, "server.yml")
	if _, err := os.Stat(configFile); err != nil {
		t.Errorf("expected config file %s to exist after Load: %v", configFile, err)
	}
}

func TestSave_WriteFileFails(t *testing.T) {
	// Override writeFileFn to simulate a WriteFile error.
	orig := writeFileFn
	writeFileFn = func(_ string, _ []byte, _ os.FileMode) error {
		return os.ErrPermission
	}
	t.Cleanup(func() { writeFileFn = orig })

	tmp := t.TempDir()
	cfg := &Config{ConfigDir: tmp}
	if err := cfg.setDefaults(); err != nil {
		t.Fatalf("setDefaults() failed: %v", err)
	}
	err := cfg.Save()
	if err == nil {
		t.Error("Save() should return error when WriteFile fails")
	}
}

func TestLoad_EnsureDirectoriesFails(t *testing.T) {
	// Use a path inside /proc that cannot be created even as root
	// (the kernel does not allow creating entries in /proc via MkdirAll).
	_, err := Load("/proc/cassocial-test-ensure-fail", "", "")
	if err == nil {
		t.Error("Load() should return error when ensureDirectories fails")
	}
}

func TestLoad_SaveDefaultsFails(t *testing.T) {
	// Override writeFileFn so that Save() fails, causing Load() to return an error
	// on the "save defaults" path (no existing config file).
	orig := writeFileFn
	writeFileFn = func(_ string, _ []byte, _ os.FileMode) error {
		return os.ErrPermission
	}
	t.Cleanup(func() { writeFileFn = orig })

	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	dataDir := filepath.Join(tmp, "data")
	logDir := filepath.Join(tmp, "logs")

	// Do NOT create a server.yml — Load will try setDefaults then Save, which fails.
	_, err := Load(configDir, dataDir, logDir)
	if err == nil {
		t.Error("Load() should return error when Save() of defaults fails")
	}
}

