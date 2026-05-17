package service

import (
	"image/color"
	"strings"
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

// ---------------------------------------------------------------------------
// NewQRService
// ---------------------------------------------------------------------------

func TestNewQRService(t *testing.T) {
	svc := NewQRService(nil)
	if svc == nil {
		t.Fatal("NewQRService returned nil")
	}
}

// ---------------------------------------------------------------------------
// GenerateQRCode — PNG
// ---------------------------------------------------------------------------

func TestGenerateQRCode_PNG(t *testing.T) {
	svc := newTestQRService()
	settings := svc.GetDefaultSettings() // format=png, size=256

	data, err := svc.GenerateQRCode("https://example.com", settings)
	if err != nil {
		t.Fatalf("GenerateQRCode(png): %v", err)
	}
	if len(data) == 0 {
		t.Error("GenerateQRCode(png) returned empty data")
	}
	// PNG magic bytes: 0x89 0x50 0x4E 0x47
	if len(data) < 4 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Error("GenerateQRCode(png) data does not start with PNG magic bytes")
	}
}

func TestGenerateQRCode_PNG_WithColors(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:            128,
		ErrorCorrection: "H",
		Style:           "square",
		DarkColor:       "#1a1a2e",
		LightColor:      "#ffffff",
		Format:          "png",
	}

	data, err := svc.GenerateQRCode("https://example.com/test", settings)
	if err != nil {
		t.Fatalf("GenerateQRCode custom colors: %v", err)
	}
	if len(data) == 0 {
		t.Error("GenerateQRCode custom colors returned empty data")
	}
}

func TestGenerateQRCode_PNG_RoundedStyle(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:            256,
		ErrorCorrection: "L",
		Style:           "rounded",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		Format:          "png",
	}

	data, err := svc.GenerateQRCode("https://example.com", settings)
	if err != nil {
		t.Fatalf("GenerateQRCode rounded: %v", err)
	}
	if len(data) == 0 {
		t.Error("GenerateQRCode rounded returned empty data")
	}
}

func TestGenerateQRCode_PNG_DotsStyle(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:            256,
		ErrorCorrection: "Q",
		Style:           "dots",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		Format:          "png",
	}

	data, err := svc.GenerateQRCode("https://example.com", settings)
	if err != nil {
		t.Fatalf("GenerateQRCode dots: %v", err)
	}
	if len(data) == 0 {
		t.Error("GenerateQRCode dots returned empty data")
	}
}

// ---------------------------------------------------------------------------
// GenerateQRCode — SVG
// ---------------------------------------------------------------------------

func TestGenerateQRCode_SVG(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:            256,
		ErrorCorrection: "M",
		Style:           "square",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		Format:          "svg",
	}

	data, err := svc.GenerateQRCode("https://example.com", settings)
	if err != nil {
		t.Fatalf("GenerateQRCode(svg): %v", err)
	}
	if len(data) == 0 {
		t.Error("GenerateQRCode(svg) returned empty data")
	}
	svg := string(data)
	if !strings.HasPrefix(svg, "<?xml") {
		t.Errorf("GenerateQRCode(svg) should start with <?xml, got: %q", svg[:min(20, len(svg))])
	}
}

// ---------------------------------------------------------------------------
// GenerateQRCode — invalid format
// ---------------------------------------------------------------------------

func TestGenerateQRCode_InvalidFormat(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:            256,
		ErrorCorrection: "M",
		Format:          "bmp", // unsupported
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
	}

	_, err := svc.GenerateQRCode("https://example.com", settings)
	if err == nil {
		t.Error("GenerateQRCode(invalid format) should return error")
	}
}

// ---------------------------------------------------------------------------
// GenerateQRCode — invalid settings (validation failure)
// ---------------------------------------------------------------------------

func TestGenerateQRCode_InvalidSettings(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:   0, // invalid size should fail Validate()
		Format: "png",
	}

	_, err := svc.GenerateQRCode("https://example.com", settings)
	if err == nil {
		t.Error("GenerateQRCode(invalid size=0) should return error")
	}
}

// ---------------------------------------------------------------------------
// GenerateBase64
// ---------------------------------------------------------------------------

func TestGenerateBase64(t *testing.T) {
	svc := newTestQRService()
	settings := svc.GetDefaultSettings()

	b64, err := svc.GenerateBase64("https://example.com", settings)
	if err != nil {
		t.Fatalf("GenerateBase64: %v", err)
	}
	if b64 == "" {
		t.Error("GenerateBase64 returned empty string")
	}
	// Valid base64 contains only [A-Za-z0-9+/=]
	for i, c := range b64 {
		if !isBase64Char(c) {
			t.Errorf("GenerateBase64 non-base64 char %q at index %d", c, i)
			break
		}
	}
}

func TestGenerateBase64_PropagatesError(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:   0, // will fail Validate
		Format: "png",
	}

	_, err := svc.GenerateBase64("https://example.com", settings)
	if err == nil {
		t.Error("GenerateBase64 with invalid settings should return error")
	}
}

// isBase64Char reports whether r is a valid base64 character.
func isBase64Char(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') || r == '+' || r == '/' || r == '='
}

// ---------------------------------------------------------------------------
// GenerateDataURI
// ---------------------------------------------------------------------------

func TestGenerateDataURI_PNG(t *testing.T) {
	svc := newTestQRService()
	settings := svc.GetDefaultSettings() // format=png

	uri, err := svc.GenerateDataURI("https://example.com", settings)
	if err != nil {
		t.Fatalf("GenerateDataURI(png): %v", err)
	}
	if uri == "" {
		t.Error("GenerateDataURI returned empty string")
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(uri, prefix) {
		t.Errorf("GenerateDataURI(png) = %q, want prefix %q", uri[:min(40, len(uri))], prefix)
	}
}

func TestGenerateDataURI_SVG(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:            256,
		ErrorCorrection: "M",
		Style:           "square",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		Format:          "svg",
	}

	uri, err := svc.GenerateDataURI("https://example.com", settings)
	if err != nil {
		t.Fatalf("GenerateDataURI(svg): %v", err)
	}
	const prefix = "data:image/svg+xml;base64,"
	if !strings.HasPrefix(uri, prefix) {
		t.Errorf("GenerateDataURI(svg) missing expected prefix %q, got: %q", prefix, uri[:min(40, len(uri))])
	}
}

func TestGenerateDataURI_PropagatesError(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:   0,
		Format: "png",
	}

	_, err := svc.GenerateDataURI("https://example.com", settings)
	if err == nil {
		t.Error("GenerateDataURI with invalid settings should return error")
	}
}

// ---------------------------------------------------------------------------
// generatePNG (private — tested through GenerateQRCode)
// Invalid colors fall back silently to black/white; no error expected.
// ---------------------------------------------------------------------------

func TestGeneratePNG_InvalidColors_FallsBackToDefault(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:            128,
		ErrorCorrection: "M",
		Style:           "square",
		DarkColor:       "not-a-color",  // invalid
		LightColor:      "also-invalid", // invalid
		Format:          "png",
	}

	data, err := svc.GenerateQRCode("https://example.com", settings)
	if err != nil {
		t.Fatalf("generatePNG with invalid colors should use fallback, got: %v", err)
	}
	if len(data) == 0 {
		t.Error("generatePNG with invalid colors returned empty data")
	}
}

// ---------------------------------------------------------------------------
// generateSVG (private — tested through GenerateQRCode)
// ---------------------------------------------------------------------------

func TestGenerateSVG_ContainsSVGTag(t *testing.T) {
	svc := newTestQRService()
	settings := &model.QRCodeSettings{
		Size:            256,
		ErrorCorrection: "M",
		Style:           "square",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		Format:          "svg",
	}

	data, err := svc.GenerateQRCode("https://cassocial.example.com/profile/test", settings)
	if err != nil {
		t.Fatalf("generateSVG: %v", err)
	}

	svg := string(data)
	for _, want := range []string{"<svg", `xmlns="http://www.w3.org/2000/svg"`, "</svg>"} {
		if !strings.Contains(svg, want) {
			t.Errorf("generateSVG output missing %q; full output:\n%s", want, svg)
		}
	}
}

// ---------------------------------------------------------------------------
// GenerateWithLogo — LogoEnabled=false (delegates to GenerateQRCode)
// ---------------------------------------------------------------------------

func TestGenerateWithLogo_LogoDisabled(t *testing.T) {
	svc := newTestQRService()
	settings := svc.GetDefaultSettings()
	settings.LogoEnabled = false

	data, err := svc.GenerateWithLogo("https://example.com", settings, nil)
	if err != nil {
		t.Fatalf("GenerateWithLogo(logo disabled): %v", err)
	}
	if len(data) == 0 {
		t.Error("GenerateWithLogo(logo disabled) returned empty data")
	}
}

func TestGenerateWithLogo_LogoEnabled_InvalidLogoData(t *testing.T) {
	svc := newTestQRService()
	settings := svc.GetDefaultSettings()
	settings.LogoEnabled = true
	settings.LogoSize = 20

	// Pass garbage logo bytes — should fail to decode the logo
	_, err := svc.GenerateWithLogo("https://example.com", settings, []byte("not a valid image"))
	if err == nil {
		t.Error("GenerateWithLogo with invalid logo data should return error")
	}
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
