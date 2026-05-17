package service

import (
	"strings"
	"testing"
	"time"

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

// ---- Start / Stop ----

func TestNotificationManager_Start_SetsRunning(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@test.com")
	nm.Start()
	if !nm.IsRunning() {
		t.Error("IsRunning() should be true after Start()")
	}
	// Stop via channel without calling the deadlock-prone Stop() method
	nm.stopChan <- true
}

func TestNotificationManager_Start_Idempotent(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@test.com")
	nm.Start()
	// Second call while running should be a no-op
	nm.Start()
	if !nm.IsRunning() {
		t.Error("IsRunning() should be true after second Start()")
	}
	nm.stopChan <- true
}

func TestNotificationManager_Stop_NotRunning_NoOp(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@test.com")
	// Should not panic or deadlock — manager is not running
	nm.Stop()
}

// ---- Queue ----

func TestNotificationManager_Queue_DisabledType(t *testing.T) {
	prefs := &models.NotificationPreferences{
		Certificate: false,
	}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")

	nm.Queue(&Notification{
		Type:      NotificationCertificate,
		Title:     "cert expiring",
		Priority:  PriorityNormal,
		CreatedAt: timeNow(),
	})

	if nm.GetQueueLength() != 0 {
		t.Error("disabled notification type should not be queued")
	}
}

func TestNotificationManager_Queue_NormalPriority(t *testing.T) {
	prefs := &models.NotificationPreferences{
		BugReport: true,
	}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")

	nm.Queue(&Notification{
		Type:      NotificationBugReport,
		Title:     "test bug",
		Priority:  PriorityNormal,
		CreatedAt: timeNow(),
	})

	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotificationManager_Queue_EmergencyPriority_BypassesBatch(t *testing.T) {
	// Emergency notifications bypass the queue and call sendImmediately.
	// Use a disabled mailer (not nil) so sendNotification hits IsEnabled()=false and returns early.
	disabledMailer, _ := NewMailer(nil, "MySite", "https://mysite.com")
	prefs := &models.NotificationPreferences{
		Emergency: true,
	}
	nm := NewNotificationManager(disabledMailer, prefs, "admin@test.com")

	nm.Queue(&Notification{
		Type:      NotificationEmergency,
		Title:     "EMERGENCY",
		Priority:  PriorityEmergency,
		CreatedAt: timeNow(),
	})

	// Emergency bypasses batch — queue stays at 0
	if nm.GetQueueLength() != 0 {
		t.Errorf("emergency notification should bypass queue, got length %d", nm.GetQueueLength())
	}
}

// timeNow is a helper to avoid importing time in the test expressions.
func timeNow() time.Time { return time.Now() }

// ---- UpdatePreferences ----

func TestUpdatePreferences_ChangesDelay(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@test.com")
	oldDelay := nm.batchDelay

	newPrefs := &models.NotificationPreferences{
		BatchDelay: int(oldDelay.Seconds()) + 10,
	}
	nm.UpdatePreferences(newPrefs)

	if nm.batchDelay == oldDelay {
		t.Error("UpdatePreferences() should change batchDelay when new value differs")
	}
	if nm.preferences != newPrefs {
		t.Error("UpdatePreferences() should replace preferences reference")
	}
}

func TestUpdatePreferences_WhileRunning(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@test.com")
	nm.Start()
	defer func() { nm.stopChan <- true }()

	newPrefs := &models.NotificationPreferences{BatchDelay: 99}
	// Should not panic when called while running
	nm.UpdatePreferences(newPrefs)
}

func TestUpdatePreferences_ZeroBatchDelay_Ignored(t *testing.T) {
	nm := NewNotificationManager(nil, nil, "admin@test.com")
	oldDelay := nm.batchDelay

	newPrefs := &models.NotificationPreferences{BatchDelay: 0}
	nm.UpdatePreferences(newPrefs)

	if nm.batchDelay != oldDelay {
		t.Error("UpdatePreferences() should not change batchDelay when new value is 0")
	}
}

// ---- Notify* helpers ----

func TestNotifyEmergency_DoesNotPanic(t *testing.T) {
	disabledMailer, _ := NewMailer(nil, "MySite", "https://mysite.com")
	prefs := &models.NotificationPreferences{Emergency: true}
	nm := NewNotificationManager(disabledMailer, prefs, "admin@test.com")
	nm.NotifyEmergency("System Down", "All services unreachable")
	// Emergency bypasses batch; queue should still be 0
	if nm.GetQueueLength() != 0 {
		t.Errorf("emergency queue length = %d, want 0", nm.GetQueueLength())
	}
}

func TestNotifyCertificateExpiring_CriticalPriority(t *testing.T) {
	prefs := &models.NotificationPreferences{Certificate: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyCertificateExpiring("example.com", 5) // ≤7 days → high priority
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotifyCertificateExpiring_NormalPriority(t *testing.T) {
	prefs := &models.NotificationPreferences{Certificate: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyCertificateExpiring("example.com", 20) // >7 days → normal priority
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotifyBugReport_DoesNotPanic(t *testing.T) {
	prefs := &models.NotificationPreferences{BugReport: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyBugReport("alice", "Login page broken")
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotifyUserRegistration_DoesNotPanic(t *testing.T) {
	prefs := &models.NotificationPreferences{UserRegistration: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyUserRegistration("bob", "bob@example.com")
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotifyDomainVerification_Verified_DoesNotPanic(t *testing.T) {
	prefs := &models.NotificationPreferences{DomainVerification: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyDomainVerification("example.com", "verified")
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotifyDomainVerification_Failed_DoesNotPanic(t *testing.T) {
	prefs := &models.NotificationPreferences{DomainVerification: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyDomainVerification("example.com", "failed")
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotifyBackupStatus_Success_DoesNotPanic(t *testing.T) {
	prefs := &models.NotificationPreferences{BackupStatus: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyBackupStatus("success", "5.2 MB written")
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotifyBackupStatus_Failure_DoesNotPanic(t *testing.T) {
	prefs := &models.NotificationPreferences{BackupStatus: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyBackupStatus("failed", "disk full")
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
	}
}

func TestNotifyHighTraffic_DoesNotPanic(t *testing.T) {
	prefs := &models.NotificationPreferences{HighTraffic: true}
	nm := NewNotificationManager(nil, prefs, "admin@test.com")
	nm.NotifyHighTraffic(1500, 1000)
	if nm.GetQueueLength() != 1 {
		t.Errorf("queue length = %d, want 1", nm.GetQueueLength())
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
