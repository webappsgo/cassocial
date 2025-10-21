package models

import (
	"database/sql"
	"errors"
	"time"
)

// BlockedPattern represents a blocked URL/domain/word pattern
type BlockedPattern struct {
	ID          string    `json:"id" db:"id"`
	Pattern     string    `json:"pattern" db:"pattern"`
	PatternType string    `json:"pattern_type" db:"pattern_type"`
	Reason      string    `json:"reason,omitempty" db:"reason"`
	Severity    string    `json:"severity" db:"severity"`
	CreatedBy   string    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ReportedContent represents a content report
type ReportedContent struct {
	ID             string       `json:"id" db:"id"`
	ContentType    string       `json:"content_type" db:"content_type"`
	ContentID      string       `json:"content_id" db:"content_id"`
	ReporterIPHash string       `json:"reporter_ip_hash,omitempty" db:"reporter_ip_hash"`
	ReporterEmail  string       `json:"reporter_email,omitempty" db:"reporter_email"`
	Reason         string       `json:"reason" db:"reason"`
	Details        string       `json:"details,omitempty" db:"details"`
	Status         string       `json:"status" db:"status"`
	ModeratorID    string       `json:"moderator_id,omitempty" db:"moderator_id"`
	ModeratorNotes string       `json:"moderator_notes,omitempty" db:"moderator_notes"`
	ActionTaken    string       `json:"action_taken,omitempty" db:"action_taken"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	ResolvedAt     sql.NullTime `json:"resolved_at,omitempty" db:"resolved_at"`
}

// Valid pattern types
const (
	PatternTypeDomain = "domain"
	PatternTypeURL    = "url"
	PatternTypeWord   = "word"
)

// Valid severity levels
const (
	SeverityWarning = "warning"
	SeverityBlock   = "block"
)

// Valid content types for reports
const (
	ContentTypeProfile = "profile"
	ContentTypeLink    = "link"
)

// Valid report reasons
const (
	ReasonSpam        = "spam"
	ReasonInappropriate = "inappropriate"
	ReasonPhishing    = "phishing"
	ReasonCopyright   = "copyright"
	ReasonOther       = "other"
)

// Valid report statuses
const (
	ReportStatusPending   = "pending"
	ReportStatusReviewing = "reviewing"
	ReportStatusResolved  = "resolved"
	ReportStatusDismissed = "dismissed"
)

// Valid actions taken
const (
	ActionNone      = "none"
	ActionWarning   = "warning"
	ActionEdited    = "edited"
	ActionSuspended = "suspended"
	ActionDeleted   = "deleted"
)

var (
	ErrPatternEmpty       = errors.New("pattern cannot be empty")
	ErrInvalidPatternType = errors.New("invalid pattern type")
	ErrInvalidSeverity    = errors.New("invalid severity")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrInvalidReason      = errors.New("invalid report reason")
	ErrInvalidReportStatus = errors.New("invalid report status")
	ErrInvalidAction      = errors.New("invalid action taken")
)

// Validate validates the blocked pattern model
func (bp *BlockedPattern) Validate() error {
	if bp.Pattern == "" {
		return ErrPatternEmpty
	}

	validTypes := map[string]bool{
		PatternTypeDomain: true,
		PatternTypeURL:    true,
		PatternTypeWord:   true,
	}
	if !validTypes[bp.PatternType] {
		return ErrInvalidPatternType
	}

	validSeverities := map[string]bool{
		SeverityWarning: true,
		SeverityBlock:   true,
	}
	if !validSeverities[bp.Severity] {
		return ErrInvalidSeverity
	}

	return nil
}

// IsBlocking checks if the pattern blocks content
func (bp *BlockedPattern) IsBlocking() bool {
	return bp.Severity == SeverityBlock
}

// Validate validates the reported content model
func (rc *ReportedContent) Validate() error {
	validContentTypes := map[string]bool{
		ContentTypeProfile: true,
		ContentTypeLink:    true,
	}
	if !validContentTypes[rc.ContentType] {
		return ErrInvalidContentType
	}

	validReasons := map[string]bool{
		ReasonSpam:        true,
		ReasonInappropriate: true,
		ReasonPhishing:    true,
		ReasonCopyright:   true,
		ReasonOther:       true,
	}
	if !validReasons[rc.Reason] {
		return ErrInvalidReason
	}

	validStatuses := map[string]bool{
		StatusPending:         true,
		ReportStatusReviewing: true,
		ReportStatusResolved:  true,
		ReportStatusDismissed: true,
	}
	if !validStatuses[rc.Status] {
		return ErrInvalidReportStatus
	}

	if rc.ActionTaken != "" {
		validActions := map[string]bool{
			ActionNone:      true,
			ActionWarning:   true,
			ActionEdited:    true,
			ActionSuspended: true,
			ActionDeleted:   true,
		}
		if !validActions[rc.ActionTaken] {
			return ErrInvalidAction
		}
	}

	return nil
}

// IsPending checks if the report is pending review
func (rc *ReportedContent) IsPending() bool {
	return rc.Status == StatusPending
}

// IsResolved checks if the report is resolved
func (rc *ReportedContent) IsResolved() bool {
	return rc.Status == ReportStatusResolved || rc.Status == ReportStatusDismissed
}

// Resolve marks the report as resolved
func (rc *ReportedContent) Resolve(moderatorID, action, notes string) {
	rc.Status = ReportStatusResolved
	rc.ModeratorID = moderatorID
	rc.ActionTaken = action
	rc.ModeratorNotes = notes
	rc.ResolvedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}

// Dismiss marks the report as dismissed
func (rc *ReportedContent) Dismiss(moderatorID, notes string) {
	rc.Status = ReportStatusDismissed
	rc.ModeratorID = moderatorID
	rc.ActionTaken = ActionNone
	rc.ModeratorNotes = notes
	rc.ResolvedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}
