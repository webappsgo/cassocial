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
