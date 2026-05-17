package service

import (
	"strings"
	"testing"

	models "github.com/casapps/cassocial/src/server/model"
)

func TestNewNotificationManager_DefaultPreferences(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@example.com")
	if nm == nil {
		t.Fatal("NewNotificationManager() returned nil")
	}
	if nm.adminEmail != "admin@example.com" {
		t.Errorf("adminEmail = %q, want admin@example.com", nm.adminEmail)
	}
	if nm.preferences == nil {
		t.Error("preferences should not be nil when passed nil")
	}
	if nm.batchDelay == 0 {
		t.Error("batchDelay should not be 0")
	}
}

func TestNewNotificationManager_CustomPreferences(t *testing.T) {
	prefs := &models.NotificationPreferences{
		Emergency:  true,
		BatchDelay: 60,
	}
	nm := NewNotificationManager(nil, prefs, "admin@example.com")
	if nm.preferences != prefs {
		t.Error("should use provided preferences")
	}
}

func TestNotificationManager_GetQueueLength_Empty(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@test.com")
	if l := nm.GetQueueLength(); l != 0 {
		t.Errorf("initial queue length = %d, want 0", l)
	}
}

func TestNotificationManager_IsRunning_Initial(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@test.com")
	if nm.IsRunning() {
		t.Error("new manager should not be running")
	}
}

func TestFormatCertificateMessage(t *testing.T) {
	msg := formatCertificateMessage("example.com", 14)
	if !strings.Contains(msg, "example.com") {
		t.Error("certificate message should contain domain name")
	}
}

func TestFormatBugReportMessage(t *testing.T) {
	msg := formatBugReportMessage("alice", "Login fails on mobile")
	if !strings.Contains(msg, "alice") {
		t.Error("bug report message should contain reporter name")
	}
	if !strings.Contains(msg, "Login fails on mobile") {
		t.Error("bug report message should contain description")
	}
}

func TestFormatUserRegistrationMessage(t *testing.T) {
	msg := formatUserRegistrationMessage("bob", "bob@example.com")
	if !strings.Contains(msg, "bob") {
		t.Error("registration message should contain username")
	}
	if !strings.Contains(msg, "bob@example.com") {
		t.Error("registration message should contain email")
	}
}

func TestFormatDomainVerificationMessage_Verified(t *testing.T) {
	msg := formatDomainVerificationMessage("mysite.com", "verified")
	if !strings.Contains(msg, "mysite.com") {
		t.Error("domain verification message should contain domain")
	}
}

func TestFormatDomainVerificationMessage_Failed(t *testing.T) {
	msg := formatDomainVerificationMessage("mysite.com", "failed")
	if !strings.Contains(msg, "mysite.com") {
		t.Error("failed domain message should contain domain")
	}
}

func TestFormatBackupStatusMessage_Success(t *testing.T) {
	msg := formatBackupStatusMessage("success", "5.2 MB")
	if !strings.Contains(msg, "5.2 MB") {
		t.Error("backup success message should contain details")
	}
}

func TestFormatBackupStatusMessage_Failure(t *testing.T) {
	msg := formatBackupStatusMessage("failed", "disk full")
	if !strings.Contains(msg, "disk full") {
		t.Error("backup failure message should contain error details")
	}
}

func TestFormatHighTrafficMessage(t *testing.T) {
	msg := formatHighTrafficMessage(500, 200)
	if msg == "" {
		t.Error("high traffic message should not be empty")
	}
}

func TestNotificationManager_isNotificationEnabled(t *testing.T) {
	prefs := &models.NotificationPreferences{
		Emergency:          true,
		Certificate:        false,
		BugReport:          true,
		UserRegistration:   false,
		DomainVerification: true,
		BackupStatus:       false,
		HighTraffic:        true,
	}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")

	tests := []struct {
		notifType NotificationType
		want      bool
	}{
		{NotificationEmergency, true},
		{NotificationCertificate, false},
		{NotificationBugReport, true},
		{NotificationUserRegistration, false},
		{NotificationDomainVerification, true},
		{NotificationBackupStatus, false},
		{NotificationHighTraffic, true},
		{"unknown-type", true}, // unknown types default to true
	}

	for _, tt := range tests {
		got := nm.isNotificationEnabled(tt.notifType)
		if got != tt.want {
			t.Errorf("isNotificationEnabled(%q) = %v, want %v", tt.notifType, got, tt.want)
		}
	}
}
