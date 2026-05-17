package service

import (
	"testing"

	models "github.com/casapps/cassocial/src/server/model"
)

// ---- NewMailer ----

func TestNewMailer_NilConfig(t *testing.T) {
	m, err := NewMailer(nil, "TestSite", "https://example.com")
	if err != nil {
		t.Errorf("NewMailer(nil) returned error: %v", err)
	}
	if m == nil {
		t.Fatal("NewMailer(nil) returned nil mailer")
	}
	if m.enabled {
		t.Error("NewMailer(nil) should have enabled=false")
	}
	if m.siteName != "TestSite" {
		t.Errorf("siteName = %q, want TestSite", m.siteName)
	}
	if m.siteURL != "https://example.com" {
		t.Errorf("siteURL = %q, want https://example.com", m.siteURL)
	}
}

func TestNewMailer_InvalidConfig_NilClient(t *testing.T) {
	// An SMTPConfig that fails Validate (empty host) causes NewClient to fail.
	// NewMailer should gracefully degrade and return a disabled mailer.
	cfg := &models.SMTPConfig{
		Enabled:     true,
		Host:        "", // invalid — required
		Port:        587,
		FromAddress: "sender@example.com",
	}
	m, err := NewMailer(cfg, "MySite", "https://mysite.com")
	if err != nil {
		t.Errorf("NewMailer(invalid config) returned error: %v", err)
	}
	if m == nil {
		t.Fatal("NewMailer(invalid config) returned nil")
	}
	// Should be disabled because the client creation failed
	if m.enabled {
		t.Error("NewMailer(invalid config) should be disabled")
	}
}

func TestNewMailer_ValidConfig_DisabledFlag(t *testing.T) {
	// A valid config with Enabled=false results in a disabled mailer.
	cfg := &models.SMTPConfig{
		Enabled:     false,
		Host:        "smtp.example.com",
		Port:        587,
		FromAddress: "sender@example.com",
	}
	m, err := NewMailer(cfg, "MySite", "https://mysite.com")
	if err != nil {
		t.Errorf("NewMailer(disabled) returned error: %v", err)
	}
	if m == nil {
		t.Fatal("NewMailer(disabled) returned nil")
	}
	// enabled comes from config.Enabled; client was created but Enabled=false
	if m.IsEnabled() {
		t.Error("IsEnabled() should be false when config.Enabled=false")
	}
}

// ---- IsEnabled ----

func TestIsEnabled_NilConfig_ReturnsFalse(t *testing.T) {
	m, _ := NewMailer(nil, "Site", "https://site.com")
	if m.IsEnabled() {
		t.Error("IsEnabled() should be false when created with nil config")
	}
}

// ---- SendWelcome (disabled mailer — no network) ----

func TestSendWelcome_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendWelcome("user@example.com", "alice", "https://mysite.com/verify?token=abc")
	if err != nil {
		t.Errorf("SendWelcome(disabled) returned error: %v", err)
	}
}

func TestSendWelcome_Disabled_EmptyTo_NoError(t *testing.T) {
	// Disabled mailer skips all validation — even empty recipient
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendWelcome("", "alice", "https://mysite.com/verify")
	if err != nil {
		t.Errorf("SendWelcome(disabled, empty to) returned unexpected error: %v", err)
	}
}

// ---- SendPasswordReset (disabled mailer) ----

func TestSendPasswordReset_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendPasswordReset("user@example.com", "alice", "https://mysite.com/reset?token=abc")
	if err != nil {
		t.Errorf("SendPasswordReset(disabled) returned error: %v", err)
	}
}

// ---- SendEmailVerification (disabled mailer) ----

func TestSendEmailVerification_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendEmailVerification("user@example.com", "alice", "https://mysite.com/verify?token=xyz")
	if err != nil {
		t.Errorf("SendEmailVerification(disabled) returned error: %v", err)
	}
}

// ---- SendTwoFactorCode (disabled mailer) ----

func TestSendTwoFactorCode_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendTwoFactorCode("user@example.com", "alice", "123456")
	if err != nil {
		t.Errorf("SendTwoFactorCode(disabled) returned error: %v", err)
	}
}

// ---- SendTeamInvite (disabled mailer) ----

func TestSendTeamInvite_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendTeamInvite("user@example.com", "alice", "MyOrg", "member", "https://mysite.com/invite?token=tok")
	if err != nil {
		t.Errorf("SendTeamInvite(disabled) returned error: %v", err)
	}
}

// ---- SendNotification (disabled mailer) ----

func TestSendNotification_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendNotification("admin@example.com", "Admin", "Test", "message body", "info", 1)
	if err != nil {
		t.Errorf("SendNotification(disabled) returned error: %v", err)
	}
}

// ---- SendPlainText (disabled mailer) ----

func TestSendPlainText_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendPlainText([]string{"user@example.com"}, "Test Subject", "Hello World")
	if err != nil {
		t.Errorf("SendPlainText(disabled) returned error: %v", err)
	}
}

func TestSendPlainText_Disabled_EmptyTo_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendPlainText([]string{}, "Subject", "Body")
	if err != nil {
		t.Errorf("SendPlainText(disabled, empty to) returned unexpected error: %v", err)
	}
}

// ---- SendHTML (disabled mailer) ----

func TestSendHTML_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendHTML([]string{"user@example.com"}, "HTML Subject", "<h1>Hello</h1>")
	if err != nil {
		t.Errorf("SendHTML(disabled) returned error: %v", err)
	}
}

// ---- TestConnection ----

func TestTestConnection_NilClient(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.TestConnection()
	if err == nil {
		t.Error("TestConnection() with nil client should return error")
	}
}

// ---- Notification helper methods (disabled mailer) ----

func TestSendUserRegistrationNotification_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendUserRegistrationNotification("admin@example.com", "alice", "alice@example.com")
	if err != nil {
		t.Errorf("SendUserRegistrationNotification(disabled) returned error: %v", err)
	}
}

func TestSendDomainVerificationNotification_Verified_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendDomainVerificationNotification("admin@example.com", "example.com", "verified")
	if err != nil {
		t.Errorf("SendDomainVerificationNotification(verified, disabled) returned error: %v", err)
	}
}

func TestSendDomainVerificationNotification_Failed_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendDomainVerificationNotification("admin@example.com", "example.com", "failed")
	if err != nil {
		t.Errorf("SendDomainVerificationNotification(failed, disabled) returned error: %v", err)
	}
}

func TestSendBackupStatusNotification_Success_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendBackupStatusNotification("admin@example.com", "success", "backup created at /tmp/backup.db")
	if err != nil {
		t.Errorf("SendBackupStatusNotification(success, disabled) returned error: %v", err)
	}
}

func TestSendBackupStatusNotification_Failure_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendBackupStatusNotification("admin@example.com", "failure", "disk full")
	if err != nil {
		t.Errorf("SendBackupStatusNotification(failure, disabled) returned error: %v", err)
	}
}

func TestSendCertificateRenewalNotification_Critical_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendCertificateRenewalNotification("admin@example.com", "example.com", 5)
	if err != nil {
		t.Errorf("SendCertificateRenewalNotification(critical, disabled) returned error: %v", err)
	}
}

func TestSendCertificateRenewalNotification_Normal_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendCertificateRenewalNotification("admin@example.com", "example.com", 20)
	if err != nil {
		t.Errorf("SendCertificateRenewalNotification(normal, disabled) returned error: %v", err)
	}
}

func TestSendEmergencyAlert_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendEmergencyAlert("admin@example.com", "System Down", "All nodes unreachable")
	if err != nil {
		t.Errorf("SendEmergencyAlert(disabled) returned error: %v", err)
	}
}

func TestSendHighTrafficNotification_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendHighTrafficNotification("admin@example.com", 1500, 1000)
	if err != nil {
		t.Errorf("SendHighTrafficNotification(disabled) returned error: %v", err)
	}
}

func TestSendBugReport_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendBugReport("admin@example.com", "bob", "Login fails after 3pm")
	if err != nil {
		t.Errorf("SendBugReport(disabled) returned error: %v", err)
	}
}

func TestSendProfileSuspensionNotice_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendProfileSuspensionNotice("user@example.com", "alice", "spam")
	if err != nil {
		t.Errorf("SendProfileSuspensionNotice(disabled) returned error: %v", err)
	}
}

func TestSendDataExportReady_Disabled_NoError(t *testing.T) {
	m, _ := NewMailer(nil, "MySite", "https://mysite.com")
	err := m.SendDataExportReady("user@example.com", "alice", "https://mysite.com/exports/abc.zip")
	if err != nil {
		t.Errorf("SendDataExportReady(disabled) returned error: %v", err)
	}
}
