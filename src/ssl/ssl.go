package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Config represents SSL/TLS configuration
type Config struct {
	Enabled     bool
	CertFile    string
	KeyFile     string
	LetsEncrypt bool
	Domain      string
	Email       string
	DataDir     string
}

// Manager handles SSL/TLS certificates
type Manager struct {
	config *Config
}

// NewManager creates a new SSL manager
func NewManager(config *Config) *Manager {
	return &Manager{
		config: config,
	}
}

// GetTLSConfig returns a TLS configuration
func (m *Manager) GetTLSConfig() (*tls.Config, error) {
	if !m.config.Enabled {
		return nil, nil
	}

	// Check if using Let's Encrypt
	if m.config.LetsEncrypt {
		return m.getLetsEncryptConfig()
	}

	// Use manual certificates
	return m.getManualCertConfig()
}

// getManualCertConfig loads manual SSL certificates
func (m *Manager) getManualCertConfig() (*tls.Config, error) {
	// Verify cert and key files exist
	if _, err := os.Stat(m.config.CertFile); err != nil {
		return nil, fmt.Errorf("certificate file not found: %s", m.config.CertFile)
	}

	if _, err := os.Stat(m.config.KeyFile); err != nil {
		return nil, fmt.Errorf("key file not found: %s", m.config.KeyFile)
	}

	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}

// getLetsEncryptConfig sets up Let's Encrypt automatic certificates
func (m *Manager) getLetsEncryptConfig() (*tls.Config, error) {
	log.Println("Let's Encrypt support not yet implemented")
	return nil, fmt.Errorf("Let's Encrypt support coming soon")
}

// CheckCertificateExpiry checks if certificates are expiring soon
func (m *Manager) CheckCertificateExpiry() (bool, time.Time, error) {
	if !m.config.Enabled || m.config.LetsEncrypt {
		return false, time.Time{}, nil
	}

	// Load certificate
	cert, err := tls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Parse certificate to get expiry
	x509Cert, err := parseX509Certificate(cert.Certificate[0])
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if expiring within 30 days
	expiryTime := x509Cert.NotAfter
	daysUntilExpiry := time.Until(expiryTime).Hours() / 24

	return daysUntilExpiry < 30, expiryTime, nil
}

// parseX509CertificateFn is the function used to parse a DER-encoded certificate.
// Tests may replace it to inject a parse error.
var parseX509CertificateFn = x509.ParseCertificate

// parseX509Certificate parses DER-encoded certificate
func parseX509Certificate(der []byte) (*x509.Certificate, error) {
	return parseX509CertificateFn(der)
}

// GetCertPath returns the path to SSL certificates
func GetCertPath(configDir string) string {
	return filepath.Join(configDir, "ssl", "letsencrypt")
}

// GetLocalCertPath returns the path to manual certificates
func GetLocalCertPath(configDir string) string {
	return filepath.Join(configDir, "ssl", "local")
}
