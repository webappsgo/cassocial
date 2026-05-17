package model

import (
	"database/sql"
	"testing"
	"time"
)

func TestOrganization_Validate_Valid(t *testing.T) {
	o := &Organization{Name: "MyOrg"}
	if err := o.Validate(); err != nil {
		t.Errorf("valid org Validate() = %v, want nil", err)
	}
}

func TestOrganization_Validate_EmptyName(t *testing.T) {
	o := &Organization{}
	if err := o.Validate(); err != ErrOrgNameEmpty {
		t.Errorf("empty name Validate() = %v, want ErrOrgNameEmpty", err)
	}
}

func TestOrganizationMember_Validate_ValidRoles(t *testing.T) {
	for _, role := range []string{OrgRoleOwner, OrgRoleAdmin, OrgRoleEditor, OrgRoleViewer} {
		om := &OrganizationMember{Role: role}
		if err := om.Validate(); err != nil {
			t.Errorf("role %q Validate() = %v, want nil", role, err)
		}
	}
}

func TestOrganizationMember_Validate_InvalidRole(t *testing.T) {
	om := &OrganizationMember{Role: "superuser"}
	if err := om.Validate(); err != ErrInvalidOrgRole {
		t.Errorf("invalid role Validate() = %v, want ErrInvalidOrgRole", err)
	}
}

func TestOrganizationMember_CanEdit(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{OrgRoleOwner, true},
		{OrgRoleAdmin, true},
		{OrgRoleEditor, true},
		{OrgRoleViewer, false},
	}
	for _, tt := range tests {
		om := &OrganizationMember{Role: tt.role}
		if got := om.CanEdit(); got != tt.want {
			t.Errorf("role %q CanEdit() = %v, want %v", tt.role, got, tt.want)
		}
	}
}

func TestOrganizationMember_CanManageMembers(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{OrgRoleOwner, true},
		{OrgRoleAdmin, true},
		{OrgRoleEditor, false},
		{OrgRoleViewer, false},
	}
	for _, tt := range tests {
		om := &OrganizationMember{Role: tt.role}
		if got := om.CanManageMembers(); got != tt.want {
			t.Errorf("role %q CanManageMembers() = %v, want %v", tt.role, got, tt.want)
		}
	}
}

func TestOrganizationMember_IsOwner(t *testing.T) {
	om := &OrganizationMember{Role: OrgRoleOwner}
	if !om.IsOwner() {
		t.Error("IsOwner() should return true for owner")
	}
	om.Role = OrgRoleAdmin
	if om.IsOwner() {
		t.Error("IsOwner() should return false for admin")
	}
}

func TestOrganizationInvite_Validate_ValidRoles(t *testing.T) {
	for _, role := range []string{OrgRoleAdmin, OrgRoleEditor, OrgRoleViewer} {
		oi := &OrganizationInvite{Role: role}
		if err := oi.Validate(); err != nil {
			t.Errorf("role %q Validate() = %v, want nil", role, err)
		}
	}
}

func TestOrganizationInvite_Validate_OwnerNotAllowed(t *testing.T) {
	oi := &OrganizationInvite{Role: OrgRoleOwner}
	if err := oi.Validate(); err != ErrInvalidOrgRole {
		t.Errorf("owner invite Validate() = %v, want ErrInvalidOrgRole", err)
	}
}

func TestOrganizationInvite_IsExpired_NoExpiry(t *testing.T) {
	oi := &OrganizationInvite{}
	if oi.IsExpired() {
		t.Error("invite with no expiry should not be expired")
	}
}

func TestOrganizationInvite_IsExpired_FutureExpiry(t *testing.T) {
	oi := &OrganizationInvite{
		ExpiresAt: sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
	}
	if oi.IsExpired() {
		t.Error("future expiry invite should not be expired")
	}
}

func TestOrganizationInvite_IsExpired_PastExpiry(t *testing.T) {
	oi := &OrganizationInvite{
		ExpiresAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
	}
	if !oi.IsExpired() {
		t.Error("past expiry invite should be expired")
	}
}

func TestOrganizationInvite_IsAccepted(t *testing.T) {
	oi := &OrganizationInvite{}
	if oi.IsAccepted() {
		t.Error("new invite should not be accepted")
	}
	oi.AcceptedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if !oi.IsAccepted() {
		t.Error("invite with accepted_at should be accepted")
	}
}

func TestOrganizationInvite_Accept_Success(t *testing.T) {
	oi := &OrganizationInvite{
		Role:      OrgRoleEditor,
		ExpiresAt: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true},
	}
	if err := oi.Accept(); err != nil {
		t.Errorf("Accept() = %v, want nil", err)
	}
	if !oi.IsAccepted() {
		t.Error("Accept() should mark invite as accepted")
	}
}

func TestOrganizationInvite_Accept_Expired(t *testing.T) {
	oi := &OrganizationInvite{
		ExpiresAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
	}
	if err := oi.Accept(); err != ErrInviteExpired {
		t.Errorf("expired Accept() = %v, want ErrInviteExpired", err)
	}
}

func TestOrganizationInvite_Accept_AlreadyAccepted(t *testing.T) {
	oi := &OrganizationInvite{
		AcceptedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}
	if err := oi.Accept(); err != ErrInviteAccepted {
		t.Errorf("double accept Accept() = %v, want ErrInviteAccepted", err)
	}
}
