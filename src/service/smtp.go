package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/casapps/cassocial/src/server/model"
)

var (
	ErrInvalidConfig     = errors.New("invalid SMTP configuration")
	ErrConnectionFailed  = errors.New("SMTP connection failed")
	ErrAuthFailed        = errors.New("SMTP authentication failed")
	ErrSendFailed        = errors.New("failed to send email")
	ErrInvalidProvider   = errors.New("invalid SMTP provider")
)

// SMTPProvider represents supported email providers
type SMTPProvider string

const (
	ProviderCustom  SMTPProvider = "CUSTOM"
	ProviderGmail   SMTPProvider = "Gmail"
	ProviderYahoo   SMTPProvider = "Yahoo"
	ProviderOutlook SMTPProvider = "Outlook"
)

// SecurityType represents SMTP security methods
type SecurityType string

const (
	SecurityNone     SecurityType = "NONE"
	SecuritySTARTTLS SecurityType = "STARTTLS"
	SecuritySSLTLS   SecurityType = "SSL/TLS"
)

// PortConfig represents predefined port/security combinations
type PortConfig struct {
	Port     int
	Security SecurityType
	Label    string
}

var (
	// Predefined port configurations matching SPEC dropdown
	PortConfigs = []PortConfig{
		{Port: 25, Security: SecurityNone, Label: "25 (NONE)"},
		{Port: 587, Security: SecuritySTARTTLS, Label: "587 (STARTTLS)"},
		{Port: 465, Security: SecuritySSLTLS, Label: "465 (SSL/TLS)"},
		{Port: 2525, Security: SecuritySTARTTLS, Label: "2525 (STARTTLS)"},
	}

	// Provider host mappings
	ProviderHosts = map[SMTPProvider]string{
		ProviderGmail:   "smtp.gmail.com",
		ProviderYahoo:   "smtp.mail.yahoo.com",
		ProviderOutlook: "smtp-mail.outlook.com",
	}
)

// Client represents an SMTP client
type Client struct {
	config *models.SMTPConfig
}

// NewClient creates a new SMTP client
func NewClient(config *models.SMTPConfig) (*Client, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	return &Client{
		config: config,
	}, nil
}

// GetProviderHost returns the SMTP host for a provider
func GetProviderHost(provider SMTPProvider) (string, error) {
	if provider == ProviderCustom {
		return "", nil // Custom provider requires manual host entry
	}

	host, ok := ProviderHosts[provider]
	if !ok {
		return "", ErrInvalidProvider
	}

	return host, nil
}

// TestConnection tests the SMTP connection and authentication
func (c *Client) TestConnection() error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// Test connection based on security type
	switch SecurityType(c.config.Security) {
	case SecuritySSLTLS:
		return c.testTLSConnection(addr)
	case SecuritySTARTTLS, SecurityNone:
		return c.testPlainConnection(addr)
	default:
		return fmt.Errorf("unsupported security type: %s", c.config.Security)
	}
}

// testTLSConnection tests SSL/TLS connection (port 465)
func (c *Client) testTLSConnection(addr string) error {
	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: c.config.Host,
		MinVersion: tls.VersionTLS12,
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer client.Quit()

	// Test authentication if credentials provided
	if c.config.User != "" {
		auth := smtp.PlainAuth("", c.config.User, c.config.Password, c.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("%w: %v", ErrAuthFailed, err)
		}
	}

	return nil
}

// testPlainConnection tests plain or STARTTLS connection (ports 25, 587, 2525)
func (c *Client) testPlainConnection(addr string) error {
	// Set connection timeout
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer client.Quit()

	// Use STARTTLS if required
	if SecurityType(c.config.Security) == SecuritySTARTTLS {
		tlsConfig := &tls.Config{
			ServerName: c.config.Host,
			MinVersion: tls.VersionTLS12,
		}

		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("%w: STARTTLS failed: %v", ErrConnectionFailed, err)
		}
	}

	// Test authentication if credentials provided
	if c.config.User != "" {
		auth := smtp.PlainAuth("", c.config.User, c.config.Password, c.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("%w: %v", ErrAuthFailed, err)
		}
	}

	return nil
}

// Send sends an email
func (c *Client) Send(to []string, subject, body string, isHTML bool) error {
	if !c.config.Enabled {
		return errors.New("SMTP is disabled")
	}

	// Build message
	message := buildMessage(c.config.FromName, c.config.FromAddress, to, subject, body, isHTML)

	// Send based on security type
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	switch SecurityType(c.config.Security) {
	case SecuritySSLTLS:
		return c.sendWithTLS(addr, to, message)
	case SecuritySTARTTLS, SecurityNone:
		return c.sendWithPlain(addr, to, message)
	default:
		return fmt.Errorf("unsupported security type: %s", c.config.Security)
	}
}

// sendWithTLS sends email using SSL/TLS (port 465)
func (c *Client) sendWithTLS(addr string, to []string, message []byte) error {
	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: c.config.Host,
		MinVersion: tls.VersionTLS12,
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer client.Quit()

	// Authenticate
	if c.config.User != "" {
		auth := smtp.PlainAuth("", c.config.User, c.config.Password, c.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("%w: %v", ErrAuthFailed, err)
		}
	}

	// Send email
	return c.sendMessage(client, to, message)
}

// sendWithPlain sends email using plain or STARTTLS connection
func (c *Client) sendWithPlain(addr string, to []string, message []byte) error {
	// Connect
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer client.Quit()

	// Use STARTTLS if required
	if SecurityType(c.config.Security) == SecuritySTARTTLS {
		tlsConfig := &tls.Config{
			ServerName: c.config.Host,
			MinVersion: tls.VersionTLS12,
		}

		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("%w: STARTTLS failed: %v", ErrSendFailed, err)
		}
	}

	// Authenticate
	if c.config.User != "" {
		auth := smtp.PlainAuth("", c.config.User, c.config.Password, c.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("%w: %v", ErrAuthFailed, err)
		}
	}

	// Send email
	return c.sendMessage(client, to, message)
}

// sendMessage sends the actual message through the SMTP client
func (c *Client) sendMessage(client *smtp.Client, to []string, message []byte) error {
	// Set sender
	if err := client.Mail(c.config.FromAddress); err != nil {
		return fmt.Errorf("%w: setting sender failed: %v", ErrSendFailed, err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("%w: setting recipient %s failed: %v", ErrSendFailed, recipient, err)
		}
	}

	// Send message data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("%w: data command failed: %v", ErrSendFailed, err)
	}

	if _, err := w.Write(message); err != nil {
		w.Close()
		return fmt.Errorf("%w: writing message failed: %v", ErrSendFailed, err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("%w: closing message failed: %v", ErrSendFailed, err)
	}

	return nil
}

// buildMessage builds the email message
func buildMessage(fromName, fromAddr string, to []string, subject, body string, isHTML bool) []byte {
	var message string

	// From header
	if fromName != "" {
		message += fmt.Sprintf("From: %s <%s>\r\n", fromName, fromAddr)
	} else {
		message += fmt.Sprintf("From: %s\r\n", fromAddr)
	}

	// To header
	message += fmt.Sprintf("To: %s\r\n", to[0])
	if len(to) > 1 {
		for _, recipient := range to[1:] {
			message += fmt.Sprintf("Cc: %s\r\n", recipient)
		}
	}

	// Subject
	message += fmt.Sprintf("Subject: %s\r\n", subject)

	// MIME headers
	message += "MIME-Version: 1.0\r\n"
	if isHTML {
		message += "Content-Type: text/html; charset=UTF-8\r\n"
	} else {
		message += "Content-Type: text/plain; charset=UTF-8\r\n"
	}

	// Body
	message += "\r\n" + body

	return []byte(message)
}

// SendWithRetry sends an email with retry logic based on priority
func (c *Client) SendWithRetry(to []string, subject, body string, isHTML bool, retries int) error {
	var lastErr error

	for i := 0; i <= retries; i++ {
		err := c.Send(to, subject, body, isHTML)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on auth failures
		if errors.Is(err, ErrAuthFailed) {
			return err
		}

		// Wait before retry (exponential backoff)
		if i < retries {
			waitTime := time.Duration(c.config.RetryDelay*(i+1)) * time.Second
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("%w (after %d retries): %v", ErrSendFailed, retries, lastErr)
}
