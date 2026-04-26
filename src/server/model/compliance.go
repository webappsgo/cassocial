package model

import (
	"database/sql"
	"errors"
	"time"
)

// UserConsent represents user consent for terms, privacy, etc.
type UserConsent struct {
	UserID                string       `json:"user_id" db:"user_id"`
	TermsVersion          string       `json:"terms_version" db:"terms_version"`
	TermsAcceptedAt       time.Time    `json:"terms_accepted_at" db:"terms_accepted_at"`
	PrivacyVersion        string       `json:"privacy_version" db:"privacy_version"`
	PrivacyAcceptedAt     time.Time    `json:"privacy_accepted_at" db:"privacy_accepted_at"`
	CookiesAccepted       bool         `json:"cookies_accepted" db:"cookies_accepted"`
	CookiesAcceptedAt     sql.NullTime `json:"cookies_accepted_at,omitempty" db:"cookies_accepted_at"`
	MarketingConsent      bool         `json:"marketing_consent" db:"marketing_consent"`
	MarketingConsentAt    sql.NullTime `json:"marketing_consent_at,omitempty" db:"marketing_consent_at"`
	DataExportRequestedAt sql.NullTime `json:"data_export_requested_at,omitempty" db:"data_export_requested_at"`
	DeletionRequestedAt   sql.NullTime `json:"deletion_requested_at,omitempty" db:"deletion_requested_at"`
	DeletionScheduledFor  sql.NullTime `json:"deletion_scheduled_for,omitempty" db:"deletion_scheduled_for"`
}

// DataExport represents a data export request
type DataExport struct {
	ID          string       `json:"id" db:"id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Status      string       `json:"status" db:"status"`
	FilePath    string       `json:"file_path,omitempty" db:"file_path"`
	ExpiresAt   sql.NullTime `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	CompletedAt sql.NullTime `json:"completed_at,omitempty" db:"completed_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id,omitempty" db:"user_id"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type,omitempty" db:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty" db:"resource_id"`
	IPAddress    string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string    `json:"user_agent,omitempty" db:"user_agent"`
	Metadata     string    `json:"metadata,omitempty" db:"metadata"` // JSONB stored as string
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ImportJob represents an import job
type ImportJob struct {
	ID          string       `json:"id" db:"id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Source      string       `json:"source" db:"source"`
	Status      string       `json:"status" db:"status"`
	FilePath    string       `json:"file_path,omitempty" db:"file_path"`
	Result      string       `json:"result,omitempty" db:"result"` // JSONB stored as string
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	CompletedAt sql.NullTime `json:"completed_at,omitempty" db:"completed_at"`
}

// Valid data export statuses
const (
	ExportStatusPending    = "pending"
	ExportStatusProcessing = "processing"
	ExportStatusCompleted  = "completed"
	ExportStatusExpired    = "expired"
)

// Valid import sources
const (
	ImportSourceLinktree   = "linktree"
	ImportSourceLinkstack  = "linkstack"
	ImportSourceCarrd      = "carrd"
	ImportSourceAboutMe    = "aboutme"
	ImportSourceCSV        = "csv"
	ImportSourceJSON       = "json"
)

// Valid import statuses
const (
	ImportStatusPending    = "pending"
	ImportStatusProcessing = "processing"
	ImportStatusCompleted  = "completed"
	ImportStatusFailed     = "failed"
)

// Common audit actions
const (
	AuditActionLogin             = "user.login"
	AuditActionLogout            = "user.logout"
	AuditActionPasswordChange    = "user.password_change"
	AuditActionProfileCreate     = "profile.create"
	AuditActionProfileUpdate     = "profile.update"
	AuditActionProfileDelete     = "profile.delete"
	AuditActionLinkCreate        = "link.create"
	AuditActionLinkUpdate        = "link.update"
	AuditActionLinkDelete        = "link.delete"
	AuditActionSettingsUpdate    = "settings.update"
	AuditActionAPIKeyCreate      = "api_key.create"
	AuditActionAPIKeyRevoke      = "api_key.revoke"
)

var (
	ErrConsentNotGiven     = errors.New("required consent not given")
	ErrExportExpired       = errors.New("data export has expired")
	ErrExportNotCompleted  = errors.New("data export not completed")
	ErrInvalidExportStatus = errors.New("invalid export status")
	ErrInvalidImportSource = errors.New("invalid import source")
	ErrInvalidImportStatus = errors.New("invalid import status")
)

// Validate validates the user consent model
func (uc *UserConsent) Validate() error {
	if uc.TermsVersion == "" || uc.PrivacyVersion == "" {
		return ErrConsentNotGiven
	}
	return nil
}

// HasGivenConsent checks if all required consents are given
func (uc *UserConsent) HasGivenConsent() bool {
	return uc.TermsVersion != "" && uc.PrivacyVersion != ""
}

// RequestDeletion schedules account deletion
func (uc *UserConsent) RequestDeletion(daysUntilDeletion int) {
	now := time.Now()
	uc.DeletionRequestedAt = sql.NullTime{Time: now, Valid: true}
	uc.DeletionScheduledFor = sql.NullTime{
		Time:  now.AddDate(0, 0, daysUntilDeletion),
		Valid: true,
	}
}

// Validate validates the data export model
func (de *DataExport) Validate() error {
	validStatuses := map[string]bool{
		ExportStatusPending:    true,
		ExportStatusProcessing: true,
		ExportStatusCompleted:  true,
		ExportStatusExpired:    true,
	}

	if !validStatuses[de.Status] {
		return ErrInvalidExportStatus
	}

	return nil
}

// IsExpired checks if the export has expired
func (de *DataExport) IsExpired() bool {
	if !de.ExpiresAt.Valid {
		return false
	}
	return time.Now().After(de.ExpiresAt.Time)
}

// IsCompleted checks if the export is completed
func (de *DataExport) IsCompleted() bool {
	return de.Status == ExportStatusCompleted
}

// Complete marks the export as completed
func (de *DataExport) Complete(filePath string, expirationDays int) {
	de.Status = ExportStatusCompleted
	de.FilePath = filePath
	de.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	de.ExpiresAt = sql.NullTime{
		Time:  time.Now().AddDate(0, 0, expirationDays),
		Valid: true,
	}
}

// Validate validates the import job model
func (ij *ImportJob) Validate() error {
	validSources := map[string]bool{
		ImportSourceLinktree:  true,
		ImportSourceLinkstack: true,
		ImportSourceCarrd:     true,
		ImportSourceAboutMe:   true,
		ImportSourceCSV:       true,
		ImportSourceJSON:      true,
	}

	if !validSources[ij.Source] {
		return ErrInvalidImportSource
	}

	validStatuses := map[string]bool{
		ImportStatusPending:    true,
		ImportStatusProcessing: true,
		ImportStatusCompleted:  true,
		ImportStatusFailed:     true,
	}

	if !validStatuses[ij.Status] {
		return ErrInvalidImportStatus
	}

	return nil
}

// Complete marks the import as completed
func (ij *ImportJob) Complete(result string) {
	ij.Status = ImportStatusCompleted
	ij.Result = result
	ij.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

// Fail marks the import as failed
func (ij *ImportJob) Fail(errorMsg string) {
	ij.Status = ImportStatusFailed
	ij.Result = errorMsg
	ij.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(userID, action, resourceType, resourceID, ipAddress, userAgent string) *AuditLog {
	return &AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
	}
}
