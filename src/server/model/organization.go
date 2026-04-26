package model

import (
	"database/sql"
	"errors"
	"time"
)

// Organization represents a team/organization
type Organization struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	OwnerID   string    `json:"owner_id" db:"owner_id"`
	LogoURL   string    `json:"logo_url,omitempty" db:"logo_url"`
	Settings  string    `json:"settings,omitempty" db:"settings"` // JSONB stored as string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	OrgID     string    `json:"org_id" db:"org_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	InvitedBy string    `json:"invited_by,omitempty" db:"invited_by"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}

// OrganizationInvite represents an invitation to join an organization
type OrganizationInvite struct {
	ID         string       `json:"id" db:"id"`
	OrgID      string       `json:"org_id" db:"org_id"`
	Email      string       `json:"email" db:"email"`
	Role       string       `json:"role" db:"role"`
	Token      string       `json:"token" db:"token"`
	InvitedBy  string       `json:"invited_by" db:"invited_by"`
	ExpiresAt  sql.NullTime `json:"expires_at,omitempty" db:"expires_at"`
	AcceptedAt sql.NullTime `json:"accepted_at,omitempty" db:"accepted_at"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
}

// Valid organization roles
const (
	OrgRoleOwner  = "owner"
	OrgRoleAdmin  = "admin"
	OrgRoleEditor = "editor"
	OrgRoleViewer = "viewer"
)

var (
	ErrOrgNameEmpty    = errors.New("organization name cannot be empty")
	ErrInvalidOrgRole  = errors.New("invalid organization role")
	ErrInviteExpired   = errors.New("invite has expired")
	ErrInviteAccepted  = errors.New("invite already accepted")
)

// Validate validates the organization model
func (o *Organization) Validate() error {
	if o.Name == "" {
		return ErrOrgNameEmpty
	}
	return nil
}

// Validate validates the organization member model
func (om *OrganizationMember) Validate() error {
	validRoles := map[string]bool{
		OrgRoleOwner:  true,
		OrgRoleAdmin:  true,
		OrgRoleEditor: true,
		OrgRoleViewer: true,
	}

	if !validRoles[om.Role] {
		return ErrInvalidOrgRole
	}

	return nil
}

// CanEdit checks if the member has edit permissions
func (om *OrganizationMember) CanEdit() bool {
	return om.Role == OrgRoleOwner || om.Role == OrgRoleAdmin || om.Role == OrgRoleEditor
}

// CanManageMembers checks if the member can manage other members
func (om *OrganizationMember) CanManageMembers() bool {
	return om.Role == OrgRoleOwner || om.Role == OrgRoleAdmin
}

// IsOwner checks if the member is the owner
func (om *OrganizationMember) IsOwner() bool {
	return om.Role == OrgRoleOwner
}

// Validate validates the organization invite model
func (oi *OrganizationInvite) Validate() error {
	validRoles := map[string]bool{
		OrgRoleAdmin:  true,
		OrgRoleEditor: true,
		OrgRoleViewer: true,
	}

	if !validRoles[oi.Role] {
		return ErrInvalidOrgRole
	}

	return nil
}

// IsExpired checks if the invite has expired
func (oi *OrganizationInvite) IsExpired() bool {
	if !oi.ExpiresAt.Valid {
		return false
	}
	return time.Now().After(oi.ExpiresAt.Time)
}

// IsAccepted checks if the invite has been accepted
func (oi *OrganizationInvite) IsAccepted() bool {
	return oi.AcceptedAt.Valid
}

// Accept marks the invite as accepted
func (oi *OrganizationInvite) Accept() error {
	if oi.IsExpired() {
		return ErrInviteExpired
	}
	if oi.IsAccepted() {
		return ErrInviteAccepted
	}
	oi.AcceptedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	return nil
}
