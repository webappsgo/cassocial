package ssl

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateSelfSignedCert creates a temporary self-signed cert+key pair and
// returns their file paths. The caller must not clean them up manually —
// t.TempDir() handles cleanup.
func generateSelfSignedCert(t *testing.T, notAfter time.Time) (certFile, keyFile string) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: failed to generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: failed to create certificate: %v", err)
	}

	tmp := t.TempDir()

	certFile = filepath.Join(tmp, "cert.pem")
	keyFile = filepath.Join(tmp, "key.pem")

	certOut, err := os.Create(certFile)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: failed to create cert file: %v", err)
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		t.Fatalf("generateSelfSignedCert: failed to encode cert: %v", err)
	}

	keyOut, err := os.Create(keyFile)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: failed to create key file: %v", err)
	}
	defer keyOut.Close()
	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: failed to marshal key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		t.Fatalf("generateSelfSignedCert: failed to encode key: %v", err)
	}

	return certFile, keyFile
}

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

// ---- getManualCertConfig ----

func TestGetManualCertConfig_MissingCertFile(t *testing.T) {
	m := NewManager(&Config{
		Enabled:  true,
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	})
	_, err := m.getManualCertConfig()
	if err == nil {
		t.Error("getManualCertConfig() with missing cert should return error")
	}
}

func TestGetManualCertConfig_MissingKeyFile(t *testing.T) {
	// Cert file exists but key does not
	certFile, _ := generateSelfSignedCert(t, time.Now().Add(365*24*time.Hour))

	m := NewManager(&Config{
		Enabled:  true,
		CertFile: certFile,
		KeyFile:  "/nonexistent/key.pem",
	})
	_, err := m.getManualCertConfig()
	if err == nil {
		t.Error("getManualCertConfig() with missing key should return error")
	}
}

func TestGetManualCertConfig_ValidCert(t *testing.T) {
	certFile, keyFile := generateSelfSignedCert(t, time.Now().Add(365*24*time.Hour))

	m := NewManager(&Config{
		Enabled:  true,
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	tlsCfg, err := m.getManualCertConfig()
	if err != nil {
		t.Fatalf("getManualCertConfig() returned error: %v", err)
	}
	if tlsCfg == nil {
		t.Fatal("getManualCertConfig() returned nil TLS config")
	}
	if len(tlsCfg.Certificates) != 1 {
		t.Errorf("tlsCfg.Certificates len = %d, want 1", len(tlsCfg.Certificates))
	}
}

// ---- GetTLSConfig with valid manual cert ----

func TestGetTLSConfig_ManualCerts_Valid(t *testing.T) {
	certFile, keyFile := generateSelfSignedCert(t, time.Now().Add(365*24*time.Hour))

	m := NewManager(&Config{
		Enabled:  true,
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	tlsCfg, err := m.GetTLSConfig()
	if err != nil {
		t.Fatalf("GetTLSConfig() with valid cert returned error: %v", err)
	}
	if tlsCfg == nil {
		t.Fatal("GetTLSConfig() with valid cert returned nil")
	}
}

// ---- CheckCertificateExpiry ----

func TestCheckCertificateExpiry_ValidCert_NotExpiringSoon(t *testing.T) {
	certFile, keyFile := generateSelfSignedCert(t, time.Now().Add(60*24*time.Hour)) // 60 days

	m := NewManager(&Config{
		Enabled:  true,
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	expiring, expiryTime, err := m.CheckCertificateExpiry()
	if err != nil {
		t.Fatalf("CheckCertificateExpiry() returned error: %v", err)
	}
	if expiring {
		t.Error("CheckCertificateExpiry() should return expiring=false for cert valid 60 days")
	}
	if expiryTime.IsZero() {
		t.Error("CheckCertificateExpiry() returned zero expiry time")
	}
}

func TestCheckCertificateExpiry_ValidCert_ExpiringSoon(t *testing.T) {
	certFile, keyFile := generateSelfSignedCert(t, time.Now().Add(10*24*time.Hour)) // 10 days

	m := NewManager(&Config{
		Enabled:  true,
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	expiring, _, err := m.CheckCertificateExpiry()
	if err != nil {
		t.Fatalf("CheckCertificateExpiry() returned error: %v", err)
	}
	if !expiring {
		t.Error("CheckCertificateExpiry() should return expiring=true for cert expiring in 10 days")
	}
}

func TestCheckCertificateExpiry_MissingCert(t *testing.T) {
	m := NewManager(&Config{
		Enabled:  true,
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	})
	_, _, err := m.CheckCertificateExpiry()
	if err == nil {
		t.Error("CheckCertificateExpiry() with missing cert should return error")
	}
}

// ---- parseX509Certificate ----

func TestParseX509Certificate_Valid(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(42),
		Subject:      pkix.Name{CommonName: "parsetest"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	cert, err := parseX509Certificate(der)
	if err != nil {
		t.Fatalf("parseX509Certificate() returned error: %v", err)
	}
	if cert == nil {
		t.Fatal("parseX509Certificate() returned nil cert")
	}
	if cert.Subject.CommonName != "parsetest" {
		t.Errorf("cert.Subject.CommonName = %q, want parsetest", cert.Subject.CommonName)
	}
}

func TestParseX509Certificate_Invalid(t *testing.T) {
	_, err := parseX509Certificate([]byte("not a valid DER certificate"))
	if err == nil {
		t.Error("parseX509Certificate() with invalid data should return error")
	}
}
