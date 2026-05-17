package service

import (
	"strings"
	"testing"

	models "github.com/casapps/cassocial/src/server/model"
)

// ---- NewClient ----

func TestNewClient_NilConfig(t *testing.T) {
	_, err := NewClient(nil)
	if err == nil {
		t.Error("NewClient(nil) should return error")
	}
}

func TestNewClient_InvalidConfig_EmptyHost(t *testing.T) {
	cfg := &models.SMTPConfig{
		Host:        "",
		Port:        587,
		FromAddress: "from@example.com",
	}
	_, err := NewClient(cfg)
	if err == nil {
		t.Error("NewClient(empty host) should return error")
	}
}

func TestNewClient_ValidConfig(t *testing.T) {
	cfg := &models.SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		FromAddress: "from@example.com",
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient(valid) returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient(valid) returned nil client")
	}
}

// ---- GetProviderHost ----

func TestGetProviderHost_Custom(t *testing.T) {
	host, err := GetProviderHost(ProviderCustom)
	if err != nil {
		t.Errorf("GetProviderHost(Custom) returned error: %v", err)
	}
	if host != "" {
		t.Errorf("GetProviderHost(Custom) = %q, want empty string", host)
	}
}

func TestGetProviderHost_Gmail(t *testing.T) {
	host, err := GetProviderHost(ProviderGmail)
	if err != nil {
		t.Errorf("GetProviderHost(Gmail) returned error: %v", err)
	}
	if host != "smtp.gmail.com" {
		t.Errorf("GetProviderHost(Gmail) = %q, want smtp.gmail.com", host)
	}
}

func TestGetProviderHost_Yahoo(t *testing.T) {
	host, err := GetProviderHost(ProviderYahoo)
	if err != nil {
		t.Errorf("GetProviderHost(Yahoo) returned error: %v", err)
	}
	if host != "smtp.mail.yahoo.com" {
		t.Errorf("GetProviderHost(Yahoo) = %q, want smtp.mail.yahoo.com", host)
	}
}

func TestGetProviderHost_Outlook(t *testing.T) {
	host, err := GetProviderHost(ProviderOutlook)
	if err != nil {
		t.Errorf("GetProviderHost(Outlook) returned error: %v", err)
	}
	if host != "smtp-mail.outlook.com" {
		t.Errorf("GetProviderHost(Outlook) = %q, want smtp-mail.outlook.com", host)
	}
}

func TestGetProviderHost_Unknown(t *testing.T) {
	_, err := GetProviderHost("UnknownProvider")
	if err == nil {
		t.Error("GetProviderHost(unknown) should return error")
	}
}

// ---- buildMessage ----

func TestBuildMessage_PlainText(t *testing.T) {
	msg := buildMessage("Sender Name", "sender@example.com", []string{"recipient@example.com"}, "Test Subject", "Hello World", false)
	msgStr := string(msg)

	if !strings.Contains(msgStr, "From: Sender Name <sender@example.com>") {
		t.Errorf("buildMessage() missing From header, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "To: recipient@example.com") {
		t.Errorf("buildMessage() missing To header, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "Subject: Test Subject") {
		t.Errorf("buildMessage() missing Subject header, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "Content-Type: text/plain") {
		t.Errorf("buildMessage() missing text/plain Content-Type, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "Hello World") {
		t.Errorf("buildMessage() missing body, got:\n%s", msgStr)
	}
}

func TestBuildMessage_HTML(t *testing.T) {
	msg := buildMessage("", "sender@example.com", []string{"recipient@example.com"}, "HTML Test", "<h1>Hi</h1>", true)
	msgStr := string(msg)

	if !strings.Contains(msgStr, "Content-Type: text/html") {
		t.Errorf("buildMessage(isHTML=true) missing text/html Content-Type, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "From: sender@example.com") {
		t.Errorf("buildMessage(noFromName) missing bare From header, got:\n%s", msgStr)
	}
}

func TestBuildMessage_MultipleRecipients(t *testing.T) {
	msg := buildMessage("Sender", "s@example.com", []string{"a@example.com", "b@example.com", "c@example.com"}, "Multi", "body", false)
	msgStr := string(msg)

	// First recipient in To:
	if !strings.Contains(msgStr, "To: a@example.com") {
		t.Errorf("buildMessage() missing primary recipient, got:\n%s", msgStr)
	}
	// Additional recipients as Cc:
	if !strings.Contains(msgStr, "Cc: b@example.com") {
		t.Errorf("buildMessage() missing Cc for b@example.com, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "Cc: c@example.com") {
		t.Errorf("buildMessage() missing Cc for c@example.com, got:\n%s", msgStr)
	}
}

func TestBuildMessage_SingleRecipient_NoCc(t *testing.T) {
	msg := buildMessage("S", "s@example.com", []string{"a@example.com"}, "Single", "body", false)
	msgStr := string(msg)

	if strings.Contains(msgStr, "Cc:") {
		t.Errorf("buildMessage() with single recipient should not have Cc header, got:\n%s", msgStr)
	}
}

func TestBuildMessage_MIMEVersion(t *testing.T) {
	msg := buildMessage("S", "s@example.com", []string{"r@example.com"}, "Sub", "body", false)
	if !strings.Contains(string(msg), "MIME-Version: 1.0") {
		t.Error("buildMessage() missing MIME-Version header")
	}
}

// ---- PortConfigs and constants ----

func TestPortConfigs_Count(t *testing.T) {
	if len(PortConfigs) != 4 {
		t.Errorf("PortConfigs has %d entries, want 4", len(PortConfigs))
	}
}

func TestPortConfigs_Values(t *testing.T) {
	ports := map[int]SecurityType{
		25:   SecurityNone,
		587:  SecuritySTARTTLS,
		465:  SecuritySSLTLS,
		2525: SecuritySTARTTLS,
	}
	for _, pc := range PortConfigs {
		expected, ok := ports[pc.Port]
		if !ok {
			t.Errorf("unexpected port %d in PortConfigs", pc.Port)
			continue
		}
		if pc.Security != expected {
			t.Errorf("PortConfigs[%d].Security = %q, want %q", pc.Port, pc.Security, expected)
		}
		if pc.Label == "" {
			t.Errorf("PortConfigs[%d].Label is empty", pc.Port)
		}
	}
}

// ---- TestConnection / testTLSConnection / testPlainConnection ----
// These make real network connections, so we only test the error path:
// a connection to a guaranteed-unreachable address must return ErrConnectionFailed.

// unreachableAddr is a TCP address that will always be refused (port 1 is
// reserved/privileged and virtually never open on loopback).
const unreachableAddr = "127.0.0.1:1"

func newSMTPClientForHost(t *testing.T, host string, port int, security string) *Client {
	t.Helper()
	cfg := &models.SMTPConfig{
		Host:        host,
		Port:        port,
		FromAddress: "from@example.com",
		Security:    security,
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func TestConnection_SSLTLS_ConnectionFailed(t *testing.T) {
	client := newSMTPClientForHost(t, "127.0.0.1", 1, string(SecuritySSLTLS))

	err := client.TestConnection()
	if err == nil {
		t.Error("TestConnection(SSL/TLS to unreachable) should return error")
	}
}

func TestConnection_STARTTLS_ConnectionFailed(t *testing.T) {
	client := newSMTPClientForHost(t, "127.0.0.1", 1, string(SecuritySTARTTLS))

	err := client.TestConnection()
	if err == nil {
		t.Error("TestConnection(STARTTLS to unreachable) should return error")
	}
}

func TestConnection_None_ConnectionFailed(t *testing.T) {
	client := newSMTPClientForHost(t, "127.0.0.1", 1, string(SecurityNone))

	err := client.TestConnection()
	if err == nil {
		t.Error("TestConnection(NONE to unreachable) should return error")
	}
}

func TestConnection_UnknownSecurity(t *testing.T) {
	// Build directly to bypass NewClient validation of the security field.
	cfg := &models.SMTPConfig{
		Host:        "127.0.0.1",
		Port:        587,
		FromAddress: "from@example.com",
		Security:    "UNKNOWN_SECURITY",
	}
	client := &Client{config: cfg}

	err := client.TestConnection()
	if err == nil {
		t.Error("TestConnection with unknown security type should return error")
	}
}

func TestTestTLSConnection_UnreachableHost(t *testing.T) {
	client := newSMTPClientForHost(t, "127.0.0.1", 1, string(SecuritySSLTLS))

	err := client.testTLSConnection(unreachableAddr)
	if err == nil {
		t.Error("testTLSConnection to unreachable address should return error")
	}
}

func TestTestTLSConnection_ErrorWrapsConnectionFailed(t *testing.T) {
	client := newSMTPClientForHost(t, "127.0.0.1", 1, string(SecuritySSLTLS))

	err := client.testTLSConnection(unreachableAddr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error must wrap ErrConnectionFailed per the implementation.
	if !smtpErrWraps(err, ErrConnectionFailed) {
		t.Errorf("testTLSConnection error should wrap ErrConnectionFailed, got: %v", err)
	}
}

func TestTestPlainConnection_UnreachableHost(t *testing.T) {
	client := newSMTPClientForHost(t, "127.0.0.1", 1, string(SecurityNone))

	err := client.testPlainConnection(unreachableAddr)
	if err == nil {
		t.Error("testPlainConnection to unreachable address should return error")
	}
}

func TestTestPlainConnection_ErrorWrapsConnectionFailed(t *testing.T) {
	client := newSMTPClientForHost(t, "127.0.0.1", 1, string(SecurityNone))

	err := client.testPlainConnection(unreachableAddr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !smtpErrWraps(err, ErrConnectionFailed) {
		t.Errorf("testPlainConnection error should wrap ErrConnectionFailed, got: %v", err)
	}
}

// smtpErrWraps reports whether target appears anywhere in err's chain.
func smtpErrWraps(err error, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

// ---- Send (disabled config) ----
// Send checks c.config.Enabled first; a disabled config returns an error
// without touching the network.

func TestSend_DisabledConfig_ReturnsError(t *testing.T) {
	cfg := &models.SMTPConfig{
		Enabled:     false,
		Host:        "smtp.example.com",
		Port:        587,
		FromAddress: "from@example.com",
		Security:    string(SecuritySTARTTLS),
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = client.Send([]string{"to@example.com"}, "Subject", "body", false)
	if err == nil {
		t.Error("Send() on disabled config should return error")
	}
}

// ---- Send / sendWithTLS / sendWithPlain — error paths (no real SMTP) ----
// We use a valid-looking config but point at an unreachable port so the
// dial fails immediately, exercising the connection-error branches.

func newEnabledClient(t *testing.T, port int, security SecurityType) *Client {
	t.Helper()
	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        "127.0.0.1",
		Port:        port,
		FromAddress: "from@example.com",
		Security:    string(security),
		RetryDelay:  0, // zero delay so retries don't slow the test
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func TestSend_SSLTLS_ConnectionFailed(t *testing.T) {
	client := newEnabledClient(t, 1, SecuritySSLTLS)

	err := client.Send([]string{"to@example.com"}, "Sub", "body", true)
	if err == nil {
		t.Error("Send(SSL/TLS, unreachable) should return error")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("Send SSL/TLS error should wrap ErrSendFailed, got: %v", err)
	}
}

func TestSend_STARTTLS_ConnectionFailed(t *testing.T) {
	client := newEnabledClient(t, 1, SecuritySTARTTLS)

	err := client.Send([]string{"to@example.com"}, "Sub", "body", true)
	if err == nil {
		t.Error("Send(STARTTLS, unreachable) should return error")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("Send STARTTLS error should wrap ErrSendFailed, got: %v", err)
	}
}

func TestSend_None_ConnectionFailed(t *testing.T) {
	client := newEnabledClient(t, 1, SecurityNone)

	err := client.Send([]string{"to@example.com"}, "Sub", "body", false)
	if err == nil {
		t.Error("Send(NONE, unreachable) should return error")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("Send NONE error should wrap ErrSendFailed, got: %v", err)
	}
}

func TestSend_UnknownSecurity_ReturnsError(t *testing.T) {
	// Build client directly to inject unknown security type on enabled config.
	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        "127.0.0.1",
		Port:        587,
		FromAddress: "from@example.com",
		Security:    "UNKNOWN",
	}
	client := &Client{config: cfg}

	err := client.Send([]string{"to@example.com"}, "Sub", "body", false)
	if err == nil {
		t.Error("Send with unknown security type should return error")
	}
}

func TestSendWithTLS_UnreachableHost(t *testing.T) {
	client := newEnabledClient(t, 1, SecuritySSLTLS)

	err := client.sendWithTLS(unreachableAddr, []string{"to@example.com"}, []byte("msg"))
	if err == nil {
		t.Error("sendWithTLS to unreachable host should return error")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("sendWithTLS error should wrap ErrSendFailed, got: %v", err)
	}
}

func TestSendWithPlain_UnreachableHost_None(t *testing.T) {
	client := newEnabledClient(t, 1, SecurityNone)

	err := client.sendWithPlain(unreachableAddr, []string{"to@example.com"}, []byte("msg"))
	if err == nil {
		t.Error("sendWithPlain(NONE) to unreachable host should return error")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("sendWithPlain error should wrap ErrSendFailed, got: %v", err)
	}
}

func TestSendWithPlain_UnreachableHost_STARTTLS(t *testing.T) {
	client := newEnabledClient(t, 1, SecuritySTARTTLS)

	err := client.sendWithPlain(unreachableAddr, []string{"to@example.com"}, []byte("msg"))
	if err == nil {
		t.Error("sendWithPlain(STARTTLS) to unreachable host should return error")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("sendWithPlain STARTTLS error should wrap ErrSendFailed, got: %v", err)
	}
}

// ---- SendWithRetry ----

func TestSendWithRetry_RetriesAndReturnsError(t *testing.T) {
	// Use an unreachable host and zero retry delay; 2 retries means 3 total attempts.
	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        "127.0.0.1",
		Port:        1,
		FromAddress: "from@example.com",
		Security:    string(SecurityNone),
		RetryDelay:  0,
	}
	client := &Client{config: cfg}

	err := client.SendWithRetry([]string{"to@example.com"}, "Sub", "body", false, 2)
	if err == nil {
		t.Error("SendWithRetry should return error when all attempts fail")
	}
	// The wrapped error indicates "after N retries"
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("SendWithRetry error should wrap ErrSendFailed, got: %v", err)
	}
}

func TestSendWithRetry_ZeroRetries_ReturnsError(t *testing.T) {
	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        "127.0.0.1",
		Port:        1,
		FromAddress: "from@example.com",
		Security:    string(SecurityNone),
		RetryDelay:  0,
	}
	client := &Client{config: cfg}

	err := client.SendWithRetry([]string{"to@example.com"}, "Sub", "body", false, 0)
	if err == nil {
		t.Error("SendWithRetry(0 retries) should return error on connection failure")
	}
}

func TestSendWithRetry_DisabledSMTP_NoRetries(t *testing.T) {
	// When SMTP is disabled, Send returns immediately without a connection;
	// SendWithRetry should propagate that error without any exponential sleep.
	cfg := &models.SMTPConfig{
		Enabled:     false,
		Host:        "smtp.example.com",
		Port:        587,
		FromAddress: "from@example.com",
		Security:    string(SecuritySTARTTLS),
		RetryDelay:  0, // zero delay so the test does not sleep between retries
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = client.SendWithRetry([]string{"to@example.com"}, "Sub", "body", false, 3)
	if err == nil {
		t.Error("SendWithRetry on disabled SMTP should return error")
	}
}
