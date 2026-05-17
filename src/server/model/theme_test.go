package model

import (
	"testing"
)

func TestTheme_Validate_Valid(t *testing.T) {
	th := &Theme{
		Name:            "Test",
		BackgroundColor: "#1a1a1a",
		TextColor:       "#ffffff",
		LinkBackground:  "#2a2a2a",
		LinkHover:       "#3a3a3a",
		LinkText:        "#ffffff",
	}
	if err := th.Validate(); err != nil {
		t.Errorf("valid theme Validate() = %v, want nil", err)
	}
}

func TestTheme_Validate_EmptyName(t *testing.T) {
	th := &Theme{Name: "", BackgroundColor: "#ffffff"}
	if err := th.Validate(); err != ErrThemeNameEmpty {
		t.Errorf("empty name Validate() = %v, want ErrThemeNameEmpty", err)
	}
}

func TestTheme_Validate_InvalidColor(t *testing.T) {
	th := &Theme{Name: "Bad", BackgroundColor: "notacolor"}
	if err := th.Validate(); err != ErrInvalidColor {
		t.Errorf("invalid color Validate() = %v, want ErrInvalidColor", err)
	}
}

func TestTheme_Validate_EmptyColorsAllowed(t *testing.T) {
	th := &Theme{Name: "Minimal"}
	if err := th.Validate(); err != nil {
		t.Errorf("empty colors Validate() = %v, want nil", err)
	}
}

func TestGetDefaultTheme(t *testing.T) {
	th := GetDefaultTheme()
	if th == nil {
		t.Fatal("GetDefaultTheme() returned nil")
	}
	if th.Name == "" {
		t.Error("default theme has empty name")
	}
	if th.ID != DefaultThemeID {
		t.Errorf("default theme ID = %q, want %q", th.ID, DefaultThemeID)
	}
	if th.IsPremium {
		t.Error("default theme should not be premium")
	}
}

func TestGetLightTheme(t *testing.T) {
	th := GetLightTheme()
	if th == nil {
		t.Fatal("GetLightTheme() returned nil")
	}
	if th.Name != "Light" {
		t.Errorf("light theme name = %q, want Light", th.Name)
	}
}

func TestGetDraculaTheme(t *testing.T) {
	th := GetDraculaTheme()
	if th == nil {
		t.Fatal("GetDraculaTheme() returned nil")
	}
	if th.Name != "Dracula" {
		t.Errorf("dracula theme name = %q, want Dracula", th.Name)
	}
}

func TestGetOceanTheme(t *testing.T) {
	th := GetOceanTheme()
	if th == nil {
		t.Fatal("GetOceanTheme() returned nil")
	}
	if th.Name != "Ocean" {
		t.Errorf("ocean theme name = %q, want Ocean", th.Name)
	}
	if !th.IsPremium {
		t.Error("ocean theme should be premium")
	}
}

func TestGetSunsetTheme(t *testing.T) {
	th := GetSunsetTheme()
	if th == nil {
		t.Fatal("GetSunsetTheme() returned nil")
	}
	if th.Name != "Sunset" {
		t.Errorf("sunset theme name = %q, want Sunset", th.Name)
	}
	if !th.IsPremium {
		t.Error("sunset theme should be premium")
	}
}

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		color string
		want  bool
	}{
		{"#ffffff", true},
		{"#000000", true},
		{"#fff", true},
		{"#abc", true},
		{"#AABBCC", true},
		{"ffffff", false},   // missing #
		{"#gggggg", false},  // invalid hex chars
		{"#12345", false},   // wrong length
		{"#12345678", false}, // too long
		{"", false},
		{"1234", false},   // length 4 but no leading #
		{"1234567", false}, // length 7 but no leading #
	}
	for _, tt := range tests {
		got := isValidHexColor(tt.color)
		if got != tt.want {
			t.Errorf("isValidHexColor(%q) = %v, want %v", tt.color, got, tt.want)
		}
	}
}

func TestTheme_ApplyToProfile(t *testing.T) {
	th := &Theme{
		BackgroundColor: "#1a1a1a",
		TextColor:       "#ffffff",
		LinkBackground:  "#2a2a2a",
		LinkHover:       "#3a3a3a",
		LinkText:        "#ffffff",
		BorderRadius:    "12px",
		FontFamily:      "Inter",
	}
	vars := th.ApplyToProfile()
	if vars == nil {
		t.Fatal("ApplyToProfile() returned nil")
	}
	if vars["--bg-primary"] != "#1a1a1a" {
		t.Errorf("--bg-primary = %q, want #1a1a1a", vars["--bg-primary"])
	}
	if vars["--text-primary"] != "#ffffff" {
		t.Errorf("--text-primary = %q, want #ffffff", vars["--text-primary"])
	}
	if vars["--radius"] != "12px" {
		t.Errorf("--radius = %q, want 12px", vars["--radius"])
	}
}
