package service

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	models "github.com/casapps/cassocial/src/server/model"
)

// smtpNewClient wraps smtp.NewClient — the stdlib function that turns a net.Conn
// into an *smtp.Client — so tests can call it without a direct import clash.
func smtpNewClient(conn net.Conn, host string) (*smtp.Client, error) {
	return smtp.NewClient(conn, host)
}

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

// ---------------------------------------------------------------------------
// fakeSMTPServer — minimal SMTP server for testing sendMessage without a real
// mail relay. It uses bufio for line-based reading and speaks just enough
// SMTP to let net/smtp.Client complete EHLO → MAIL FROM → RCPT TO → DATA → QUIT.
// ---------------------------------------------------------------------------

// fakeSMTPResponse describes how a fake server handles a given SMTP verb.
type fakeSMTPResponse struct {
	rejectMAIL          bool
	rejectRCPT          bool
	rejectDATA          bool
	closeAfterDATAAccept bool // send 354 then close connection immediately
}

// startFakeSMTPServerWith starts a minimal TCP SMTP listener on a random port.
// The goroutine handles exactly one connection using line-by-line reading.
// It closes doneCh when the connection handling goroutine exits.
func startFakeSMTPServerWith(t *testing.T, cfg fakeSMTPResponse) (addr string, done <-chan struct{}) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startFakeSMTPServerWith: listen: %v", err)
	}

	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		defer ln.Close()

		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

		writeLine := func(s string) {
			fmt.Fprintf(rw, "%s\r\n", s)
			rw.Flush() //nolint:errcheck
		}

		writeLine("220 localhost ESMTP")

		inData := false
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")

			if inData {
				if line == "." {
					writeLine("250 OK: queued")
					inData = false
				}
				// other data lines are silently consumed
				continue
			}

			switch {
			case strings.HasPrefix(strings.ToUpper(line), "EHLO"),
				strings.HasPrefix(strings.ToUpper(line), "HELO"):
				writeLine("250-localhost")
				writeLine("250 OK")
			case strings.HasPrefix(strings.ToUpper(line), "MAIL FROM"):
				if cfg.rejectMAIL {
					writeLine("550 Rejected by policy")
				} else {
					writeLine("250 OK")
				}
			case strings.HasPrefix(strings.ToUpper(line), "RCPT TO"):
				if cfg.rejectRCPT {
					writeLine("550 No such user")
				} else {
					writeLine("250 OK")
				}
			case strings.ToUpper(line) == "DATA":
				if cfg.rejectDATA {
					writeLine("554 Transaction failed")
				} else if cfg.closeAfterDATAAccept {
					// Accept DATA then immediately close the connection so
					// w.Write() in sendMessage encounters a broken pipe.
					writeLine("354 Start input, end with <CRLF>.<CRLF>")
					return
				} else {
					writeLine("354 Start input, end with <CRLF>.<CRLF>")
					inData = true
				}
			case strings.HasPrefix(strings.ToUpper(line), "QUIT"):
				writeLine("221 Bye")
				return
			case strings.HasPrefix(strings.ToUpper(line), "RSET"):
				writeLine("250 OK")
			default:
				writeLine("500 Command unrecognised")
			}
		}
	}()

	return ln.Addr().String(), doneCh
}

// clientForFakeServer builds an SMTP Client pointed at a fake server address.
func clientForFakeServer(t *testing.T, addr string) *Client {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("clientForFakeServer SplitHostPort(%q): %v", addr, err)
	}
	var port int
	if _, scanErr := fmt.Sscanf(portStr, "%d", &port); scanErr != nil {
		t.Fatalf("clientForFakeServer parse port %q: %v", portStr, scanErr)
	}
	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        host,
		Port:        port,
		FromAddress: "from@example.com",
		Security:    string(SecurityNone),
		RetryDelay:  0,
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("clientForFakeServer NewClient: %v", err)
	}
	return c
}

// ---------------------------------------------------------------------------
// sendMessage tests — require a real (fake) SMTP connection
// ---------------------------------------------------------------------------

func TestSendMessage_Success(t *testing.T) {
	addr, done := startFakeSMTPServerWith(t, fakeSMTPResponse{})

	client := clientForFakeServer(t, addr)
	err := client.Send([]string{"to@example.com"}, "Test Subject", "Hello World", false)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if err != nil {
		t.Errorf("Send() to fake SMTP server returned unexpected error: %v", err)
	}
}

func TestSendMessage_MailFromError(t *testing.T) {
	addr, done := startFakeSMTPServerWith(t, fakeSMTPResponse{rejectMAIL: true})

	// Dial and build an smtp.Client manually so we can call sendMessage directly.
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial fake server: %v", err)
	}

	smtpClient, err := smtpNewClient(conn, "127.0.0.1")
	if err != nil {
		conn.Close()
		t.Fatalf("smtp.NewClient: %v", err)
	}

	c := clientForFakeServer(t, addr)
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	sendErr := c.sendMessage(smtpClient, []string{"to@example.com"}, msg)

	// Close the connection so the fake server unblocks and its goroutine exits.
	smtpClient.Quit() //nolint:errcheck
	conn.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if sendErr == nil {
		t.Error("sendMessage should return error when MAIL FROM is rejected")
	}
	if !smtpErrWraps(sendErr, ErrSendFailed) {
		t.Errorf("sendMessage MAIL FROM error should wrap ErrSendFailed, got: %v", sendErr)
	}
}

func TestSendMessage_RcptToError(t *testing.T) {
	addr, done := startFakeSMTPServerWith(t, fakeSMTPResponse{rejectRCPT: true})

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial fake server: %v", err)
	}

	smtpClient, err := smtpNewClient(conn, "127.0.0.1")
	if err != nil {
		conn.Close()
		t.Fatalf("smtp.NewClient: %v", err)
	}

	c := clientForFakeServer(t, addr)
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	sendErr := c.sendMessage(smtpClient, []string{"to@example.com"}, msg)

	smtpClient.Quit() //nolint:errcheck
	conn.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if sendErr == nil {
		t.Error("sendMessage should return error when RCPT TO is rejected")
	}
	if !smtpErrWraps(sendErr, ErrSendFailed) {
		t.Errorf("sendMessage RCPT TO error should wrap ErrSendFailed, got: %v", sendErr)
	}
}

// ---------------------------------------------------------------------------
// sendWithPlain — connects through the fake server to reach sendMessage
// ---------------------------------------------------------------------------

func TestSendWithPlain_Success_None(t *testing.T) {
	addr, done := startFakeSMTPServerWith(t, fakeSMTPResponse{})

	client := clientForFakeServer(t, addr)
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err := client.sendWithPlain(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if err != nil {
		t.Errorf("sendWithPlain(NONE) to fake server returned unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// testPlainConnection via fake SMTP server (post-connection path)
// ---------------------------------------------------------------------------

func TestTestPlainConnection_ConnectsSuccessfully(t *testing.T) {
	// Start a fake SMTP server. testPlainConnection is called by TestConnection
	// which is only accessible via the Client.TestConnection() method.
	addr, done := startFakeSMTPServerWith(t, fakeSMTPResponse{})

	client := clientForFakeServer(t, addr)
	// SecurityNone routes to testPlainConnection → no STARTTLS attempted.
	// The fake server handles EHLO and QUIT; testPlainConnection has no credentials
	// so it won't try AUTH — it should succeed.
	err := client.TestConnection()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if err != nil {
		t.Errorf("TestConnection() to fake SMTP server returned unexpected error: %v", err)
	}
}

func TestTestPlainConnection_STARTTLS_UpgradeFails(t *testing.T) {
	// Fake server supports plain SMTP but NOT STARTTLS. The client will connect
	// successfully, then fail when it tries StartTLS (server doesn't advertise it).
	addr, _ := startFakeSMTPServerWith(t, fakeSMTPResponse{})

	// Build a client with STARTTLS security so testPlainConnection tries StartTLS.
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck
	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        host,
		Port:        port,
		FromAddress: "from@example.com",
		Security:    string(SecuritySTARTTLS),
		RetryDelay:  0,
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// This should reach the StartTLS call and fail (server doesn't support it).
	err = client.TestConnection()
	if err == nil {
		t.Error("TestConnection(STARTTLS) to non-TLS server should return error")
	}
}

// ---------------------------------------------------------------------------
// sendMessage — DATA write failure path
// ---------------------------------------------------------------------------

// startFakeSMTPServerRejectDATA starts a server that rejects the DATA command.
func startFakeSMTPServerRejectDATA(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startFakeSMTPServerRejectDATA: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(s string) {
			fmt.Fprintf(rw, "%s\r\n", s)
			rw.Flush() //nolint:errcheck
		}

		writeLine("220 localhost ESMTP")
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-localhost")
				writeLine("250 OK")
			case strings.HasPrefix(upper, "MAIL FROM"):
				writeLine("250 OK")
			case strings.HasPrefix(upper, "RCPT TO"):
				writeLine("250 OK")
			case upper == "DATA":
				writeLine("452 Too many messages")
			case strings.HasPrefix(upper, "QUIT"):
				writeLine("221 Bye")
				return
			default:
				writeLine("500 Command unrecognised")
			}
		}
	}()
	return ln.Addr().String(), doneCh
}

// startFakeSMTPServerRejectAfterBody starts a server that sends "354 Start input"
// but then sends an error after the "." terminator (rejecting the message body).
func startFakeSMTPServerRejectAfterBody(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startFakeSMTPServerRejectAfterBody: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(s string) {
			fmt.Fprintf(rw, "%s\r\n", s)
			rw.Flush() //nolint:errcheck
		}

		writeLine("220 localhost ESMTP")
		inData := false
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if inData {
				if line == "." {
					writeLine("552 Message rejected — content policy violation")
					inData = false
				}
				continue
			}
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-localhost")
				writeLine("250 OK")
			case strings.HasPrefix(upper, "MAIL FROM"):
				writeLine("250 OK")
			case strings.HasPrefix(upper, "RCPT TO"):
				writeLine("250 OK")
			case upper == "DATA":
				writeLine("354 Start input, end with <CRLF>.<CRLF>")
				inData = true
			case strings.HasPrefix(upper, "QUIT"):
				writeLine("221 Bye")
				return
			case strings.HasPrefix(upper, "RSET"):
				writeLine("250 OK")
			default:
				writeLine("500 Command unrecognised")
			}
		}
	}()
	return ln.Addr().String(), doneCh
}

func TestSendMessage_DataCommandRejected(t *testing.T) {
	addr, done := startFakeSMTPServerRejectDATA(t)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial fake server: %v", err)
	}

	smtpClient, err := smtpNewClient(conn, "127.0.0.1")
	if err != nil {
		conn.Close()
		t.Fatalf("smtp.NewClient: %v", err)
	}

	c := clientForFakeServer(t, addr)
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	sendErr := c.sendMessage(smtpClient, []string{"to@example.com"}, msg)

	smtpClient.Quit() //nolint:errcheck
	conn.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if sendErr == nil {
		t.Error("sendMessage should return error when DATA command is rejected")
	}
	if !smtpErrWraps(sendErr, ErrSendFailed) {
		t.Errorf("sendMessage DATA rejection should wrap ErrSendFailed, got: %v", sendErr)
	}
}

func TestSendMessage_BodyRejectedAtClose(t *testing.T) {
	addr, done := startFakeSMTPServerRejectAfterBody(t)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial fake server: %v", err)
	}

	smtpClient, err := smtpNewClient(conn, "127.0.0.1")
	if err != nil {
		conn.Close()
		t.Fatalf("smtp.NewClient: %v", err)
	}

	c := clientForFakeServer(t, addr)
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	sendErr := c.sendMessage(smtpClient, []string{"to@example.com"}, msg)

	smtpClient.Quit() //nolint:errcheck
	conn.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if sendErr == nil {
		t.Error("sendMessage should return error when body is rejected at close")
	}
	if !smtpErrWraps(sendErr, ErrSendFailed) {
		t.Errorf("sendMessage body rejection should wrap ErrSendFailed, got: %v", sendErr)
	}
}

// startFakeSMTPServerDropAfterDATA starts a server that accepts the DATA command
// but then immediately closes the connection, causing the client's write to fail.
func startFakeSMTPServerDropAfterDATA(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startFakeSMTPServerDropAfterDATA: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(s string) {
			fmt.Fprintf(rw, "%s\r\n", s)
			rw.Flush() //nolint:errcheck
		}

		writeLine("220 localhost ESMTP")
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			switch {
			case strings.HasPrefix(strings.ToUpper(line), "EHLO"),
				strings.HasPrefix(strings.ToUpper(line), "HELO"):
				writeLine("250-localhost")
				writeLine("250 OK")
			case strings.HasPrefix(strings.ToUpper(line), "MAIL FROM"):
				writeLine("250 OK")
			case strings.HasPrefix(strings.ToUpper(line), "RCPT TO"):
				writeLine("250 OK")
			case strings.ToUpper(line) == "DATA":
				writeLine("354 Start input")
				// Drop the connection immediately after sending DATA response.
				return
			case strings.HasPrefix(strings.ToUpper(line), "QUIT"):
				writeLine("221 Bye")
				return
			}
		}
	}()
	return ln.Addr().String(), doneCh
}

func TestSendMessage_DataWriteError(t *testing.T) {
	addr, done := startFakeSMTPServerDropAfterDATA(t)

	client := clientForFakeServer(t, addr)
	err := client.Send([]string{"to@example.com"}, "Test Subject", "Hello World", false)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if err == nil {
		t.Error("Send() should return error when server drops connection after DATA")
	}
}

// ---------------------------------------------------------------------------
// SendWithRetry — auth failure causes immediate return (no retry)
// ---------------------------------------------------------------------------

// startFakeSMTPServerRejectAuth starts a fake server that accepts EHLO/MAIL/RCPT/DATA
// normally but advertises AUTH and then rejects AUTH LOGIN/PLAIN credentials.
func startFakeSMTPServerRejectAuth(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startFakeSMTPServerRejectAuth: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(s string) {
			fmt.Fprintf(rw, "%s\r\n", s)
			rw.Flush() //nolint:errcheck
		}

		writeLine("220 localhost ESMTP")
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-localhost")
				writeLine("250-AUTH PLAIN LOGIN")
				writeLine("250 OK")
			case strings.HasPrefix(upper, "AUTH"):
				// Reject all auth attempts
				writeLine("535 Authentication credentials invalid")
			case strings.HasPrefix(upper, "QUIT"):
				writeLine("221 Bye")
				return
			default:
				writeLine("500 Command unrecognised")
			}
		}
	}()
	return ln.Addr().String(), doneCh
}

func TestSendWithRetry_AuthFailed_NoRetry(t *testing.T) {
	// A fake server that rejects AUTH. sendWithPlain will get ErrAuthFailed.
	// SendWithRetry must return immediately (no further retries) on auth failure.
	addr, done := startFakeSMTPServerRejectAuth(t)

	host, portStr, _ := net.SplitHostPort(addr) //nolint:errcheck
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        host,
		Port:        port,
		FromAddress: "from@example.com",
		Security:    string(SecurityNone),
		User:        "user@example.com",
		Password:    "wrongpassword",
		RetryDelay:  0,
	}
	client := &Client{config: cfg}

	err := client.SendWithRetry([]string{"to@example.com"}, "Sub", "body", false, 3)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if err == nil {
		t.Error("SendWithRetry with auth failure should return error")
	}
	if !smtpErrWraps(err, ErrAuthFailed) {
		t.Errorf("SendWithRetry auth failure should wrap ErrAuthFailed, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// sendWithPlain — credentials path (User set, auth attempted)
// ---------------------------------------------------------------------------

// startFakeSMTPServerAcceptAuth starts a fake server that advertises AUTH PLAIN
// and accepts any credentials, completing the full SMTP exchange.
func startFakeSMTPServerAcceptAuth(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startFakeSMTPServerAcceptAuth: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(s string) {
			fmt.Fprintf(rw, "%s\r\n", s)
			rw.Flush() //nolint:errcheck
		}

		writeLine("220 localhost ESMTP")
		inData := false
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")

			if inData {
				if line == "." {
					writeLine("250 OK: queued")
					inData = false
				}
				continue
			}

			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-localhost")
				writeLine("250-AUTH PLAIN LOGIN")
				writeLine("250 OK")
			case strings.HasPrefix(upper, "AUTH PLAIN"):
				// AUTH PLAIN sends credentials inline: "AUTH PLAIN <base64>"
				writeLine("235 Authentication successful")
			case strings.HasPrefix(upper, "AUTH"):
				// Other AUTH methods: send challenge, accept response
				writeLine("334 ")
			case strings.HasPrefix(upper, "MAIL FROM"):
				writeLine("250 OK")
			case strings.HasPrefix(upper, "RCPT TO"):
				writeLine("250 OK")
			case upper == "DATA":
				writeLine("354 Start input, end with <CRLF>.<CRLF>")
				inData = true
			case strings.HasPrefix(upper, "QUIT"):
				writeLine("221 Bye")
				return
			case strings.HasPrefix(upper, "RSET"):
				writeLine("250 OK")
			default:
				// Catch-all for continuation lines (base64 auth)
				writeLine("235 Authentication successful")
			}
		}
	}()
	return ln.Addr().String(), doneCh
}

func TestSendWithPlain_WithCredentials_Success(t *testing.T) {
	addr, done := startFakeSMTPServerAcceptAuth(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        host,
		Port:        port,
		FromAddress: "from@example.com",
		Security:    string(SecurityNone),
		User:        "user@example.com",
		Password:    "secret",
		RetryDelay:  0,
	}
	client := &Client{config: cfg}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithPlain(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if err != nil {
		t.Errorf("sendWithPlain with credentials returned unexpected error: %v", err)
	}
}

func TestTestPlainConnection_WithCredentials_AuthAccepted(t *testing.T) {
	addr, done := startFakeSMTPServerAcceptAuth(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        host,
		Port:        port,
		FromAddress: "from@example.com",
		Security:    string(SecurityNone),
		User:        "user@example.com",
		Password:    "secret",
		RetryDelay:  0,
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = client.TestConnection()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if err != nil {
		t.Errorf("TestConnection() with accepted credentials returned unexpected error: %v", err)
	}
}

func TestTestPlainConnection_WithCredentials_AuthRejected(t *testing.T) {
	addr, done := startFakeSMTPServerRejectAuth(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        host,
		Port:        port,
		FromAddress: "from@example.com",
		Security:    string(SecurityNone),
		User:        "user@example.com",
		Password:    "wrongpass",
		RetryDelay:  0,
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = client.TestConnection()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if err == nil {
		t.Error("TestConnection() with rejected credentials should return error")
	}
	if !smtpErrWraps(err, ErrAuthFailed) {
		t.Errorf("TestConnection auth rejection should wrap ErrAuthFailed, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TLS fake server — self-signed cert for testTLSConnection / sendWithTLS /
// sendWithPlain (STARTTLS) success paths. Tests inject InsecureSkipVerify via
// the Client.tlsConfig override added for testability.
// ---------------------------------------------------------------------------

// generateSelfSignedCert creates a throwaway ECDSA-P256 self-signed cert.
func generateSelfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: CreateCertificate: %v", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: MarshalECPrivateKey: %v", err)
	}
	cert, err := tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}),
	)
	if err != nil {
		t.Fatalf("generateSelfSignedCert: X509KeyPair: %v", err)
	}
	return cert
}

// startFakeTLSSMTPServer starts an SSL/TLS (port-465-style) SMTP server backed
// by a self-signed cert. One connection is served; doneCh closes when done.
func startFakeTLSSMTPServer(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	cert := generateSelfSignedCert(t)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatalf("startFakeTLSSMTPServer: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		serveConnFromGreeting(conn)
	}()
	return ln.Addr().String(), doneCh
}

// serveConnFromGreeting handles an already-connected (or already-upgraded) net.Conn.
// It sends the 220 greeting and then handles all SMTP commands.
func serveConnFromGreeting(conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	writeLine := func(s string) {
		fmt.Fprintf(rw, "%s\r\n", s)
		rw.Flush() //nolint:errcheck
	}
	writeLine("220 localhost ESMTP")
	serveConnCommands(rw, writeLine)
}

// serveConnAfterTLS handles an already-upgraded TLS net.Conn — no greeting sent,
// because net/smtp's StartTLS already consumed the "220 Ready" and then re-sends
// EHLO which we must answer here.
func serveConnAfterTLS(conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	writeLine := func(s string) {
		fmt.Fprintf(rw, "%s\r\n", s)
		rw.Flush() //nolint:errcheck
	}
	serveConnCommands(rw, writeLine)
}

func serveConnCommands(rw *bufio.ReadWriter, writeLine func(string)) {
	inData := false
	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if inData {
			if line == "." {
				writeLine("250 OK: queued")
				inData = false
			}
			continue
		}
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			writeLine("250-localhost")
			writeLine("250-AUTH PLAIN LOGIN")
			writeLine("250 OK")
		case strings.HasPrefix(upper, "AUTH PLAIN"):
			writeLine("235 Authentication successful")
		case strings.HasPrefix(upper, "AUTH"):
			writeLine("334 ")
		case strings.HasPrefix(upper, "MAIL FROM"):
			writeLine("250 OK")
		case strings.HasPrefix(upper, "RCPT TO"):
			writeLine("250 OK")
		case upper == "DATA":
			writeLine("354 Start input, end with <CRLF>.<CRLF>")
			inData = true
		case strings.HasPrefix(upper, "QUIT"):
			writeLine("221 Bye")
			return
		case strings.HasPrefix(upper, "RSET"):
			writeLine("250 OK")
		default:
			writeLine("235 Authentication successful")
		}
	}
}

// startFakeSTARTTLSSMTPServer starts a plain TCP server that advertises STARTTLS
// and upgrades the connection when the client sends STARTTLS.
func startFakeSTARTTLSSMTPServer(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	cert := generateSelfSignedCert(t)
	serverTLSCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startFakeSTARTTLSSMTPServer: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(s string) {
			fmt.Fprintf(rw, "%s\r\n", s)
			rw.Flush() //nolint:errcheck
		}

		writeLine("220 localhost ESMTP")
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-localhost")
				writeLine("250-STARTTLS")
				writeLine("250-AUTH PLAIN LOGIN")
				writeLine("250 OK")
			case upper == "STARTTLS":
				writeLine("220 Ready to start TLS")
				// Upgrade the connection to TLS (flush already done by writeLine)
				tlsConn := tls.Server(conn, serverTLSCfg)
				if err := tlsConn.Handshake(); err != nil {
					return
				}
				// After TLS handshake, net/smtp will re-send EHLO — serve without greeting
				serveConnAfterTLS(tlsConn)
				return
			case strings.HasPrefix(upper, "QUIT"):
				writeLine("221 Bye")
				return
			default:
				writeLine("500 Command unrecognised")
			}
		}
	}()
	return ln.Addr().String(), doneCh
}

// insecureTLSConfig returns a TLS config with InsecureSkipVerify for tests.
func insecureTLSConfig(host string) *tls.Config {
	return &tls.Config{
		ServerName:         host,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, //nolint:gosec // test only — never in production
	}
}

// ---------------------------------------------------------------------------
// testTLSConnection — success path via fake TLS server
// ---------------------------------------------------------------------------

func TestTestTLSConnection_FakeTLSServer_Succeeds(t *testing.T) {
	addr, done := startFakeTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
		},
		tlsConfig: insecureTLSConfig(host),
	}
	err = client.testTLSConnection(addr)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake TLS server did not complete in time")
	}

	if err != nil {
		t.Errorf("testTLSConnection() with fake TLS server returned unexpected error: %v", err)
	}
}

func TestTestTLSConnection_FakeTLSServer_WithCredentials(t *testing.T) {
	addr, done := startFakeTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
			User:        "user@example.com",
			Password:    "secret",
		},
		tlsConfig: insecureTLSConfig(host),
	}
	err = client.testTLSConnection(addr)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake TLS server did not complete in time")
	}

	if err != nil {
		t.Errorf("testTLSConnection() with credentials returned unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// sendWithTLS — success path via fake TLS server
// ---------------------------------------------------------------------------

func TestSendWithTLS_FakeTLSServer_Succeeds(t *testing.T) {
	addr, done := startFakeTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
		},
		tlsConfig: insecureTLSConfig(host),
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithTLS(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake TLS server did not complete in time")
	}

	if err != nil {
		t.Errorf("sendWithTLS() with fake TLS server returned unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// sendWithPlain STARTTLS — success path via fake STARTTLS server
// ---------------------------------------------------------------------------

func TestSendWithPlain_STARTTLS_FakeServer_Succeeds(t *testing.T) {
	addr, done := startFakeSTARTTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySTARTTLS),
		},
		tlsConfig: insecureTLSConfig(host),
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithPlain(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake STARTTLS server did not complete in time")
	}

	if err != nil {
		t.Errorf("sendWithPlain(STARTTLS) returned unexpected error: %v", err)
	}
}

func TestTestPlainConnection_STARTTLS_FakeServer_Succeeds(t *testing.T) {
	addr, done := startFakeSTARTTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySTARTTLS),
		},
		tlsConfig: insecureTLSConfig(host),
	}
	err = client.testPlainConnection(addr)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake STARTTLS server did not complete in time")
	}

	if err != nil {
		t.Errorf("testPlainConnection(STARTTLS) returned unexpected error: %v", err)
	}
}

func TestSendWithTLS_FakeTLSServer_WithCredentials(t *testing.T) {
	addr, done := startFakeTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
			User:        "user@example.com",
			Password:    "secret",
		},
		tlsConfig: insecureTLSConfig(host),
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithTLS(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake TLS server did not complete in time")
	}

	if err != nil {
		t.Errorf("sendWithTLS() with credentials returned unexpected error: %v", err)
	}
}

func TestSendWithPlain_STARTTLS_WithCredentials(t *testing.T) {
	addr, done := startFakeSTARTTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySTARTTLS),
			User:        "user@example.com",
			Password:    "secret",
		},
		tlsConfig: insecureTLSConfig(host),
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithPlain(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake STARTTLS server did not complete in time")
	}

	if err != nil {
		t.Errorf("sendWithPlain(STARTTLS) with credentials returned unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// sendWithPlain STARTTLS — nil tlsConfig branch (uses default TLS config which
// will reject the self-signed cert; this exercises lines 272-276 in smtp.go)
// ---------------------------------------------------------------------------

func TestSendWithPlain_STARTTLS_DefaultTLSConfig_CertFails(t *testing.T) {
	// Start a STARTTLS server. The client will use the default TLS config
	// (no InsecureSkipVerify), so the cert verification will fail.
	// The point is to exercise the `if startTLSConfig == nil` branch in smtp.go.
	addr, done := startFakeSTARTTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	// No tlsConfig override → uses the default (InsecureSkipVerify=false)
	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySTARTTLS),
		},
		// tlsConfig intentionally nil — exercises the default-config branch
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithPlain(addr, []string{"to@example.com"}, msg)

	// Server may or may not complete depending on how TLS failure propagates;
	// drain the channel with a short timeout.
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		// Server may block waiting for more data; that's fine
	}

	// STARTTLS should fail because the self-signed cert is not trusted.
	if err == nil {
		t.Error("sendWithPlain(STARTTLS) with default TLS config against self-signed cert should return error")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("sendWithPlain STARTTLS default config failure should wrap ErrSendFailed, got: %v", err)
	}
}

func TestTestPlainConnection_STARTTLS_DefaultTLSConfig_CertFails(t *testing.T) {
	addr, done := startFakeSTARTTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySTARTTLS),
		},
	}
	err = client.testPlainConnection(addr)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	if err == nil {
		t.Error("testPlainConnection(STARTTLS) with default TLS config should fail on self-signed cert")
	}
}

func TestTestTLSConnection_DefaultTLSConfig_CertFails(t *testing.T) {
	addr, done := startFakeTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
		},
		// nil tlsConfig — uses default (no InsecureSkipVerify)
	}
	err = client.testTLSConnection(addr)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	if err == nil {
		t.Error("testTLSConnection with default TLS config should fail on self-signed cert")
	}
	if !smtpErrWraps(err, ErrConnectionFailed) {
		t.Errorf("testTLSConnection cert failure should wrap ErrConnectionFailed, got: %v", err)
	}
}

func TestSendWithTLS_DefaultTLSConfig_CertFails(t *testing.T) {
	addr, done := startFakeTLSSMTPServer(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
		},
		// nil tlsConfig — uses default (no InsecureSkipVerify)
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithTLS(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	if err == nil {
		t.Error("sendWithTLS with default TLS config should fail on self-signed cert")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("sendWithTLS cert failure should wrap ErrSendFailed, got: %v", err)
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

// ---------------------------------------------------------------------------
// TLS auth rejection — covers lines 144 (testTLSConnection) and 244 (sendWithTLS)
// ---------------------------------------------------------------------------

// startFakeTLSSMTPServerRejectAuth starts a TLS SMTP server backed by a
// self-signed cert that advertises AUTH PLAIN but rejects every AUTH attempt.
func startFakeTLSSMTPServerRejectAuth(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	cert := generateSelfSignedCert(t)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatalf("startFakeTLSSMTPServerRejectAuth: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(s string) {
			fmt.Fprintf(rw, "%s\r\n", s)
			rw.Flush() //nolint:errcheck
		}

		writeLine("220 localhost ESMTP")
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-localhost")
				writeLine("250-AUTH PLAIN LOGIN")
				writeLine("250 OK")
			case strings.HasPrefix(upper, "AUTH"):
				writeLine("535 Authentication credentials invalid")
			case strings.HasPrefix(upper, "QUIT"):
				writeLine("221 Bye")
				return
			default:
				writeLine("500 Command unrecognised")
			}
		}
	}()
	return ln.Addr().String(), doneCh
}

// startFakeTLSSMTPServerBadGreeting starts a TLS server that sends a non-220
// response instead of the SMTP banner, causing smtp.NewClient to fail.
func startFakeTLSSMTPServerBadGreeting(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	cert := generateSelfSignedCert(t)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatalf("startFakeTLSSMTPServerBadGreeting: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// Send a non-SMTP response; smtp.NewClient expects "220 ..." banner
		fmt.Fprintf(conn, "500 Not an SMTP server\r\n") //nolint:errcheck
	}()
	return ln.Addr().String(), doneCh
}

// startFakePlainSMTPServerBadGreeting starts a plain TCP server that sends a
// non-220 response, causing smtp.NewClient to fail after the TCP dial.
func startFakePlainSMTPServerBadGreeting(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startFakePlainSMTPServerBadGreeting: listen: %v", err)
	}
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		fmt.Fprintf(conn, "500 Not an SMTP server\r\n") //nolint:errcheck
	}()
	return ln.Addr().String(), doneCh
}

func TestTestTLSConnection_BadGreeting_NewClientFails(t *testing.T) {
	addr, done := startFakeTLSSMTPServerBadGreeting(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
		},
		tlsConfig: insecureTLSConfig(host),
	}
	err = client.testTLSConnection(addr)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake TLS server did not complete in time")
	}

	if err == nil {
		t.Error("testTLSConnection should return error when server sends bad greeting")
	}
	if !smtpErrWraps(err, ErrConnectionFailed) {
		t.Errorf("testTLSConnection bad greeting should wrap ErrConnectionFailed, got: %v", err)
	}
}

func TestSendWithTLS_BadGreeting_NewClientFails(t *testing.T) {
	addr, done := startFakeTLSSMTPServerBadGreeting(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
		},
		tlsConfig: insecureTLSConfig(host),
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithTLS(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake TLS server did not complete in time")
	}

	if err == nil {
		t.Error("sendWithTLS should return error when server sends bad greeting")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("sendWithTLS bad greeting should wrap ErrSendFailed, got: %v", err)
	}
}

func TestTestPlainConnection_BadGreeting_NewClientFails(t *testing.T) {
	addr, done := startFakePlainSMTPServerBadGreeting(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecurityNone),
		},
	}
	err = client.testPlainConnection(addr)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake plain server did not complete in time")
	}

	if err == nil {
		t.Error("testPlainConnection should return error when server sends bad greeting")
	}
	if !smtpErrWraps(err, ErrConnectionFailed) {
		t.Errorf("testPlainConnection bad greeting should wrap ErrConnectionFailed, got: %v", err)
	}
}

func TestSendWithPlain_BadGreeting_NewClientFails(t *testing.T) {
	addr, done := startFakePlainSMTPServerBadGreeting(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecurityNone),
		},
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithPlain(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake plain server did not complete in time")
	}

	if err == nil {
		t.Error("sendWithPlain should return error when server sends bad greeting")
	}
	if !smtpErrWraps(err, ErrSendFailed) {
		t.Errorf("sendWithPlain bad greeting should wrap ErrSendFailed, got: %v", err)
	}
}

func TestTestTLSConnection_FakeTLSServer_AuthRejected(t *testing.T) {
	addr, done := startFakeTLSSMTPServerRejectAuth(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
			User:        "user@example.com",
			Password:    "wrongpassword",
		},
		tlsConfig: insecureTLSConfig(host),
	}
	err = client.testTLSConnection(addr)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake TLS server did not complete in time")
	}

	if err == nil {
		t.Error("testTLSConnection should return error when auth is rejected")
	}
	if !smtpErrWraps(err, ErrAuthFailed) {
		t.Errorf("testTLSConnection auth rejection should wrap ErrAuthFailed, got: %v", err)
	}
}

func TestSendWithTLS_FakeTLSServer_AuthRejected(t *testing.T) {
	addr, done := startFakeTLSSMTPServerRejectAuth(t)

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port) //nolint:errcheck

	client := &Client{
		config: &models.SMTPConfig{
			Enabled:     true,
			Host:        host,
			Port:        port,
			FromAddress: "from@example.com",
			Security:    string(SecuritySSLTLS),
			User:        "user@example.com",
			Password:    "wrongpassword",
		},
		tlsConfig: insecureTLSConfig(host),
	}
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	err = client.sendWithTLS(addr, []string{"to@example.com"}, msg)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake TLS server did not complete in time")
	}

	if err == nil {
		t.Error("sendWithTLS should return error when auth is rejected")
	}
	if !smtpErrWraps(err, ErrAuthFailed) {
		t.Errorf("sendWithTLS auth rejection should wrap ErrAuthFailed, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// sendMessage — DATA command rejection path
// ---------------------------------------------------------------------------

// TestSendMessage_WriteError triggers the w.Write() error branch inside sendMessage by
// having the fake server close the connection immediately after accepting DATA, so
// the subsequent write to the data writer encounters a broken-pipe error.
func TestSendMessage_WriteError(t *testing.T) {
	addr, done := startFakeSMTPServerWith(t, fakeSMTPResponse{closeAfterDATAAccept: true})

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial fake server: %v", err)
	}

	smtpClient, err := smtpNewClient(conn, "127.0.0.1")
	if err != nil {
		conn.Close()
		t.Fatalf("smtp.NewClient: %v", err)
	}

	c := clientForFakeServer(t, addr)
	// Build a large message so the buffered write is more likely to flush to
	// the (now-closed) connection before the smtp layer notices.
	bigBody := strings.Repeat("x", 65536)
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", bigBody, false)
	sendErr := c.sendMessage(smtpClient, []string{"to@example.com"}, msg)

	smtpClient.Quit() //nolint:errcheck
	conn.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	// The write or close should fail with ErrSendFailed.
	if sendErr == nil {
		t.Error("sendMessage should return error when server closes connection during DATA write")
	}
	if !smtpErrWraps(sendErr, ErrSendFailed) {
		t.Errorf("sendMessage write error should wrap ErrSendFailed, got: %v", sendErr)
	}
}

// TestSendMessage_DataCommandError triggers the client.Data() error branch inside
// sendMessage by having the fake server reject the DATA command with a 554 response.
func TestSendMessage_DataCommandError(t *testing.T) {
	addr, done := startFakeSMTPServerWith(t, fakeSMTPResponse{rejectDATA: true})

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial fake server: %v", err)
	}

	smtpClient, err := smtpNewClient(conn, "127.0.0.1")
	if err != nil {
		conn.Close()
		t.Fatalf("smtp.NewClient: %v", err)
	}

	c := clientForFakeServer(t, addr)
	msg := buildMessage("", "from@example.com", []string{"to@example.com"}, "Sub", "body", false)
	sendErr := c.sendMessage(smtpClient, []string{"to@example.com"}, msg)

	smtpClient.Quit() //nolint:errcheck
	conn.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("fake SMTP server did not complete in time")
	}

	if sendErr == nil {
		t.Error("sendMessage should return error when DATA command is rejected")
	}
	if !smtpErrWraps(sendErr, ErrSendFailed) {
		t.Errorf("sendMessage DATA error should wrap ErrSendFailed, got: %v", sendErr)
	}
}
