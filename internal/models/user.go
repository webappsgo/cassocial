package models

import (
	"database/sql"
	"errors"
	"regexp"
	"time"
)

// User represents a user account in the system
type User struct {
	ID                    string       `json:"id" db:"id"`
	Username              string       `json:"username" db:"username"`
	Email                 string       `json:"email" db:"email"`
	PasswordHash          string       `json:"-" db:"password_hash"`
	Role                  string       `json:"role" db:"role"`
	Status                string       `json:"status" db:"status"`
	CreatedAt             time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at" db:"updated_at"`
	LastLogin             sql.NullTime `json:"last_login,omitempty" db:"last_login"`
	EmailVerified         bool         `json:"email_verified" db:"email_verified"`
	TwoFactorEnabled      bool         `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret       string       `json:"-" db:"two_factor_secret"`
	PasswordResetToken    string       `json:"-" db:"password_reset_token"`
	PasswordResetExpires  sql.NullTime `json:"-" db:"password_reset_expires"`
}

// Valid user roles
const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleViewer = "viewer"
)

// Valid user statuses
const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusPending   = "pending"
)

var (
	ErrInvalidUsername = errors.New("username must be between 3 and 30 characters")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidRole     = errors.New("invalid role")
	ErrInvalidStatus   = errors.New("invalid status")
)

// Validate validates the user model
func (u *User) Validate() error {
	// Validate username
	if len(u.Username) < 3 || len(u.Username) > 30 {
		return ErrInvalidUsername
	}

	// Validate email
	if !isValidEmail(u.Email) {
		return ErrInvalidEmail
	}

	// Validate role
	if u.Role != RoleAdmin && u.Role != RoleUser && u.Role != RoleViewer {
		return ErrInvalidRole
	}

	// Validate status
	if u.Status != StatusActive && u.Status != StatusSuspended && u.Status != StatusPending {
		return ErrInvalidStatus
	}

	return nil
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsActive checks if the user status is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// CanLogin checks if the user can login (active and email verified)
func (u *User) CanLogin() bool {
	return u.IsActive() && u.EmailVerified
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	u.LastLogin = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}

// SanitizeForJSON returns a user struct safe for JSON responses (without sensitive fields)
func (u *User) SanitizeForJSON() *User {
	return &User{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		Role:             u.Role,
		Status:           u.Status,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
		LastLogin:        u.LastLogin,
		EmailVerified:    u.EmailVerified,
		TwoFactorEnabled: u.TwoFactorEnabled,
	}
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
