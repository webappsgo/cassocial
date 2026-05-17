package service

import (
	"image/color"
	"testing"

	"github.com/casapps/cassocial/src/server/model"
)

// newTestQRService creates a QRService with nil db for testing pure helper methods.
func newTestQRService() *QRService {
	return &QRService{db: nil}
}

func TestQRService_GetDefaultSettings(t *testing.T) {
	s := newTestQRService()
	cfg := s.GetDefaultSettings()
	if cfg == nil {
		t.Fatal("GetDefaultSettings() returned nil")
	}
	if cfg.Size != 256 {
		t.Errorf("Size = %d, want 256", cfg.Size)
	}
	if cfg.ErrorCorrection != "M" {
		t.Errorf("ErrorCorrection = %q, want M", cfg.ErrorCorrection)
	}
	if cfg.Format != "png" {
		t.Errorf("Format = %q, want png", cfg.Format)
	}
	if cfg.LogoEnabled {
		t.Error("LogoEnabled should be false by default")
	}
}

func TestQRService_MapErrorCorrection(t *testing.T) {
	s := newTestQRService()
	tests := []struct {
		level string
	}{
		{"L"},
		{"M"},
		{"Q"},
		{"H"},
		{"X"}, // invalid - should default to Medium
		{""},
	}
	for _, tt := range tests {
		// Just verify no panic
		_ = s.mapErrorCorrection(tt.level)
	}
}

func TestQRService_ParseColor_Valid6Char(t *testing.T) {
	s := newTestQRService()
	c, err := s.parseColor("#ff0000")
	if err != nil {
		t.Errorf("parseColor(#ff0000) error = %v, want nil", err)
	}
	r, g, b, a := c.RGBA()
	// RGBA() returns 16-bit values; red should be max
	if r == 0 {
		t.Error("red channel should be non-zero for #ff0000")
	}
	if g != 0 || b != 0 {
		t.Errorf("green and blue should be 0 for #ff0000, got g=%d b=%d", g, b)
	}
	if a == 0 {
		t.Error("alpha should be non-zero")
	}
}

func TestQRService_ParseColor_Valid3Char(t *testing.T) {
	s := newTestQRService()
	c, err := s.parseColor("#fff")
	if err != nil {
		t.Errorf("parseColor(#fff) error = %v, want nil", err)
	}
	if c == nil {
		t.Fatal("parseColor(#fff) returned nil color")
	}
}

func TestQRService_ParseColor_NoHash(t *testing.T) {
	s := newTestQRService()
	_, err := s.parseColor("ff0000")
	if err == nil {
		t.Error("parseColor without # should return error")
	}
}

func TestQRService_ParseColor_Empty(t *testing.T) {
	s := newTestQRService()
	_, err := s.parseColor("")
	if err == nil {
		t.Error("parseColor empty string should return error")
	}
}

func TestQRService_ParseColor_WrongLength(t *testing.T) {
	s := newTestQRService()
	_, err := s.parseColor("#12345")
	if err == nil {
		t.Error("parseColor wrong length should return error")
	}
}

func TestQRService_ParseColor_Black(t *testing.T) {
	s := newTestQRService()
	c, err := s.parseColor("#000000")
	if err != nil {
		t.Errorf("parseColor(#000000) error = %v", err)
	}
	rgba, ok := c.(color.RGBA)
	if !ok {
		t.Fatal("parseColor should return color.RGBA")
	}
	if rgba.R != 0 || rgba.G != 0 || rgba.B != 0 {
		t.Errorf("black should have R=G=B=0, got %+v", rgba)
	}
}

func TestQRCodeSettings_Validate(t *testing.T) {
	// Valid settings
	cfg := &model.QRCodeSettings{
		Size:            256,
		ErrorCorrection: "M",
		Format:          "png",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid settings Validate() = %v, want nil", err)
	}
}
