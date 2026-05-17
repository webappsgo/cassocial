package model

import (
	"database/sql"
	"testing"
	"time"
)

func TestUserConsent_Validate_Valid(t *testing.T) {
	uc := &UserConsent{TermsVersion: "1.0", PrivacyVersion: "1.0"}
	if err := uc.Validate(); err != nil {
		t.Errorf("valid consent Validate() = %v, want nil", err)
	}
}

func TestUserConsent_Validate_MissingTerms(t *testing.T) {
	uc := &UserConsent{PrivacyVersion: "1.0"}
	if err := uc.Validate(); err != ErrConsentNotGiven {
		t.Errorf("missing terms Validate() = %v, want ErrConsentNotGiven", err)
	}
}

func TestUserConsent_Validate_MissingPrivacy(t *testing.T) {
	uc := &UserConsent{TermsVersion: "1.0"}
	if err := uc.Validate(); err != ErrConsentNotGiven {
		t.Errorf("missing privacy Validate() = %v, want ErrConsentNotGiven", err)
	}
}

func TestUserConsent_HasGivenConsent(t *testing.T) {
	uc := &UserConsent{TermsVersion: "1.0", PrivacyVersion: "1.0"}
	if !uc.HasGivenConsent() {
		t.Error("HasGivenConsent() should return true when both versions set")
	}
	uc.TermsVersion = ""
	if uc.HasGivenConsent() {
		t.Error("HasGivenConsent() should return false when terms missing")
	}
}

func TestUserConsent_RequestDeletion(t *testing.T) {
	uc := &UserConsent{}
	uc.RequestDeletion(30)
	if !uc.DeletionRequestedAt.Valid {
		t.Error("RequestDeletion() should set DeletionRequestedAt")
	}
	if !uc.DeletionScheduledFor.Valid {
		t.Error("RequestDeletion() should set DeletionScheduledFor")
	}
	// Should be approximately 30 days from now
	expectedTime := time.Now().AddDate(0, 0, 30)
	diff := uc.DeletionScheduledFor.Time.Sub(expectedTime)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("DeletionScheduledFor off by %v, want ~30 days", diff)
	}
}

func TestDataExport_Validate_ValidStatuses(t *testing.T) {
	for _, status := range []string{ExportStatusPending, ExportStatusProcessing, ExportStatusCompleted, ExportStatusExpired} {
		de := &DataExport{Status: status}
		if err := de.Validate(); err != nil {
			t.Errorf("status %q Validate() = %v, want nil", status, err)
		}
	}
}

func TestDataExport_Validate_InvalidStatus(t *testing.T) {
	de := &DataExport{Status: "unknown"}
	if err := de.Validate(); err != ErrInvalidExportStatus {
		t.Errorf("invalid status Validate() = %v, want ErrInvalidExportStatus", err)
	}
}

func TestDataExport_IsExpired_NoExpiry(t *testing.T) {
	de := &DataExport{}
	if de.IsExpired() {
		t.Error("export with no expiry should not be expired")
	}
}

func TestDataExport_IsExpired_Future(t *testing.T) {
	de := &DataExport{ExpiresAt: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}}
	if de.IsExpired() {
		t.Error("future expiry export should not be expired")
	}
}

func TestDataExport_IsExpired_Past(t *testing.T) {
	de := &DataExport{ExpiresAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true}}
	if !de.IsExpired() {
		t.Error("past expiry export should be expired")
	}
}

func TestDataExport_IsCompleted(t *testing.T) {
	de := &DataExport{Status: ExportStatusPending}
	if de.IsCompleted() {
		t.Error("pending export should not be completed")
	}
	de.Status = ExportStatusCompleted
	if !de.IsCompleted() {
		t.Error("completed export should return IsCompleted=true")
	}
}

func TestDataExport_Complete(t *testing.T) {
	de := &DataExport{Status: ExportStatusPending}
	de.Complete("/exports/data.zip", 7)
	if de.Status != ExportStatusCompleted {
		t.Errorf("Complete() status = %q, want completed", de.Status)
	}
	if de.FilePath != "/exports/data.zip" {
		t.Errorf("Complete() FilePath = %q, want /exports/data.zip", de.FilePath)
	}
	if !de.CompletedAt.Valid {
		t.Error("Complete() should set CompletedAt")
	}
	if !de.ExpiresAt.Valid {
		t.Error("Complete() should set ExpiresAt")
	}
}

func TestImportJob_Validate_Valid(t *testing.T) {
	ij := &ImportJob{Source: ImportSourceLinktree, Status: ImportStatusPending}
	if err := ij.Validate(); err != nil {
		t.Errorf("valid import Validate() = %v, want nil", err)
	}
}

func TestImportJob_Validate_InvalidSource(t *testing.T) {
	ij := &ImportJob{Source: "unknown", Status: ImportStatusPending}
	if err := ij.Validate(); err != ErrInvalidImportSource {
		t.Errorf("invalid source Validate() = %v, want ErrInvalidImportSource", err)
	}
}

func TestImportJob_Validate_InvalidStatus(t *testing.T) {
	ij := &ImportJob{Source: ImportSourceCSV, Status: "queued"}
	if err := ij.Validate(); err != ErrInvalidImportStatus {
		t.Errorf("invalid status Validate() = %v, want ErrInvalidImportStatus", err)
	}
}

func TestImportJob_Validate_AllSources(t *testing.T) {
	sources := []string{ImportSourceLinktree, ImportSourceLinkstack, ImportSourceCarrd, ImportSourceAboutMe, ImportSourceCSV, ImportSourceJSON}
	for _, src := range sources {
		ij := &ImportJob{Source: src, Status: ImportStatusPending}
		if err := ij.Validate(); err != nil {
			t.Errorf("source %q Validate() = %v, want nil", src, err)
		}
	}
}

func TestImportJob_Complete(t *testing.T) {
	ij := &ImportJob{Source: ImportSourceCSV, Status: ImportStatusProcessing}
	ij.Complete("imported 10 links")
	if ij.Status != ImportStatusCompleted {
		t.Errorf("Complete() status = %q, want completed", ij.Status)
	}
	if ij.Result != "imported 10 links" {
		t.Errorf("Complete() result = %q, want 'imported 10 links'", ij.Result)
	}
	if !ij.CompletedAt.Valid {
		t.Error("Complete() should set CompletedAt")
	}
}

func TestImportJob_Fail(t *testing.T) {
	ij := &ImportJob{Source: ImportSourceJSON, Status: ImportStatusProcessing}
	ij.Fail("parse error")
	if ij.Status != ImportStatusFailed {
		t.Errorf("Fail() status = %q, want failed", ij.Status)
	}
	if ij.Result != "parse error" {
		t.Errorf("Fail() result = %q, want 'parse error'", ij.Result)
	}
}

func TestNewAuditLog(t *testing.T) {
	al := NewAuditLog("user1", AuditActionLogin, "user", "user1", "127.0.0.1", "curl/7.0")
	if al == nil {
		t.Fatal("NewAuditLog() returned nil")
	}
	if al.UserID != "user1" {
		t.Errorf("UserID = %q, want user1", al.UserID)
	}
	if al.Action != AuditActionLogin {
		t.Errorf("Action = %q, want %q", al.Action, AuditActionLogin)
	}
	if al.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}
