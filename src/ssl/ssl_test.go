package ssl

import (
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	cfg := &Config{Enabled: false}
	m := NewManager(cfg)
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.config != cfg {
		t.Error("NewManager() did not store config")
	}
}

func TestGetTLSConfig_Disabled(t *testing.T) {
	m := NewManager(&Config{Enabled: false})
	tlsCfg, err := m.GetTLSConfig()
	if err != nil {
		t.Errorf("GetTLSConfig(disabled) error = %v, want nil", err)
	}
	if tlsCfg != nil {
		t.Error("GetTLSConfig(disabled) should return nil config")
	}
}

func TestGetTLSConfig_LetsEncrypt(t *testing.T) {
	m := NewManager(&Config{
		Enabled:     true,
		LetsEncrypt: true,
		Domain:      "example.com",
		Email:       "admin@example.com",
	})
	_, err := m.GetTLSConfig()
	// Let's Encrypt is not implemented — expect an error
	if err == nil {
		t.Error("GetTLSConfig(letsencrypt) should return error (not implemented)")
	}
}

func TestGetTLSConfig_ManualCerts_Missing(t *testing.T) {
	m := NewManager(&Config{
		Enabled:  true,
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	})
	_, err := m.GetTLSConfig()
	if err == nil {
		t.Error("GetTLSConfig with missing cert file should return error")
	}
}

func TestCheckCertificateExpiry_Disabled(t *testing.T) {
	m := NewManager(&Config{Enabled: false})
	expiring, expiry, err := m.CheckCertificateExpiry()
	if err != nil {
		t.Errorf("CheckCertificateExpiry(disabled) error = %v, want nil", err)
	}
	if expiring {
		t.Error("CheckCertificateExpiry(disabled) should return expiring=false")
	}
	if !expiry.IsZero() {
		t.Errorf("CheckCertificateExpiry(disabled) expiry = %v, want zero time", expiry)
	}
}

func TestCheckCertificateExpiry_LetsEncrypt(t *testing.T) {
	m := NewManager(&Config{Enabled: true, LetsEncrypt: true})
	expiring, _, err := m.CheckCertificateExpiry()
	if err != nil {
		t.Errorf("CheckCertificateExpiry(letsencrypt) error = %v, want nil", err)
	}
	if expiring {
		t.Error("CheckCertificateExpiry(letsencrypt) should return expiring=false")
	}
}

func TestGetCertPath(t *testing.T) {
	path := GetCertPath("/etc/myapp")
	expected := filepath.Join("/etc/myapp", "ssl", "letsencrypt")
	if path != expected {
		t.Errorf("GetCertPath() = %q, want %q", path, expected)
	}
}

func TestGetLocalCertPath(t *testing.T) {
	path := GetLocalCertPath("/etc/myapp")
	expected := filepath.Join("/etc/myapp", "ssl", "local")
	if path != expected {
		t.Errorf("GetLocalCertPath() = %q, want %q", path, expected)
	}
}
