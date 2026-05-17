package model

import (
	"testing"
)

func TestBlockedPattern_Validate_Valid(t *testing.T) {
	bp := &BlockedPattern{
		Pattern:     "spam.com",
		PatternType: PatternTypeDomain,
		Severity:    SeverityBlock,
	}
	if err := bp.Validate(); err != nil {
		t.Errorf("valid pattern Validate() = %v, want nil", err)
	}
}

func TestBlockedPattern_Validate_EmptyPattern(t *testing.T) {
	bp := &BlockedPattern{PatternType: PatternTypeDomain, Severity: SeverityBlock}
	if err := bp.Validate(); err != ErrPatternEmpty {
		t.Errorf("empty pattern Validate() = %v, want ErrPatternEmpty", err)
	}
}

func TestBlockedPattern_Validate_InvalidType(t *testing.T) {
	bp := &BlockedPattern{Pattern: "x", PatternType: "regex", Severity: SeverityBlock}
	if err := bp.Validate(); err != ErrInvalidPatternType {
		t.Errorf("invalid type Validate() = %v, want ErrInvalidPatternType", err)
	}
}

func TestBlockedPattern_Validate_InvalidSeverity(t *testing.T) {
	bp := &BlockedPattern{Pattern: "x", PatternType: PatternTypeURL, Severity: "critical"}
	if err := bp.Validate(); err != ErrInvalidSeverity {
		t.Errorf("invalid severity Validate() = %v, want ErrInvalidSeverity", err)
	}
}

func TestBlockedPattern_Validate_AllTypes(t *testing.T) {
	for _, pt := range []string{PatternTypeDomain, PatternTypeURL, PatternTypeWord} {
		bp := &BlockedPattern{Pattern: "test", PatternType: pt, Severity: SeverityWarning}
		if err := bp.Validate(); err != nil {
			t.Errorf("type %q Validate() = %v, want nil", pt, err)
		}
	}
}

func TestBlockedPattern_IsBlocking(t *testing.T) {
	bp := &BlockedPattern{Severity: SeverityBlock}
	if !bp.IsBlocking() {
		t.Error("IsBlocking() should return true for SeverityBlock")
	}
	bp.Severity = SeverityWarning
	if bp.IsBlocking() {
		t.Error("IsBlocking() should return false for SeverityWarning")
	}
}

func validReportedContent() *ReportedContent {
	return &ReportedContent{
		ContentType: ContentTypeProfile,
		Reason:      ReasonSpam,
		Status:      ReportStatusPending,
	}
}

func TestReportedContent_Validate_Valid(t *testing.T) {
	rc := validReportedContent()
	if err := rc.Validate(); err != nil {
		t.Errorf("valid report Validate() = %v, want nil", err)
	}
}

func TestReportedContent_Validate_InvalidContentType(t *testing.T) {
	rc := validReportedContent()
	rc.ContentType = "comment"
	if err := rc.Validate(); err != ErrInvalidContentType {
		t.Errorf("invalid content type Validate() = %v, want ErrInvalidContentType", err)
	}
}

func TestReportedContent_Validate_InvalidReason(t *testing.T) {
	rc := validReportedContent()
	rc.Reason = "dislike"
	if err := rc.Validate(); err != ErrInvalidReason {
		t.Errorf("invalid reason Validate() = %v, want ErrInvalidReason", err)
	}
}

func TestReportedContent_Validate_InvalidStatus(t *testing.T) {
	rc := validReportedContent()
	rc.Status = "unknown"
	if err := rc.Validate(); err != ErrInvalidReportStatus {
		t.Errorf("invalid status Validate() = %v, want ErrInvalidReportStatus", err)
	}
}

func TestReportedContent_Validate_InvalidAction(t *testing.T) {
	rc := validReportedContent()
	rc.ActionTaken = "banned"
	if err := rc.Validate(); err != ErrInvalidAction {
		t.Errorf("invalid action Validate() = %v, want ErrInvalidAction", err)
	}
}

func TestReportedContent_Validate_ValidAction(t *testing.T) {
	rc := validReportedContent()
	rc.ActionTaken = ActionDeleted
	if err := rc.Validate(); err != nil {
		t.Errorf("valid action Validate() = %v, want nil", err)
	}
}

func TestReportedContent_IsPending(t *testing.T) {
	rc := &ReportedContent{Status: ReportStatusPending}
	if !rc.IsPending() {
		t.Error("IsPending() should return true for pending status")
	}
	rc.Status = ReportStatusResolved
	if rc.IsPending() {
		t.Error("IsPending() should return false for resolved status")
	}
}

func TestReportedContent_IsResolved(t *testing.T) {
	rc := &ReportedContent{Status: ReportStatusResolved}
	if !rc.IsResolved() {
		t.Error("IsResolved() should return true for resolved status")
	}
	rc.Status = ReportStatusDismissed
	if !rc.IsResolved() {
		t.Error("IsResolved() should return true for dismissed status")
	}
	rc.Status = ReportStatusPending
	if rc.IsResolved() {
		t.Error("IsResolved() should return false for pending status")
	}
}

func TestReportedContent_Resolve(t *testing.T) {
	rc := validReportedContent()
	rc.Resolve("mod1", ActionWarning, "First offense")
	if rc.Status != ReportStatusResolved {
		t.Errorf("Resolve() status = %q, want resolved", rc.Status)
	}
	if rc.ModeratorID != "mod1" {
		t.Errorf("Resolve() ModeratorID = %q, want mod1", rc.ModeratorID)
	}
	if !rc.ResolvedAt.Valid {
		t.Error("Resolve() should set ResolvedAt")
	}
}

func TestReportedContent_Dismiss(t *testing.T) {
	rc := validReportedContent()
	rc.Dismiss("mod2", "Not a violation")
	if rc.Status != ReportStatusDismissed {
		t.Errorf("Dismiss() status = %q, want dismissed", rc.Status)
	}
	if rc.ActionTaken != ActionNone {
		t.Errorf("Dismiss() ActionTaken = %q, want none", rc.ActionTaken)
	}
	if !rc.ResolvedAt.Valid {
		t.Error("Dismiss() should set ResolvedAt")
	}
}

func TestReportedContent_Validate_AllReasons(t *testing.T) {
	for _, reason := range []string{ReasonSpam, ReasonInappropriate, ReasonPhishing, ReasonCopyright, ReasonOther} {
		rc := &ReportedContent{ContentType: ContentTypeLink, Reason: reason, Status: ReportStatusReviewing}
		if err := rc.Validate(); err != nil {
			t.Errorf("reason %q Validate() = %v, want nil", reason, err)
		}
	}
}
