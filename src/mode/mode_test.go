package mode

import (
	"testing"
)

func TestCurrent_Default(t *testing.T) {
	t.Setenv("MODE", "")
	if Current() != Production {
		t.Errorf("Current() with no MODE = %q, want %q", Current(), Production)
	}
}

func TestCurrent_Production(t *testing.T) {
	for _, val := range []string{"production", "prod", "PRODUCTION", "PROD"} {
		t.Run(val, func(t *testing.T) {
			t.Setenv("MODE", val)
			if Current() != Production {
				t.Errorf("Current() with MODE=%q = %q, want %q", val, Current(), Production)
			}
		})
	}
}

func TestCurrent_Development(t *testing.T) {
	for _, val := range []string{"development", "dev", "debug", "DEVELOPMENT", "DEV", "DEBUG"} {
		t.Run(val, func(t *testing.T) {
			t.Setenv("MODE", val)
			if Current() != Development {
				t.Errorf("Current() with MODE=%q = %q, want %q", val, Current(), Development)
			}
		})
	}
}

func TestCurrent_Unknown(t *testing.T) {
	t.Setenv("MODE", "staging")
	if Current() != Production {
		t.Errorf("Current() with unknown MODE = %q, want %q (default)", Current(), Production)
	}
}

func TestIsProduction(t *testing.T) {
	t.Setenv("MODE", "production")
	if !IsProduction() {
		t.Error("IsProduction() = false, want true in production mode")
	}

	t.Setenv("MODE", "development")
	if IsProduction() {
		t.Error("IsProduction() = true, want false in development mode")
	}
}

func TestIsDevelopment(t *testing.T) {
	t.Setenv("MODE", "dev")
	if !IsDevelopment() {
		t.Error("IsDevelopment() = false, want true in dev mode")
	}

	t.Setenv("MODE", "production")
	if IsDevelopment() {
		t.Error("IsDevelopment() = true, want false in production mode")
	}
}

func TestIsDebug(t *testing.T) {
	for _, val := range []string{"1", "yes", "true", "on", "enable", "enabled", "y", "t"} {
		t.Run("truthy_"+val, func(t *testing.T) {
			t.Setenv("DEBUG", val)
			if !IsDebug() {
				t.Errorf("IsDebug() = false with DEBUG=%q, want true", val)
			}
		})
	}

	for _, val := range []string{"0", "no", "false", "off", "", "disable", "disabled"} {
		t.Run("falsy_"+val, func(t *testing.T) {
			t.Setenv("DEBUG", val)
			if IsDebug() {
				t.Errorf("IsDebug() = true with DEBUG=%q, want false", val)
			}
		})
	}
}

func TestModeString(t *testing.T) {
	if Production.String() != "production" {
		t.Errorf("Production.String() = %q, want %q", Production.String(), "production")
	}
	if Development.String() != "development" {
		t.Errorf("Development.String() = %q, want %q", Development.String(), "development")
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		input string
		want  Mode
	}{
		{"development", Development},
		{"dev", Development},
		{"debug", Development},
		{"DEVELOPMENT", Development},
		{"DEV", Development},
		{"DEBUG", Development},
		{"production", Production},
		{"prod", Production},
		{"PRODUCTION", Production},
		{"PROD", Production},
		{"unknown", Production},
		{"", Production},
		{"staging", Production},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FromString(tt.input)
			if got != tt.want {
				t.Errorf("FromString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
