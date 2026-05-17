package service

import (
	"strings"
	"testing"
)

func TestNewTemplateData(t *testing.T) {
	td := NewTemplateData("MySite", "https://example.com", "Alice")
	if td == nil {
		t.Fatal("NewTemplateData() returned nil")
	}
	if td.SiteName != "MySite" {
		t.Errorf("SiteName = %q, want MySite", td.SiteName)
	}
	if td.SiteURL != "https://example.com" {
		t.Errorf("SiteURL = %q, want https://example.com", td.SiteURL)
	}
	if td.RecipientName != "Alice" {
		t.Errorf("RecipientName = %q, want Alice", td.RecipientName)
	}
	if td.Year == 0 {
		t.Error("Year should not be 0")
	}
	if td.Data == nil {
		t.Error("Data map should not be nil")
	}
}

func TestWelcomeEmail(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Bob")
	url := "https://test.com/verify?token=abc123"
	email := WelcomeEmail(td, url)

	if email == nil {
		t.Fatal("WelcomeEmail() returned nil")
	}
	if !strings.Contains(email.Subject, "Welcome") {
		t.Errorf("Subject = %q, should contain 'Welcome'", email.Subject)
	}
	if !strings.Contains(email.Subject, "TestSite") {
		t.Errorf("Subject = %q, should contain 'TestSite'", email.Subject)
	}
	if !strings.Contains(email.HTMLBody, url) {
		t.Error("HTMLBody should contain verification URL")
	}
	if !strings.Contains(email.TextBody, url) {
		t.Error("TextBody should contain verification URL")
	}
	if !strings.Contains(email.HTMLBody, "Bob") {
		t.Error("HTMLBody should contain recipient name")
	}
	if email.Preheader == "" {
		t.Error("Preheader should not be empty")
	}
}

func TestPasswordResetEmail(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Carol")
	url := "https://test.com/reset?token=xyz789"
	email := PasswordResetEmail(td, url)

	if email == nil {
		t.Fatal("PasswordResetEmail() returned nil")
	}
	if !strings.Contains(email.Subject, "Password Reset") {
		t.Errorf("Subject = %q, should contain 'Password Reset'", email.Subject)
	}
	if !strings.Contains(email.HTMLBody, url) {
		t.Error("HTMLBody should contain reset URL")
	}
	if !strings.Contains(email.TextBody, url) {
		t.Error("TextBody should contain reset URL")
	}
}

func TestEmailVerificationEmail(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Dave")
	url := "https://test.com/verify?token=verify123"
	email := EmailVerificationEmail(td, url)

	if email == nil {
		t.Fatal("EmailVerificationEmail() returned nil")
	}
	if !strings.Contains(email.Subject, "Verify") {
		t.Errorf("Subject = %q, should contain 'Verify'", email.Subject)
	}
	if !strings.Contains(email.HTMLBody, url) {
		t.Error("HTMLBody should contain verification URL")
	}
}

func TestTwoFactorCodeEmail(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Eve")
	code := "123456"
	email := TwoFactorCodeEmail(td, code)

	if email == nil {
		t.Fatal("TwoFactorCodeEmail() returned nil")
	}
	if !strings.Contains(email.Subject, "Two-Factor") {
		t.Errorf("Subject = %q, should contain 'Two-Factor'", email.Subject)
	}
	if !strings.Contains(email.HTMLBody, code) {
		t.Error("HTMLBody should contain the 2FA code")
	}
	if !strings.Contains(email.TextBody, code) {
		t.Error("TextBody should contain the 2FA code")
	}
	if !strings.Contains(email.Preheader, code) {
		t.Errorf("Preheader = %q, should contain code %q", email.Preheader, code)
	}
}

func TestNotificationEmail_Info(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Frank")
	email := NotificationEmail(td, "Test Alert", "Something happened", "info")

	if email == nil {
		t.Fatal("NotificationEmail() returned nil")
	}
	if !strings.Contains(email.Subject, "Test Alert") {
		t.Errorf("Subject = %q, should contain 'Test Alert'", email.Subject)
	}
	if strings.HasPrefix(email.Subject, "[URGENT]") {
		t.Error("info notification should not have [URGENT] prefix")
	}
}

func TestNotificationEmail_Critical(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Grace")
	email := NotificationEmail(td, "Critical Error", "System failure", "critical")

	if !strings.HasPrefix(email.Subject, "[URGENT]") {
		t.Errorf("Subject = %q, should start with [URGENT]", email.Subject)
	}
	if !strings.Contains(email.HTMLBody, "danger-box") {
		t.Error("critical notification should use danger-box CSS class")
	}
}

func TestNotificationEmail_Warning(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Hank")
	email := NotificationEmail(td, "Warning", "Disk almost full", "warning")

	if !strings.HasPrefix(email.Subject, "[WARNING]") {
		t.Errorf("Subject = %q, should start with [WARNING]", email.Subject)
	}
	if !strings.Contains(email.HTMLBody, "warning-box") {
		t.Error("warning notification should use warning-box CSS class")
	}
}

func TestNotificationEmail_Emergency(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Ivy")
	email := NotificationEmail(td, "Emergency", "Data loss", "emergency")

	if !strings.HasPrefix(email.Subject, "[URGENT]") {
		t.Errorf("Subject = %q, should start with [URGENT] for emergency", email.Subject)
	}
}

func TestTeamInviteEmail(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Jack")
	inviteURL := "https://test.com/invite?token=inv123"
	email := TeamInviteEmail(td, "MyOrg", "admin", inviteURL)

	if email == nil {
		t.Fatal("TeamInviteEmail() returned nil")
	}
	if !strings.Contains(email.Subject, "MyOrg") {
		t.Errorf("Subject = %q, should contain org name", email.Subject)
	}
	if !strings.Contains(email.HTMLBody, inviteURL) {
		t.Error("HTMLBody should contain invite URL")
	}
	if !strings.Contains(email.HTMLBody, "admin") {
		t.Error("HTMLBody should contain role")
	}
	if !strings.Contains(email.Preheader, "MyOrg") {
		t.Errorf("Preheader = %q, should contain org name", email.Preheader)
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<p>Hello</p>", "Hello\n"},
		{"<b>Bold</b>", "Bold"},
		{"Hello <br> World", "Hello \n World"},
		{"Hello <br/> World", "Hello \n World"},
		{"Hello <br /> World", "Hello \n World"},
		{"Plain text", "Plain text"},
		{"<div><p>Nested</p></div>", "Nested\n"},
	}

	for _, tt := range tests {
		got := stripHTML(tt.input)
		if got != tt.want {
			t.Errorf("stripHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRenderHTML(t *testing.T) {
	td := NewTemplateData("TestSite", "https://test.com", "Kate")
	content := "<p>Hello, World!</p>"
	html := renderHTML(td, content)

	if html == "" {
		t.Error("renderHTML() returned empty string")
	}
	if !strings.Contains(html, "TestSite") {
		t.Error("renderHTML() should contain site name")
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("renderHTML() should contain DOCTYPE")
	}
}
